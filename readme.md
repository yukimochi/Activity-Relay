# Activity Relay Server

## Yet another powerful customizable ActivityPub relay server written in Go.

[![GitHub Actions](https://github.com/yukimochi/activity-relay/workflows/Test/badge.svg)](https://github.com/yukimochi/Activity-Relay)
[![codecov](https://codecov.io/gh/yukimochi/Activity-Relay/branch/master/graph/badge.svg)](https://codecov.io/gh/yukimochi/Activity-Relay)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fyukimochi%2FActivity-Relay.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fyukimochi%2FActivity-Relay?ref=badge_shield)

![Powered by Ayame](docs/ayame.png)

## Packages

 - `github.com/yukimochi/Activity-Relay`
 - `github.com/yukimochi/Activity-Relay/api`
 - `github.com/yukimochi/Activity-Relay/deliver`
 - `github.com/yukimochi/Activity-Relay/control`
 - `github.com/yukimochi/Activity-Relay/models`

## Requirements

 - [Redis](https://github.com/antirez/redis)

## Run

### API Server

```bash
relay --config /path/to/config.yml server
```

### Job Worker

```bash
relay --config /path/to/config.yml worker
```

### CLI Management Utility

```bash
relay --config /path/to/config.yml control
```

## Config

### YAML Format

```yaml config.yml
ACTOR_PEM: /var/lib/relay/actor.pem
REDIS_URL: redis://redis:6379

RELAY_BIND: 0.0.0.0:8080
RELAY_DOMAIN: relay.toot.yukimochi.jp
RELAY_SERVICENAME: YUKIMOCHI Toot Relay Service
JOB_CONCURRENCY: 50
# RELAY_SUMMARY: |

# RELAY_ICON: https://
# RELAY_IMAGE: https://
```

### Environment Variable

 This is **Optional** : When `config.yml` not exists, use environment variable.

 - ACTOR_PEM
 - REDIS_URL
 - RELAY_BIND
 - RELAY_DOMAIN
 - RELAY_SERVICENAME
 - JOB_CONCURRENCY
 - RELAY_SUMMARY
 - RELAY_ICON
 - RELAY_IMAGE

## [Document](https://github.com/yukimochi/Activity-Relay/wiki)

See [GitHub wiki](https://github.com/yukimochi/Activity-Relay/wiki) to build / install / manage relay.

## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fyukimochi%2FActivity-Relay.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fyukimochi%2FActivity-Relay?ref=badge_large)

## Project Sponsors

Thank you for your support.

### Monthly Donation

**[My Donator List](https://relay.toot.yukimochi.jp#patreon-list)**
  
#### Donation Platform
 - [Patreon](https://www.patreon.com/yukimochi)
 - [pixiv fanbox](https://yukimochi.fanbox.cc)
 - [fantia](https://fantia.jp/fanclubs/11264)
