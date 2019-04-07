package main

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestMain(m *testing.M) {
	viper.Set("actor_pem", "../misc/testKey.pem")
	viper.Set("relay_domain", "relay.yukimochi.example.org")
	initConfig()

	relayState.RedisClient.FlushAll().Result()
	code := m.Run()
	os.Exit(code)
	relayState.RedisClient.FlushAll().Result()
}
