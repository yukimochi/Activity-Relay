package main

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/go-redis/redis"
	cache "github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
	activitypub "github.com/yukimochi/Activity-Relay/ActivityPub"
	keyloader "github.com/yukimochi/Activity-Relay/KeyLoader"
	state "github.com/yukimochi/Activity-Relay/State"
)

// Actor : Relay's Actor
var Actor activitypub.Actor

// WebfingerResource : Relay's Webfinger resource
var WebfingerResource activitypub.WebfingerResource

var hostURL *url.URL
var hostPrivatekey *rsa.PrivateKey
var actorCache *cache.Cache
var machineryServer *machinery.Server
var relayState state.RelayState
var uaString string

func initConfig() {
	viper.BindEnv("actor_pem")
	viper.BindEnv("relay_domain")
	viper.BindEnv("relay_bind")
	viper.BindEnv("redis_url")
	hostURL, _ = url.Parse("https://" + viper.GetString("relay_domain"))
	hostPrivatekey, _ = keyloader.ReadPrivateKeyRSAfromPath(viper.GetString("actor_pem"))
	redisOption, err := redis.ParseURL(viper.GetString("redis_url"))
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(redisOption)
	machineryConfig := &config.Config{
		Broker:          viper.GetString("redis_url"),
		DefaultQueue:    "relay",
		ResultBackend:   viper.GetString("redis_url"),
		ResultsExpireIn: 5,
	}
	machineryServer, err = machinery.NewServer(machineryConfig)
	if err != nil {
		panic(err)
	}
	uaString = viper.GetString("relay_servicename") + " (golang net/http; Activity-Relay v0.2.2; " + hostURL.Host + ")"
	relayState = state.NewState(redisClient)
	actorCache = cache.New(5*time.Minute, 10*time.Minute)
	Actor.GenerateSelfKey(hostURL, &hostPrivatekey.PublicKey)
	WebfingerResource.GenerateFromActor(hostURL, &Actor)

	fmt.Println("Welcome to YUKIMOCHI Activity-Relay [Server]\n - Configrations")
	fmt.Println("RELAY DOMAIN : ", hostURL.Host)
	fmt.Println("REDIS URL : ", viper.GetString("redis_url"))
	fmt.Println("BIND ADDRESS : ", viper.GetString("relay_bind"))
	fmt.Println(" - Blocked Domain")
	domains, _ := redisClient.HKeys("relay:config:blockedDomain").Result()
	for _, domain := range domains {
		fmt.Println(domain)
	}
	fmt.Println(" - Limited Domain")
	domains, _ = redisClient.HKeys("relay:config:limitedDomain").Result()
	for _, domain := range domains {
		fmt.Println(domain)
	}
}

func main() {
	// Load Config
	initConfig()

	http.HandleFunc("/.well-known/webfinger", handleWebfinger)
	http.HandleFunc("/actor", handleActor)
	http.HandleFunc("/inbox", func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, decodeActivity)
	})

	http.ListenAndServe(viper.GetString("relay_bind"), nil)
}
