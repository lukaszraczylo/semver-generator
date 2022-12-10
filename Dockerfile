FROM golang:1-bullseye as baseimg

WORKDIR /go/src/app
ENV GO111MODULE=on CGO_ENABLED=1 GOOS=linux
COPY . /go/src/app/
RUN make build

FROM ubuntu:jammy
WORKDIR /go/src/app
COPY --from=baseimg /go/src/app/semver-gen .
COPY --from=baseimg /go/src/app/config-release.yaml config.yaml
COPY --from=baseimg /go/src/app/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]