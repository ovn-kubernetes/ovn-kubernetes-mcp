# Build stage – use Go image to match go.mod (set via build-arg from Makefile)
ARG GOLANG_IMAGE=quay.io/projectquay/golang
ARG GOLANG_VERSION=1.24
FROM ${GOLANG_IMAGE}:${GOLANG_VERSION} AS builder
WORKDIR /build

ENV CGO_ENABLED=0
ENV GOPATH=/go
ENV PATH="${GOPATH}/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

# Runtime stage – minimal image
FROM alpine:3.23
RUN apk add --no-cache ca-certificates
RUN adduser -D -u 1000 -g root mcp
USER 1000:0

COPY --from=builder /build/_output/ovnk-mcp-server /usr/local/bin/ovnk-mcp-server
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/ovnk-mcp-server"]
CMD ["--transport=http", "--host=0.0.0.0", "--port=8080", "--mode=live-cluster"]
