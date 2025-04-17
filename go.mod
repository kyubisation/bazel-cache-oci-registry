module bazel-cache-oci-registry

go 1.24.1

require github.com/spf13/cobra v1.8.1

require (
	github.com/opencontainers/image-spec v1.1.0
	oras.land/oras-go/v2 v2.5.0
)

require github.com/opencontainers/go-digest v1.0.0 // indirect

require (
	github.com/google/go-containerregistry v0.20.3
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/spf13/pflag v1.0.5
	golang.org/x/sync v0.11.0 // indirect
)
