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
	hostPrivatekey, _ = keyloader.ReadPrivateKeyRSAfromPath(pemPath)
	hostURL, _ = url.Parse("https://" + relayDomain)
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	var macConfig = &config.Config{
		Broker:          "redis://" + redisURL,
		DefaultQueue:    "relay",
		ResultBackend:   "redis://" + redisURL,
		ResultsExpireIn: 5,
	}
	machineryServer, _ = machinery.NewServer(macConfig)

	Actor.GenerateSelfKey(hostURL, &hostPrivatekey.PublicKey)
	WebfingerResource.GenerateFromActor(hostURL, &Actor)

	// Load Config
	redisClient.FlushAll().Result()
	relayState = state.NewState(redisClient)
	code := m.Run()
	os.Exit(code)
	redisClient.FlushAll().Result()
}
