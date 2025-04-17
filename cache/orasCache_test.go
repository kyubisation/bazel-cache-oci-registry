package cache

import (
	"bytes"
	"io"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-containerregistry/pkg/registry"
)

func TestStore(t *testing.T) {
	cacheInstance := NewOras(t.Context(), setupInMemoryRegistry(t), "cache", nil)
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
	cacheInstance := NewOras(t.Context(), setupInMemoryRegistry(t), "cache", nil)
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

func setupInMemoryRegistry(t *testing.T) string {
	registry := httptest.NewServer(
		registry.New(
			registry.WithBlobHandler(registry.NewInMemoryBlobHandler()),
			registry.Logger(log.New(io.Discard, "", 0))))
	t.Cleanup(registry.Close)
	return strings.TrimPrefix(registry.URL, "http://")
}
