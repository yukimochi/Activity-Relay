package models

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
)

var globalConfig *RelayConfig
var relayState RelayState
var ch chan bool

func TestMain(m *testing.M) {
	var err error

	testConfigPath := "../misc/test/config.yml"
	file, _ := os.Open(testConfigPath)
	defer file.Close()

	viper.SetConfigType("yaml")
	viper.ReadConfig(file)
	viper.Set("ACTOR_PEM", "../misc/test/testKey.pem")
	viper.BindEnv("REDIS_URL")

	globalConfig, err = NewRelayConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	relayState = NewState(globalConfig.RedisClient(), true)
	ch = make(chan bool)
	relayState.ListenNotify(ch)
	relayState.RedisClient.FlushAll(context.TODO()).Result()
	code := m.Run()
	os.Exit(code)
}
