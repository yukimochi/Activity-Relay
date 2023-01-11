all: relay

relay:
	go build -o relay -ldflags "-X main.version=$(git describe --tags HEAD)"

clean:
	rm relay