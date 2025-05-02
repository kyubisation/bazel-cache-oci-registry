package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/memory"
	"oras.land/oras-go/v2/registry/remote"
)

type OrasCache struct {
	context    context.Context
	repository *remote.Repository
}

func NewOras(
	context context.Context,
	repository *remote.Repository) OrasCache {
	return OrasCache{context, repository}
}

func (c OrasCache) Exists(key string) bool {
	if len(key) == 0 {
		return false
	}

	_, r, err := c.repository.Manifests().FetchReference(c.context, key)
	if err != nil {
		return false
	}
	defer r.Close()

	var packManifest v1.Manifest
	err = json.NewDecoder(r).Decode(&packManifest)
	return err == nil
}

func (c OrasCache) Store(key string, reader io.Reader) error {
	if len(key) == 0 {
		return fmt.Errorf("key must not be empty")
	}

	// We first create the oras configuration locally in memory to atomically
	// copy the whole state in one copy operation below.
	m := memory.New()
	var buffer bytes.Buffer
	_, err := io.Copy(&buffer, reader)
	if err != nil {
		return err
	}

	fileDescriptor, err := oras.PushBytes(c.context, m, "", buffer.Bytes())
	if err != nil {
		return err
	}

	opts := oras.PackManifestOptions{
		Layers:              []v1.Descriptor{fileDescriptor},
		ManifestAnnotations: make(map[string]string),
	}
	manifestDescriptor, err := oras.PackManifest(
		c.context, m, oras.PackManifestVersion1_1, "application/vnd.bazel.cache.http", opts)
	if err != nil {
		return err
	}

	err = m.Tag(c.context, manifestDescriptor, key)
	if err != nil {
		return err
	}

	_, err = oras.Copy(c.context, m, key, c.repository, key, oras.DefaultCopyOptions)
	if err != nil {
		return err
	}

	return nil
}

func (c OrasCache) Restore(key string, writer io.Writer) error {
	if len(key) == 0 {
		return fmt.Errorf("key must not be empty")
	}

	_, r, err := c.repository.FetchReference(c.context, key)
	if err != nil {
		return err
	}
	defer r.Close()

	var packManifest v1.Manifest
	err = json.NewDecoder(r).Decode(&packManifest)
	if err != nil {
		return err
	}

	content, err := c.repository.Fetch(c.context, packManifest.Layers[0])
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
