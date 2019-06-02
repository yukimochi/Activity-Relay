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

## Configration

### `config.yml`

```yaml config.yml
actor_pem: /actor.pem
redis_url: redis://redis:6379

relay_bind: 0.0.0.0:8080
relay_domain: relay.toot.yukimochi.jp
relay_servicename: YUKIMOCHI Toot Relay Service
# relay_summary: |

# relay_icon: https://
# relay_image: https://
```

### `Environment Variable`

 This is **Optional** : When `config.yml` not exists, use environment variable.

 - `ACTOR_PEM` (ex. `/actor.pem`)
 - `REDIS_URL` (ex. `redis://127.0.0.1:6379/0`)
 - `RELAY_BIND` (ex. `0.0.0.0:8080`)
 - `RELAY_DOMAIN` (ex. `relay.toot.yukimochi.jp`)
 - `RELAY_SERVICENAME` (ex. `YUKIMOCHI Toot Relay Service`)

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fyukimochi%2FActivity-Relay.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fyukimochi%2FActivity-Relay?ref=badge_large)