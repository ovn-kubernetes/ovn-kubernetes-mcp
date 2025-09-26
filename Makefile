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

EXCLUDE_DIRS ?= test/
TEST_PKGS := $$(go list ./... | grep -v $(EXCLUDE_DIRS))

.PHONY: test
test:
	go test -v $(TEST_PKGS)

.PHONY: deploy-kind-ovnk
deploy-kind-ovnk:
	./hack/deploy-kind-ovnk.sh

.PHONY: undeploy-kind-ovnk
undeploy-kind-ovnk:
	./hack/undeploy-kind-ovnk.sh

NVM_VERSION := 0.40.3
NODE_VERSION := 22.20.0
NPM_VERSION := 11.6.1

.PHONY: run-e2e
run-e2e:
	./hack/run-e2e.sh $(NVM_VERSION) $(NODE_VERSION) $(NPM_VERSION)

.PHONY: test-e2e
test-e2e: build deploy-kind-ovnk run-e2e undeploy-kind-ovnk
