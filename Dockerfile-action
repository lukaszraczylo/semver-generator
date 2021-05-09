# syntax=docker/dockerfile:1.2.1-labs

FROM golang:1-alpine as baseimg

RUN apk add make
WORKDIR /go/src/app
ENV GO111MODULE=on CGO_ENABLED=1 GOOS=linux
COPY . /go/src/app/
RUN make

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /go/src/app
COPY --from=baseimg /go/src/app/semver-gen .
COPY --from=baseimg /go/src/app/config-release.yaml config.yaml
COPY entrypoint.sh entrypoint.sh
ENTRYPOINT ["./entrypoint.sh"]