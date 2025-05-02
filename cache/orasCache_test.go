package cache

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	testregistry "github.com/google/go-containerregistry/pkg/registry"
	"oras.land/oras-go/v2/registry/remote"
)

func TestStore(t *testing.T) {
	cacheInstance := NewOras(t.Context(), setupInMemoryRepository(t))
	key := "test-key"
	expected := "my test value"
	err := cacheInstance.Store(key, strings.NewReader(expected), CacheOptions{
		ArtifactType: "application/vnd.bazel.cache.http",
	})
	if err != nil {
		t.Fatalf("failed to store artifact in cache: %s", err.Error())
	}
}

func TestRestore(t *testing.T) {
	cacheInstance := NewOras(t.Context(), setupInMemoryRepository(t))
	key := "test-key"
	expected := "my test value"
	err := cacheInstance.Store(key, strings.NewReader(expected), CacheOptions{
		ArtifactType: "application/vnd.bazel.cache.http",
	})
	if err != nil {
		t.Fatalf("failed to store artifact in cache: %s", err.Error())
	}

	var buffer bytes.Buffer
	cacheInstance.Restore(key, &buffer)
	if buffer.String() != expected {
		t.Fatalf("expected %s but got %s", expected, buffer.String())
	}
}

func setupInMemoryRepository(t *testing.T) *remote.Repository {
	registryServer := httptest.NewServer(
		testregistry.New(
			testregistry.WithBlobHandler(testregistry.NewInMemoryBlobHandler()),
			testregistry.Logger(log.New(io.Discard, "", 0))))
	t.Cleanup(registryServer.Close)
	repositoryUri := fmt.Sprintf("%s/cache", strings.TrimPrefix(registryServer.URL, "http://"))
	repo, err := remote.NewRepository(repositoryUri)
	if err != nil {
		t.Fatalf("failed to create registry: %s", err.Error())
	}
	repo.PlainHTTP = true
	return repo
}
