all: build

build:
	@go build -o semver *.go

run: build
	@./semver

test:
	@go test ./... -v -race -cover