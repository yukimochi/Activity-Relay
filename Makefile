all: relay

relay: api/* control/* deliver/* models/* api/templates/* main.go
	go build -o relay -ldflags "-X main.version=$(git describe --tags HEAD)"

install: relay
	cp -f relay /usr/bin/relay

install-systemd: relay install
	cp misc/dist/init/relay-api.service /etc/systemd/system
	cp misc/dist/init/relay-worker.service /etc/systemd/system
	systemctl daemon-reload

clean:
	rm relay