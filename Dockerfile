FROM golang:1-bullseye as baseimg
WORKDIR /go/src/app
COPY . /go/src/app/
RUN CGO_ENABLED=1 make build

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /go/src/app
ADD entrypoint.sh /entrypoint.sh
ADD config-release.yaml /go/src/app/config.yaml
COPY --from=baseimg /go/src/app/semver-gen .
ENTRYPOINT ["/entrypoint.sh"]