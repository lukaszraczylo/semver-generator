# syntax=docker/dockerfile:1.2.1-labs

FROM golang:1-alpine as baseimg

RUN apk add --no-cache make ca-certificates
WORKDIR /go/src/app
ENV GO111MODULE=on CGO_ENABLED=1 GOOS=linux
COPY . /go/src/app/
RUN make build

FROM alpine:latest
WORKDIR /go/src/app
RUN apk upgrade --available
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=baseimg /go/src/app/semver-gen .
COPY --from=baseimg /go/src/app/config-release.yaml config.yaml
COPY --from=baseimg /go/src/app/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]