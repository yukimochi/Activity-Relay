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
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	activitypub "github.com/yukimochi/Activity-Relay/ActivityPub"
	keyloader "github.com/yukimochi/Activity-Relay/KeyLoader"
)

// Actor : Relay's Actor
var Actor activitypub.Actor

var hostURL *url.URL
var hostPrivatekey *rsa.PrivateKey
var machineryServer *machinery.Server
var redisClient *redis.Client
var uaString string

func relayActivity(args ...string) error {
	inboxURL := args[0]
	body := args[1]
	err := sendActivity(inboxURL, Actor.ID, []byte(body), hostPrivatekey)
	if err != nil {
		domain, _ := url.Parse(inboxURL)
		mod, _ := redisClient.HSetNX("relay:statistics:"+domain.Host, "last_error", err.Error()).Result()
		if mod {
			redisClient.Expire("relay:statistics:"+domain.Host, time.Duration(time.Minute))
		}
	}
	return err
}

func registorActivity(args ...string) error {
	inboxURL := args[0]
	body := args[1]
	err := sendActivity(inboxURL, Actor.ID, []byte(body), hostPrivatekey)
	return err
}

func initConfig() {
	viper.BindEnv("actor_pem")
	viper.BindEnv("relay_domain")
	viper.BindEnv("relay_servicename")
	viper.BindEnv("redis_url")
	hostURL, _ = url.Parse("https://" + viper.GetString("relay_domain"))
	hostPrivatekey, _ = keyloader.ReadPrivateKeyRSAfromPath(viper.GetString("actor_pem"))
	redisClient = redis.NewClient(&redis.Options{
		Addr: viper.GetString("redis_url"),
	})
	machineryConfig := &config.Config{
		Broker:          "redis://" + viper.GetString("redis_url"),
		DefaultQueue:    "relay",
		ResultBackend:   "redis://" + viper.GetString("redis_url"),
		ResultsExpireIn: 5,
	}
	machineryServer, _ = machinery.NewServer(machineryConfig)
	uaString = viper.GetString("relay_servicename") + " (golang net/http; Activity-Relay v0.2.0; " + hostURL.Host + ")"
	Actor.GenerateSelfKey(hostURL, &hostPrivatekey.PublicKey)

	fmt.Println("Welcome to YUKIMOCHI Activity-Relay [Worker]\n - Configrations")
	fmt.Println("RELAY DOMAIN : ", hostURL.Host)
	fmt.Println("REDIS URL : ", viper.GetString("redis_url"))
}

func main() {
	initConfig()

	err := machineryServer.RegisterTask("registor", registorActivity)
	if err != nil {
		panic(err.Error())
	}
	err = machineryServer.RegisterTask("relay", relayActivity)
	if err != nil {
		panic(err.Error())
	}
	workerID := uuid.NewV4()
	worker := machineryServer.NewWorker(workerID.String(), 200)
	err = worker.Launch()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
