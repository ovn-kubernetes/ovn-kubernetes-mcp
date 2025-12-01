# Get the Git repository root directory
GIT_ROOT := $(shell git rev-parse --show-toplevel)

export MCP_SERVER_PATH := $(GIT_ROOT)/_output/ovnk-mcp-server
export KUBECONFIG := $(HOME)/ovn.conf

# CONTAINER_RUNNABLE determines if the tests can be run inside a container. It checks to see if
# podman/docker is installed on the system.
PODMAN ?= $(shell podman -v > /dev/null 2>&1; echo $$?)
ifeq ($(PODMAN), 0)
CONTAINER_RUNTIME?=podman
else
CONTAINER_RUNTIME?=docker
endif
CONTAINER_RUNNABLE ?= $(shell $(CONTAINER_RUNTIME) -v > /dev/null 2>&1; echo $$?)

GOPATH ?= $(shell go env GOPATH)

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
	./hack/run-e2e.sh $(NVM_VERSION) $(NODE_VERSION) $(NPM_VERSION) "$(MCP_MODE)"

.PHONY: test-e2e
test-e2e: build deploy-kind-ovnk run-e2e undeploy-kind-ovnk

.PHONY: lint
lint:
ifeq ($(CONTAINER_RUNNABLE), 0)
	@GOPATH=${GOPATH} ./hack/lint.sh $(CONTAINER_RUNTIME) || { echo "lint failed! Try running 'make lint-fix'"; exit 1; }
else
	echo "linter can only be run within a container since it needs a specific golangci-lint version"; exit 1
endif

.PHONY: lint-fix
lint-fix:
ifeq ($(CONTAINER_RUNNABLE), 0)
	@GOPATH=${GOPATH} ./hack/lint.sh ${CONTAINER_RUNTIME} fix || { echo "ERROR: lint fix failed! There is a bug that changes file ownership to root \
	when this happens. To fix it, simply run 'chown -R <user>:<group> *' from the repo root."; exit 1; }
else
	echo "linter can only be run within a container since it needs a specific golangci-lint version"; exit 1
endif
