# Activity Relay Server

## Yet another powerful customizable ActivityPub relay server written in Go.

![Powered by Ayame](docs/ayame.png)

## Packages

 - `github.com/yukimochi/Activity-Relay`
 - `github.com/yukimochi/Activity-Relay/worker`

## Requirements

 - [Redis](https://github.com/antirez/redis)

## Environment Variable

 - `ACTOR_PEM` (ex. `/actor.pem`)
 - `RELAY_DOMAIN` (ex. `relay.toot.yukimochi.jp`)
 - `RELAY_BIND` (ex. `0.0.0.0:8080`)
 - `REDIS_URL` (ex. `127.0.0.1:6379`)
