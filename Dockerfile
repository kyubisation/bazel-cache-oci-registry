FROM golang:1.24 AS build
RUN apt-get update && apt-get -y install musl musl-tools
COPY . .
RUN CGO_ENABLED=1 CC=musl-gcc go build --ldflags "-linkmode=external -extldflags=-static"

FROM scratch
COPY --from=build /etc/ssl/ /etc/ssl/
COPY --from=build /go/bazel-cache-oci-registry /bazel-cache-oci-registry

ENTRYPOINT [ "/bazel-cache-oci-registry" ]
CMD [ "server" ]