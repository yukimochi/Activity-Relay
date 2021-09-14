package control

import (
	"os"

	"github.com/RichardKnop/machinery/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yukimochi/Activity-Relay/models"
)

var (
	globalConfig *models.RelayConfig

	initProxy  = initializeProxy
	initProxyE = initializeProxyE

	// Actor : Relay's Actor
	Actor models.Actor

	relayState      models.RelayState
	machineryServer *machinery.Server
)

func BuildCommand(command *cobra.Command) {
	command.AddCommand(configCmdInit())
	command.AddCommand(domainCmdInit())
	command.AddCommand(followCmdInit())
}

func initializeProxy(function func(cmd *cobra.Command, args []string), cmd *cobra.Command, args []string) {
	initConfig(cmd)
	function(cmd, args)
}

func initializeProxyE(function func(cmd *cobra.Command, args []string) error, cmd *cobra.Command, args []string) error {
	initConfig(cmd)
	return function(cmd, args)
}

func initConfig(cmd *cobra.Command) error {
	var err error

	configPath := cmd.Flag("config").Value.String()
	file, err := os.Open(configPath)
	defer file.Close()

	if err == nil {
		viper.SetConfigType("yaml")
		viper.ReadConfig(file)
	} else {
		logrus.Warn("Config file not exist. Use environment variables.")

		viper.BindEnv("ACTOR_PEM")
		viper.BindEnv("REDIS_URL")
		viper.BindEnv("RELAY_BIND")
		viper.BindEnv("RELAY_DOMAIN")
		viper.BindEnv("RELAY_SERVICENAME")
		viper.BindEnv("JOB_CONCURRENCY")
		viper.BindEnv("RELAY_SUMMARY")
		viper.BindEnv("RELAY_ICON")
		viper.BindEnv("RELAY_IMAGE")
	}

	globalConfig, err = models.NewRelayConfig()
	if err != nil {
		logrus.Fatal(err)
	}

	initialize(globalConfig)

	return nil
}

func initialize(globalconfig *models.RelayConfig) error {
	var err error

	redisClient := globalConfig.RedisClient()
	relayState = models.NewState(redisClient, true)
	relayState.ListenNotify(nil)

	machineryServer, err = models.NewMachineryServer(globalConfig)
	if err != nil {
		return err
	}

	Actor = models.NewActivityPubActorFromSelfKey(globalConfig)

	return nil
}
