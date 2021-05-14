LOCAL_VERSION?=$(shell semver-gen generate -l -c config-release.yaml | sed -e 's|SEMVER ||g')

all: build

build:
	go build -o semver-gen -ldflags="-s -w -X main.PKG_VERSION=${LOCAL_VERSION}" *.go

run: build
	@./semver-gen

test:
	@go test ./... -v -race -cover -coverprofile=coverage.out

update:
	@go get -u ./...