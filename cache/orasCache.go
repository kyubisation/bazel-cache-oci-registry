package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type OrasCache struct {
	registryUri string
	repository  string
	credentials *auth.Credential
	context     context.Context
}

const lastUsedAnnotation = "bazel-cache-oci-registry.last-used"

func NewOras(
	context context.Context,
	registryUri,
	repository string,
	credentials *auth.Credential) OrasCache {
	return OrasCache{registryUri, repository, credentials, context}
}

func (c OrasCache) Exists(key string) bool {
	if len(key) == 0 {
		return false
	}

	repository, err := c.prepareRepositoy()
	if err != nil {
		return false
	}

	_, r, err := repository.Manifests().FetchReference(c.context, key)
	if err != nil {
		return false
	}
	defer r.Close()

	var packManifest v1.Manifest
	err = json.NewDecoder(r).Decode(&packManifest)
	return err == nil
}

func (c OrasCache) Store(key string, reader io.Reader, options CacheOptions) error {
	if len(key) == 0 {
		return fmt.Errorf("key must not be empty")
	}

	repository, err := c.prepareRepositoy()
	if err != nil {
		return err
	}

	// We first create the oras configuration locally in memory to atomically
	// copy the whole state in one copy operation below.
	m := memory.New()
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, reader)
	if err != nil {
		return err
	}

	fileDescriptor, err := oras.PushBytes(c.context, m, "", buffer.Bytes())
	if err != nil {
		return err
	}

	opts := createPackManifestOptions(options, fileDescriptor)
	manifestDescriptor, err := oras.PackManifest(
		c.context, m, oras.PackManifestVersion1_1, options.ArtifactType, opts)
	if err != nil {
		return err
	}

	err = m.Tag(c.context, manifestDescriptor, key)
	if err != nil {
		return err
	}

	_, err = oras.Copy(c.context, m, key, repository, key, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	return nil
}

func (c OrasCache) Restore(key string, writer io.Writer) error {
	if len(key) == 0 {
		return fmt.Errorf("key must not be empty")
	}

	repository, err := c.prepareRepositoy()
	if err != nil {
		return err
	}

	_, r, err := repository.Manifests().FetchReference(c.context, key)
	if err != nil {
		return err
	}
	defer r.Close()

	var packManifest v1.Manifest
	err = json.NewDecoder(r).Decode(&packManifest)
	if err != nil {
		return err
	}

	err = c.updateLastUsed(repository, packManifest, key)
	if err != nil {
		return err
	}

	content, err := repository.Fetch(c.context, packManifest.Layers[0])
	if err != nil {
		return err
	}
	defer content.Close()

	_, err = io.Copy(writer, content)
	if err != nil {
		return err
	}

	return nil
}

func (c OrasCache) prepareRepositoy() (registry.Repository, error) {
	registry, err := remote.NewRegistry(c.registryUri)
	if err != nil {
		return nil, err
	} else if strings.HasPrefix(c.registryUri, "127.0.0.1") || strings.HasPrefix(c.registryUri, "localhost") {
		registry.PlainHTTP = true
	}
	if c.credentials != nil {
		registry.Client = &auth.Client{
			Credential: auth.StaticCredential(c.registryUri, *c.credentials),
		}
	}

	return registry.Repository(c.context, c.repository)
}

func createPackManifestOptions(
	options CacheOptions, fileDescriptor v1.Descriptor) oras.PackManifestOptions {
	annotations := make(map[string]string)
	if len(options.Annotations) > 0 {
		for key, value := range options.Annotations {
			annotations[key] = value
		}
	}

	return oras.PackManifestOptions{
		Layers:              []v1.Descriptor{fileDescriptor},
		ManifestAnnotations: annotations,
	}
}

func (c OrasCache) updateLastUsed(
	repository registry.Repository, manifest v1.Manifest, key string) error {
	if durationString, ok :=
		manifest.Annotations[lastUsedAnnotation]; ok {
		duration, err := time.ParseDuration(durationString)
		if err == nil {
			manifest.Annotations[lastUsedAnnotation] =
				time.Now().Add(duration).Format("2006-01-02T15:04:05.000Z")
		}
	}

	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	desc := content.NewDescriptorFromBytes(manifest.MediaType, manifestBytes)
	err = repository.Manifests().PushReference(c.context, desc, bytes.NewReader(manifestBytes), key)
	if err != nil {
		return err
	}

	return nil
}
