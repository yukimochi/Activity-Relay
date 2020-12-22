package main

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/log"
	"github.com/go-redis/redis"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	activitypub "github.com/yukimochi/Activity-Relay/ActivityPub"
	keyloader "github.com/yukimochi/Activity-Relay/KeyLoader"
)

var (
	version string

	// Actor : Relay's Actor
	Actor activitypub.Actor

	hostURL         *url.URL
	hostPrivatekey  *rsa.PrivateKey
	redisClient     *redis.Client
	machineryServer *machinery.Server
	httpClient      *http.Client
)

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
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file is not exists. Use environment variables.")
		viper.BindEnv("actor_pem")
		viper.BindEnv("redis_url")
		viper.BindEnv("relay_bind")
		viper.BindEnv("relay_domain")
		viper.BindEnv("relay_servicename")
	} else {
		Actor.Summary = viper.GetString("relay_summary")
		Actor.Icon = activitypub.Image{URL: viper.GetString("relay_icon")}
		Actor.Image = activitypub.Image{URL: viper.GetString("relay_image")}
	}
	Actor.Name = viper.GetString("relay_servicename")

	hostURL, _ = url.Parse("https://" + viper.GetString("relay_domain"))
	hostPrivatekey, _ = keyloader.ReadPrivateKeyRSAfromPath(viper.GetString("actor_pem"))
	redisOption, err := redis.ParseURL(viper.GetString("redis_url"))
	if err != nil {
		panic(err)
	}
	redisClient = redis.NewClient(redisOption)
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
	httpClient = &http.Client{Timeout: time.Duration(5) * time.Second}

	Actor.GenerateSelfKey(hostURL, &hostPrivatekey.PublicKey)
	newNullLogger := NewNullLogger()
	log.DEBUG = newNullLogger

	fmt.Println("Welcome to YUKIMOCHI Activity-Relay [Worker]", version)
	fmt.Println(" - Configurations")
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
	worker := machineryServer.NewWorker(workerID.String(), 20)
	err = worker.Launch()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
