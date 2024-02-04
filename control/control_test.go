package control

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yukimochi/Activity-Relay/models"
)

func TestMain(m *testing.M) {
	var err error

	testConfigPath := "../misc/test/config.yml"
	file, _ := os.Open(testConfigPath)
	defer file.Close()

	viper.SetConfigType("yaml")
	viper.ReadConfig(file)
	viper.Set("ACTOR_PEM", "../misc/test/testKey.pem")
	viper.BindEnv("REDIS_URL")

	GlobalConfig, err = models.NewRelayConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = initialize()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	RelayState = models.NewState(GlobalConfig.RedisClient(), false)
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	InitProxy = emptyProxy
	InitProxyE = emptyProxyE

	code := m.Run()
	os.Exit(code)
}

func emptyProxy(function func(cmd *cobra.Command, args []string), cmd *cobra.Command, args []string) {
	function(cmd, args)
}

func emptyProxyE(function func(cmd *cobra.Command, args []string) error, cmd *cobra.Command, args []string) error {
	return function(cmd, args)
}
