SHELL=/bin/bash
VERS=$(shell git describe --tags HEAD)

.PHONY: all
all: actor.pem config.yml relay Caddyfile
	@echo Things and stuff

.PHONY: setup
setup: genconf

.PHONY: genconf
genconf: config.yml

config.yml: .conf_redis .conf_domain .conf_desc config.yml.example | /var/lib/relay/actor.pem
	@sed \
		-e 's/^REDIS_URL:.*/REDIS_URL: redis:\/\/$(shell cat .conf_redis):6379/' \
		-e 's/^RELAY_DOMAIN:.*/RELAY_DOMAIN: $(shell cat .conf_domain)/' \
		-e "s/^RELAY_SERVICENAME:.*/RELAY_SERVICENAME: $(shell cat .conf_desc)/" < config.yml.example > $@

.conf_redis:
	@read -e -p "Redis host? [127.0.0.1] " r; \
                R=$$(echo $$r | tr '[:upper:]' '[:lower:]'); \
                if [ ! "$$R" ]; then \
			R="127.0.0.1"; \
                fi; echo "$$R" > $@

.conf_domain:
	@read -e -p "Relay Domain? [relay.wig.gl] " r; \
                R=$$(echo $$r | tr '[:upper:]' '[:lower:]'); \
                if [ ! "$$R" ]; then \
			R="relay.wig.gl"; \
                fi; echo "$$R" > $@

.conf_desc:
	@read -e -p "Description? [Honest Rob's Relay] " r; \
                R=$$(echo $$r | tr -d '"'); \
                if [ ! "$$R" ]; then \
			R="'Honest' Rob's Relay"; \
                fi; echo "$$R" > $@

WEBROOT = /var/lib/relay/webroot
APIPORT = $(shell awk -F: '/^RELAY_BIND:/ { print $$3 }' config.yml)
HOSTNAME = $(shell cat .conf_domain)

.PHONY: caddyfile
caddyfile: /etc/caddy/Caddyfile

/etc/caddy/Caddyfile: Caddyfile
	@if [ ! -d /etc/caddy ]; then echo "/etc/caddy not present, can not continue"; exit 1; fi
	@cp $< $@ && echo "Restarting caddy service" && systemctl restart caddy

Caddyfile: Caddyfile.tmpl config.yml $(WEBROOT)
	@sed -e 's!__WEBROOT__!$(WEBROOT)!' -e 's/__APIPORT__/$(APIPORT)/' -e 's/__HOSTNAME__/$(HOSTNAME)/' < Caddyfile.tmpl > $@

relay: $(wildcard **/*go)
	go build -o $@ -ldflags "-X main.version=$(VERS)" .

$(WEBROOT):
	@mkdir -p $(WEBROOT)
	@cp webroot/index.html $(WEBROOT)/index.html
	@chown -R relay.relay $(WEBROOT)

actor.pem: | /usr/bin/openssl
	/usr/bin/openssl genrsa -traditional > $@
	chmod 600 $@

/var/lib/relay/actor.pem: actor.pem | /var/lib/relay
	cp $< $@
	chown relay.relay $@
	chmod 600 $@

/var/lib/relay:
	groupadd --system relay
	useradd --system --gid relay --create-home --home-dir /var/lib/relay --shell /usr/sbin/nologin --comment "YUKIMOCHI Activity-Relay" relay
	chmod 755 /var/lib/relay
	passwd -l relay

/usr/bin/openssl:
	@echo "Can't continue - please install openssl"
	@exit 1
