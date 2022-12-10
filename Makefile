LOCAL_VERSION?=""
CI_RUN?=false
ADDITIONAL_BUILD_FLAGS=""

ifeq ($(CI_RUN), true)
	ADDITIONAL_BUILD_FLAGS="-test.short"
endif

ifneq ($(shell which semver-gen), "")
	LOCAL_VERSION="0.0.0-dev"
else
	LOCAL_VERSION=$(shell semver-gen generate -l -c config-release.yaml | sed -e 's|SEMVER ||g')
endif

.PHONY: help
help:  ## display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: all
all: build ## Build all targets

.PHONY: build
build: ## Build binary
	go build -o semver-gen -ldflags="-s -w -X main.PKG_VERSION=${LOCAL_VERSION}" *.go

# .PHONY: run
# run: build ## Build binary and execute it
# 	@./semver-gen

.PHONY: test
test: ## Run whole test suite
	@go test ./... $(ADDITIONAL_BUILD_FLAGS) -v -race -cover -coverprofile=coverage.out

.PHONY: update
update: ## Update dependencies
	@go get ./...

.PHONY: update-all
update-all: ## Update all dependencies and sub-packages
	@go get -u ./...
