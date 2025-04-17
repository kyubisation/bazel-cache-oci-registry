package cache

import (
	"fmt"
	"io"
)

type NoopCache struct{}

func Noop() NoopCache {
	return NoopCache{}
}

func (c NoopCache) Exists(key string) (bool, map[string]string) {
	return false, nil
}

func (c NoopCache) Store(key string, reader io.Reader, options CacheOptions) error {
	if len(key) == 0 {
		return fmt.Errorf("key must not be empty")
	}
	return nil
}

func (c NoopCache) Restore(key string, writer io.Writer) error {
	if len(key) == 0 {
		return fmt.Errorf("key must not be empty")
	}
	return fmt.Errorf("not supported with noop cache")
}
