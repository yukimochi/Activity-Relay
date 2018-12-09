package main

import (
	"net/url"
	"os"
	"testing"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/go-redis/redis"
	"github.com/yukimochi/Activity-Relay/KeyLoader"
	"github.com/yukimochi/Activity-Relay/State"
)

func TestMain(m *testing.M) {
	os.Setenv("ACTOR_PEM", "misc/testKey.pem")
	os.Setenv("RELAY_DOMAIN", "relay.yukimochi.example.org")
	pemPath := os.Getenv("ACTOR_PEM")
	relayDomain := os.Getenv("RELAY_DOMAIN")
	redisURL := os.Getenv("REDIS_URL")
	hostkey, _ = keyloader.ReadPrivateKeyRSAfromPath(pemPath)
	hostname, _ = url.Parse("https://" + relayDomain)
	redClient = redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	var macConfig = &config.Config{
		Broker:          "redis://" + redisURL,
		DefaultQueue:    "relay",
		ResultBackend:   "redis://" + redisURL,
		ResultsExpireIn: 5,
	}
	macServer, _ = machinery.NewServer(macConfig)

	Actor.GenerateSelfKey(hostname, &hostkey.PublicKey)
	WebfingerResource.GenerateFromActor(hostname, &Actor)

	// Load Config
	redClient.FlushAll().Result()
	relayState = state.NewState(redClient)
	code := m.Run()
	os.Exit(code)
	redClient.FlushAll().Result()
}
