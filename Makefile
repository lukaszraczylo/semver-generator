LOCAL_VERSION?=$(shell semver-gen generate -l -c config-release.yaml | sed -e 's|SEMVER ||g')
CI_RUN?=false
ADDITIONAL_BUILD_FLAGS=""

ifeq ($(CI_RUN), true)
	ADDITIONAL_BUILD_FLAGS="-test.short"
endif

all: build

build:
	go build -o semver-gen -ldflags="-s -w -X main.PKG_VERSION=${LOCAL_VERSION}" *.go

run: build
	@./semver-gen

test:
	@go test ./... $(ADDITIONAL_BUILD_FLAGS) -v -race -cover -coverprofile=coverage.out

update:
	@go get -u ./...
