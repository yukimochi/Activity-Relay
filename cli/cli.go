package main

import (
	"crypto/rsa"
	"net/url"

	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	keyloader "github.com/yukimochi/Activity-Relay/KeyLoader"
	state "github.com/yukimochi/Activity-Relay/State"
)

var hostname *url.URL
var hostkey *rsa.PrivateKey
var macServer *machinery.Server
var relayState state.RelayState

func initConfig() {
	viper.BindEnv("actor_pem")
	viper.BindEnv("relay_domain")
	viper.BindEnv("redis_url")
	hostkey, _ = keyloader.ReadPrivateKeyRSAfromPath(viper.GetString("actor_pem"))
	hostname, _ = url.Parse("https://" + viper.GetString("relay_domain"))
	redOption, err := redis.ParseURL(viper.GetString("redis_url"))
	if err != nil {
		panic(err)
	}
	redClient := redis.NewClient(redOption)
	var macConfig = &config.Config{
		Broker:          viper.GetString("redis_url"),
		DefaultQueue:    "relay",
		ResultBackend:   viper.GetString("redis_url"),
		ResultsExpireIn: 5,
	}
	macServer, err = machinery.NewServer(macConfig)
	if err != nil {
		panic(err)
	}
	relayState = state.NewState(redClient)
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
