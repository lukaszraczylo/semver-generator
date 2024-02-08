FROM golang:1-bullseye as baseimg
WORKDIR /go/src/app
COPY . /go/src/app/
RUN CGO_ENABLED=1 make build

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /go/src/app
COPY --from=baseimg /go/src/app/semver-gen .
COPY --from=baseimg /go/src/app/config-release.yaml config.yaml
COPY --from=baseimg /go/src/app/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]