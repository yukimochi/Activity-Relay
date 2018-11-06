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

var Hostname *url.URL
var Hostkey *rsa.PrivateKey
var RedClient *redis.Client
var MacServer *machinery.Server

// Actor : Relay's Actor
var Actor activitypub.Actor

// WebfingerResource : Relay's Webfinger resource
var WebfingerResource activitypub.WebfingerResource

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
	Hostkey, err = keyloader.ReadPrivateKeyRSAfromPath(pemPath)
	if err != nil {
		panic("Can't read Hostkey Pemfile")
	}
	Hostname, err = url.Parse("https://" + relayDomain)
	if err != nil {
		panic("Can't parse Relay Domain")
	}
	RedClient = redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	var macConfig = &config.Config{
		Broker:          "redis://" + redisURL,
		DefaultQueue:    "relay",
		ResultBackend:   "redis://" + redisURL,
		ResultsExpireIn: 5,
	}

	MacServer, err = machinery.NewServer(macConfig)
	if err != nil {
		fmt.Println(err)
	}

	Actor = activitypub.GenerateActor(Hostname, &Hostkey.PublicKey)
	WebfingerResource = activitypub.GenerateWebfingerResource(Hostname, &Actor)

	http.HandleFunc("/.well-known/webfinger", handleWebfinger)
	http.HandleFunc("/actor", handleActor)
	http.HandleFunc("/inbox", handleInbox)

	fmt.Println("Welcome to YUKIMOCHI Activity-Relay [Server]\n - Configrations")
	fmt.Println("RELAY DOMAIN : ", relayDomain)
	fmt.Println("REDIS URL : ", redisURL)
	fmt.Println("BIND ADDRESS : ", relayBind)
	http.ListenAndServe(relayBind, nil)
}
