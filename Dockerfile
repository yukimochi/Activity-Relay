FROM golang:alpine AS build

WORKDIR /go/src/github.com/yukimochi/Activity-Relay
COPY . /go/src/github.com/yukimochi/Activity-Relay

RUN  mkdir -p /rootfs/usr/bin && \
     apk add -U --no-cache git && \
     go get -u github.com/golang/dep/cmd/dep && \
     dep ensure && \
     go build -o /rootfs/usr/bin/server . && \
     go build -o /rootfs/usr/bin/worker ./worker

FROM alpine

COPY --from=build /rootfs/usr/bin /usr/bin
RUN  chmod +x /usr/bin/server /usr/bin/worker && \
     apk add -U --no-cache ca-certificates
