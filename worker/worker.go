package main

import (
	"crypto/rsa"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
	"github.com/yukimochi/Activity-Relay/ActivityPub"
	"github.com/yukimochi/Activity-Relay/KeyLoader"
)

// Hostname : Hostname of Relay
var Hostname *url.URL

// Hostkey : PrivateKey of Relay
var Hostkey *rsa.PrivateKey

// Actor : Relay's Actor
var Actor activitypub.Actor

var redClient *redis.Client

func relayActivity(args ...string) error {
	inboxURL := args[0]
	body := args[1]
	err := activitypub.SendActivity(inboxURL, Actor.ID, []byte(body), Hostkey)
	if err != nil {
		domain, _ := url.Parse(inboxURL)
		mod, _ := redClient.HSetNX("relay:statistics:"+domain.Host, "last_error", err.Error()).Result()
		if mod {
			redClient.Expire("relay:statistics:"+domain.Host, time.Duration(time.Minute))
		}
	}
	return err
}

func registorActivity(args ...string) error {
	inboxURL := args[0]
	body := args[1]
	err := activitypub.SendActivity(inboxURL, Actor.ID, []byte(body), Hostkey)
	return err
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
	Hostkey, err = keyloader.ReadPrivateKeyRSAfromPath(pemPath)
	if err != nil {
		panic("Can't read Hostkey Pemfile")
	}
	Hostname, err = url.Parse("https://" + relayDomain)
	if err != nil {
		panic("Can't parse Relay Domain")
	}
	redClient = redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	Actor.GenerateSelfKey(Hostname, &Hostkey.PublicKey)

	var macConfig = &config.Config{
		Broker:          "redis://" + redisURL,
		DefaultQueue:    "relay",
		ResultBackend:   "redis://" + redisURL,
		ResultsExpireIn: 5,
	}
	server, err := machinery.NewServer(macConfig)
	if err != nil {
		panic(err.Error())
	}
	err = server.RegisterTask("registor", registorActivity)
	if err != nil {
		panic(err.Error())
	}
	err = server.RegisterTask("relay", relayActivity)
	if err != nil {
		panic(err.Error())
	}
	workerID := uuid.NewV4()
	worker := server.NewWorker(workerID.String(), 200)

	fmt.Println("Welcome to YUKIMOCHI Activity-Relay [Worker]\n - Configrations")
	fmt.Println("RELAY DOMAIN : ", relayDomain)
	fmt.Println("REDIS URL : ", redisURL)

	err = worker.Launch()
	if err != nil {
		fmt.Println(err.Error())
	}
}
