FROM public.ecr.aws/docker/library/golang:1.25.0-alpine3.22 AS build

WORKDIR /Activity-Relay
COPY . /Activity-Relay

RUN  mkdir -p /rootfs/usr/bin && \
     apk add -U --no-cache git && \
     go build -o /rootfs/usr/bin/relay -ldflags "-X main.version=$(git describe --tags HEAD | sed -r 's/v(.*)/\1/')" .

FROM public.ecr.aws/docker/library/alpine:3.22.1

COPY --from=build /rootfs/usr/bin /usr/bin
RUN  chmod +x /usr/bin/relay && \
     apk add -U --no-cache ca-certificates
