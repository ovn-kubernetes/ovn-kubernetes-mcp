# CONTAINER_RUNNABLE determines if the tests can be run inside a container. It checks to see if
# podman/docker is installed on the system.
PODMAN ?= $(shell podman -v > /dev/null 2>&1; echo $$?)
ifeq ($(PODMAN), 0)
CONTAINER_RUNTIME?=podman
else
CONTAINER_RUNTIME?=docker
endif
CONTAINER_RUNNABLE ?= $(shell $(CONTAINER_RUNTIME) -v > /dev/null 2>&1; echo $$?)

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
	@GOPATH=${GOPATH} ./hack/lint.sh $(CONTAINER_RUNTIME) fix || { echo "ERROR: lint fix failed! There is a bug that changes file ownership to root \
	when this happens. To fix it, simply run 'chown -R <user>:<group> *' from the repo root."; exit 1; }
else
	echo "linter can only be run within a container since it needs a specific golangci-lint version"; exit 1
endif
