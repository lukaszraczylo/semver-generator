all: build

build:
	@go build -o semver-gen *.go

run: build
	@./semver-gen

test:
	@go test ./... -v -race -cover