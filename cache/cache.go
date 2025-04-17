package cache

import (
	"io"
)

type CacheOptions struct {
	ArtifactType string
	Annotations  map[string]string
}

type Cache interface {
	Exists(key string) bool
	Store(key string, reader io.Reader, options CacheOptions) error
	Restore(key string, writer io.Writer) (err error)
}
