package main

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	state "github.com/yukimochi/Activity-Relay/State"
)

func TestMain(m *testing.M) {
	viper.Set("actor_pem", "misc/testKey.pem")
	viper.Set("relay_domain", "relay.yukimochi.example.org")
	initConfig()
	relayState = state.NewState(relayState.RedisClient, false)

	// Load Config
	code := m.Run()
	os.Exit(code)
	relayState.RedisClient.FlushAll().Result()
}
