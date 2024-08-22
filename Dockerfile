FROM golang:1-bullseye as baseimg
WORKDIR /go/src/app
COPY . /go/src/app/
RUN CGO_ENABLED=1 make build

FROM ubuntu:jammy
COPY --from=baseimg /go/src/app/semver-gen /go/src/app/semver-gen
COPY --from=baseimg /go/src/app/config-release.yaml /go/src/app/config.yaml
COPY --from=baseimg /go/src/app/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]