package main

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/go-redis/redis"
	"github.com/yukimochi/Activity-Relay/ActivityPub"
	"github.com/yukimochi/Activity-Relay/KeyLoader"
)

// Actor : Relay's Actor
var Actor activitypub.Actor

// WebfingerResource : Relay's Webfinger resource
var WebfingerResource activitypub.WebfingerResource

type relayConfig struct {
	blockService bool
}

var hostname *url.URL
var hostkey *rsa.PrivateKey
var redClient *redis.Client
var macServer *machinery.Server
var relConfig relayConfig

func loadConfig() relayConfig {
	blockService, err := redClient.HGet("relay:config", "block_service").Result()
	if err != nil {
		redClient.HSet("relay:config", "block_service", 0)
		blockService = "0"
	}
	return relayConfig{
		blockService: blockService == "1",
	}
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
	relayBind := os.Getenv("RELAY_BIND")
	if relayBind == "" {
		relayBind = "0.0.0.0:8080"
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
		fmt.Println(err)
	}

	Actor = activitypub.GenerateActor(hostname, &hostkey.PublicKey)
	WebfingerResource = activitypub.GenerateWebfingerResource(hostname, &Actor)

	// Load Config
	relConfig = loadConfig()

	http.HandleFunc("/.well-known/webfinger", handleWebfinger)
	http.HandleFunc("/actor", handleActor)
	http.HandleFunc("/inbox", handleInbox)

	fmt.Println("Welcome to YUKIMOCHI Activity-Relay [Server]\n - Configrations")
	fmt.Println("RELAY DOMAIN : ", relayDomain)
	fmt.Println("REDIS URL : ", redisURL)
	fmt.Println("BIND ADDRESS : ", relayBind)
	fmt.Println(" - Blocked Domain")
	domains, _ := redClient.HKeys("relay:config:blockedDomain").Result()
	for _, domain := range domains {
		fmt.Println(domain)
	}
	fmt.Println(" - Limited Domain")
	domains, _ = redClient.HKeys("relay:config:limitedDomain").Result()
	for _, domain := range domains {
		fmt.Println(domain)
	}
	http.ListenAndServe(relayBind, nil)
}
