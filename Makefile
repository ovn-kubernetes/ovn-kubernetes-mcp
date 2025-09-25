# Get the Git repository root directory
GIT_ROOT := $(shell git rev-parse --show-toplevel)

export MCP_SERVER_PATH := $(GIT_ROOT)/_output/ovnk-mcp-server
export KUBECONFIG := $(HOME)/ovn.conf

.PHONY: build
build:
	go build -o $(MCP_SERVER_PATH) cmd/ovnk-mcp-server/main.go

.PHONY: clean
clean:
	rm -Rf _output/

.PHONY: test
test:
	go test -v ./...

.PHONY: deploy-kind-ovnk
deploy-kind-ovnk:
	./hack/deploy-kind-ovnk.sh

.PHONY: undeploy-kind-ovnk
undeploy-kind-ovnk:
	./hack/undeploy-kind-ovnk.sh
