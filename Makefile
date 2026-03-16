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

# Container image build targets (use IMAGE to override tag, e.g. make build-image IMAGE=quay.io/myorg/ovnk-mcp-server:v1.0)
IMAGE ?= localhost/ovnk-mcp-server:dev
export IMAGE
GOLANG_IMAGE ?= quay.io/projectquay/golang
GOLANG_VERSION ?= 1.24
KUSTOMIZE_VERSION ?= v5.8.1
K8S_VERSION ?= v1.34.1

.PHONY: build-image
build-image:
	$(CONTAINER_RUNTIME) build -f Dockerfile \
		--build-arg GOLANG_IMAGE=$(GOLANG_IMAGE) \
		--build-arg GOLANG_VERSION=$(GOLANG_VERSION) \
		-t $(IMAGE) .

LOCALBIN ?= $(GIT_ROOT)/_output/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) go install -mod=mod $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef

# Prefer kustomize on PATH; install a pinned binary under LOCALBIN only when none is found.
KUSTOMIZE_PATH := $(shell command -v kustomize 2>/dev/null)
ifeq ($(strip $(KUSTOMIZE_PATH)),)
KUSTOMIZE ?= $(LOCALBIN)/kustomize-$(KUSTOMIZE_VERSION)
else
KUSTOMIZE ?= $(KUSTOMIZE_PATH)
endif
export KUSTOMIZE

.PHONY: kustomize
ifeq ($(strip $(KUSTOMIZE_PATH)),)
kustomize: $(KUSTOMIZE) ## Download kustomize to LOCALBIN when not on PATH.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))
else
kustomize: ## Using kustomize from PATH ($(KUSTOMIZE)); skip LOCALBIN install.
	@true
endif

# Prefer kubectl on PATH; otherwise download the release binary to LOCALBIN (dl.k8s.io).
KUBECTL_PATH := $(shell command -v kubectl 2>/dev/null)
ifeq ($(strip $(KUBECTL_PATH)),)
KUBECTL ?= $(LOCALBIN)/kubectl-$(K8S_VERSION)
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')
else
KUBECTL ?= $(KUBECTL_PATH)
endif
export KUBECTL

.PHONY: kubectl
ifeq ($(strip $(KUBECTL_PATH)),)
kubectl: $(KUBECTL) ## Download kubectl to LOCALBIN when not on PATH.
$(KUBECTL): $(LOCALBIN)
	@test -s $(KUBECTL) || { \
	set -e; \
	url="https://dl.k8s.io/$(K8S_VERSION)/bin/$(OS)/$(ARCH)/kubectl"; \
	echo "Downloading kubectl $(K8S_VERSION) ($${url}) ..."; \
	curl -fsSL "$${url}" -o "$(KUBECTL).tmp"; \
	curl -fsSL "$${url}.sha256" -o "$(KUBECTL).sha256.tmp"; \
	if command -v sha256sum >/dev/null 2>&1; then \
		echo "$$(cat "$(KUBECTL).sha256.tmp")  $(KUBECTL).tmp" | sha256sum --check -; \
	else \
		echo "$$(cat "$(KUBECTL).sha256.tmp")  $(KUBECTL).tmp" | shasum -a 256 -c -; \
	fi; \
	chmod +x "$(KUBECTL).tmp" && mv -f "$(KUBECTL).tmp" "$(KUBECTL)"; \
	rm -f "$(KUBECTL).sha256.tmp"; \
	}
else
kubectl: ## Using kubectl from PATH ($(KUBECTL)); skip LOCALBIN install.
	@true
endif

# Ephemeral copy of config/ for deploy-k8s so kustomize edit does not modify tracked files.
DEPLOY_K8S_CONFIG := $(GIT_ROOT)/_output/deploy-k8s-config

.PHONY: deploy-k8s
deploy-k8s: kustomize kubectl
	rm -rf $(DEPLOY_K8S_CONFIG)
	mkdir -p $(DEPLOY_K8S_CONFIG)
	cp -a $(GIT_ROOT)/config/. $(DEPLOY_K8S_CONFIG)/
	cd $(DEPLOY_K8S_CONFIG) && $(KUSTOMIZE) edit set image localhost/ovnk-mcp-server=$(IMAGE)
	$(KUSTOMIZE) build $(DEPLOY_K8S_CONFIG) | $(KUBECTL) apply -f -

.PHONY: undeploy-k8s
undeploy-k8s: kustomize kubectl
	$(KUSTOMIZE) build $(GIT_ROOT)/config | $(KUBECTL) delete --ignore-not-found=true -f -

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
	@$(GIT_ROOT)/hack/deploy-kind-ovnk.sh

.PHONY: undeploy-kind-ovnk
undeploy-kind-ovnk:
	@$(GIT_ROOT)/hack/undeploy-kind-ovnk.sh

NVM_VERSION := 0.40.3
NODE_VERSION := 22.20.0
NPM_VERSION := 11.6.1
GINKGO_VERSION := v2.26.0
MCP_MODE ?= live-cluster

.PHONY: run-e2e
run-e2e:
	@$(GIT_ROOT)/hack/run-e2e.sh $(NVM_VERSION) $(NODE_VERSION) $(NPM_VERSION) $(GINKGO_VERSION) "$(MCP_MODE)" "$(FOCUS)"

.PHONY: test-e2e
test-e2e: build
	if [ "$(MCP_MODE)" = "live-cluster" ]; then $(MAKE) deploy-kind-ovnk || exit 1; fi; \
	$(MAKE) run-e2e || EXIT_CODE=$$?; \
	if [ "$(MCP_MODE)" = "live-cluster" ]; then $(MAKE) undeploy-kind-ovnk || exit 1; fi; \
	exit $${EXIT_CODE:-0}

.PHONY: lint
lint:
ifeq ($(CONTAINER_RUNNABLE), 0)
	@GOPATH=${GOPATH} $(GIT_ROOT)/hack/lint.sh $(CONTAINER_RUNTIME) || { echo "lint failed! Try running 'make lint-fix'"; exit 1; }
else
	echo "linter can only be run within a container since it needs a specific golangci-lint version"; exit 1
endif

.PHONY: update-readme-tools
update-readme-tools:
	go run $(GIT_ROOT)/hack/gen-readme-tools.go

.PHONY: lint-fix
lint-fix:
ifeq ($(CONTAINER_RUNNABLE), 0)
	@GOPATH=${GOPATH} $(GIT_ROOT)/hack/lint.sh ${CONTAINER_RUNTIME} fix || { echo "ERROR: lint fix failed! There is a bug that changes file ownership to root \
	when this happens. To fix it, simply run 'chown -R <user>:<group> *' from the repo root."; exit 1; }
else
	echo "linter can only be run within a container since it needs a specific golangci-lint version"; exit 1
endif
