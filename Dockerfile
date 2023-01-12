FROM golang:1.19-alpine AS build

WORKDIR /Activity-Relay
COPY . /Activity-Relay

RUN  mkdir -p /rootfs/usr/bin && \
     apk add -U --no-cache git && \
     go build -o /rootfs/usr/bin/relay -ldflags "-X main.version=$(git describe --tags HEAD)" .

FROM alpine:3.17.1

COPY --from=build /rootfs/usr/bin /usr/bin
RUN  chmod +x /usr/bin/relay && \
     apk add -U --no-cache ca-certificates
