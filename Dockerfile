FROM --platform=$BUILDPLATFORM golang:1.27-alpine3.24 AS build

WORKDIR /Activity-Relay
COPY . /Activity-Relay

ARG TARGETOS
ARG TARGETARCH

RUN  mkdir -p /rootfs/usr/bin && \
     apk add -U --no-cache git && \
     CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
     go build -o /rootfs/usr/bin/relay -ldflags "-X main.version=$(git describe --tags HEAD | sed -r 's/v(.*)/\1/')" .

FROM alpine:3.24

COPY --from=build /rootfs/usr/bin /usr/bin
RUN  chmod +x /usr/bin/relay && \
     apk add -U --no-cache ca-certificates
