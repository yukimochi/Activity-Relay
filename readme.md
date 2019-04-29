# Activity Relay Server

## Yet another powerful customizable ActivityPub relay server written in Go.

[![CircleCI](https://circleci.com/gh/yukimochi/Activity-Relay.svg?style=svg)](https://circleci.com/gh/yukimochi/Activity-Relay)
[![codecov](https://codecov.io/gh/yukimochi/Activity-Relay/branch/master/graph/badge.svg)](https://codecov.io/gh/yukimochi/Activity-Relay)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fyukimochi%2FActivity-Relay.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fyukimochi%2FActivity-Relay?ref=badge_shield)

![Powered by Ayame](docs/ayame.png)

## Packages

 - `github.com/yukimochi/Activity-Relay`
 - `github.com/yukimochi/Activity-Relay/worker`
 - `github.com/yukimochi/Activity-Relay/cli`

## Requirements

 - [Redis](https://github.com/antirez/redis)

## Installation Manual

See [GitHub wiki](https://github.com/yukimochi/Activity-Relay/wiki)

## Environment Variable

 - `ACTOR_PEM` (ex. `/actor.pem`)
 - `RELAY_DOMAIN` (ex. `relay.toot.yukimochi.jp`)
 - `RELAY_SERVICENAME` (ex. `YUKIMOCHI Toot Relay Service`)
 - `RELAY_BIND` (ex. `0.0.0.0:8080`)
 - `REDIS_URL` (ex. `redis://127.0.0.1:6379/0`)

## Appended HTTP Header

 - `X-Relay-ActivityHost` Indicates an activity domain.

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fyukimochi%2FActivity-Relay.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fyukimochi%2FActivity-Relay?ref=badge_large)