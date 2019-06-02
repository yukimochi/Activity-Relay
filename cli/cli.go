package main

import (
	"crypto/rsa"
	"fmt"
	"net/url"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	activitypub "github.com/yukimochi/Activity-Relay/ActivityPub"
	keyloader "github.com/yukimochi/Activity-Relay/KeyLoader"
	state "github.com/yukimochi/Activity-Relay/State"
)

var (
	version string

	// Actor : Relay's Actor
	Actor activitypub.Actor

	hostname        *url.URL
	hostkey         *rsa.PrivateKey
	relayState      state.RelayState
	machineryServer *machinery.Server
)

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

	hostname, err = url.Parse("https://" + viper.GetString("relay_domain"))
	if err != nil {
		panic(err)
	}
	hostkey, err := keyloader.ReadPrivateKeyRSAfromPath(viper.GetString("actor_pem"))
	if err != nil {
		panic(err)
	}
	redisOption, err := redis.ParseURL(viper.GetString("redis_url"))
	if err != nil {
		panic(err)
	}
	redisClient := redis.NewClient(redisOption)
	relayState = state.NewState(redisClient)
	var machineryConfig = &config.Config{
		Broker:          viper.GetString("redis_url"),
		DefaultQueue:    "relay",
		ResultBackend:   viper.GetString("redis_url"),
		ResultsExpireIn: 5,
	}
	machineryServer, err = machinery.NewServer(machineryConfig)
	if err != nil {
		panic(err)
	}

	Actor.GenerateSelfKey(hostname, &hostkey.PublicKey)
}

func buildNewCmd() *cobra.Command {
	var app = &cobra.Command{}
	app.AddCommand(domainCmdInit())
	app.AddCommand(followCmdInit())
	app.AddCommand(configCmdInit())
	return app
}

func main() {
	initConfig()
	var app = buildNewCmd()
	app.Execute()
}
