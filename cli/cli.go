package main

import (
	"crypto/rsa"
	"fmt"
	"net/url"
	"os"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
	"github.com/yukimochi/Activity-Relay/KeyLoader"
	"github.com/yukimochi/Activity-Relay/State"
)

var hostname *url.URL
var hostkey *rsa.PrivateKey
var redClient *redis.Client
var macServer *machinery.Server
var relayState state.RelayState

func buildNewCmd() *cobra.Command {
	var app = &cobra.Command{}
	app.AddCommand(domainCmdInit())
	app.AddCommand(followCmdInit())
	app.AddCommand(configCmdInit())
	return app
}

func main() {
	pemPath := os.Getenv("ACTOR_PEM")
	if pemPath == "" {
		panic("Require ACTOR_PEM environment variable.")
	}
	relayDomain := os.Getenv("RELAY_DOMAIN")
	if relayDomain == "" {
		panic("Require RELAY_DOMAIN environment variable.")
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "127.0.0.1:6379"
	}

	var err error
	hostkey, err = keyloader.ReadPrivateKeyRSAfromPath(pemPath)
	if err != nil {
		panic("Can't read Hostkey Pemfile")
	}
	hostname, err = url.Parse("https://" + relayDomain)
	if err != nil {
		panic("Can't parse Relay Domain")
	}
	redClient = redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	var macConfig = &config.Config{
		Broker:          "redis://" + redisURL,
		DefaultQueue:    "relay",
		ResultBackend:   "redis://" + redisURL,
		ResultsExpireIn: 5,
	}

	macServer, err = machinery.NewServer(macConfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	relayState = state.NewState(redClient)

	var app = buildNewCmd()
	app.Execute()
}
