package control

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yukimochi/Activity-Relay/models"
)

const (
	PersonOnly models.Config = iota
	ManuallyAccept
)

func configCmdInit() *cobra.Command {
	var config = &cobra.Command{
		Use:   "config",
		Short: "Manage configuration for relay",
		Long:  "Enable/disable relay customize and import/export relay database.",
	}

	var configList = &cobra.Command{
		Use:   "list",
		Short: "List all relay configuration",
		Long:  "List all relay configuration.",
		Run: func(cmd *cobra.Command, args []string) {
			InitProxy(listConfig, cmd, args)
		},
	}
	config.AddCommand(configList)

	var configExport = &cobra.Command{
		Use:   "export",
		Short: "Export all relay information",
		Long:  "Export all relay information by JSON format.",
		Run: func(cmd *cobra.Command, args []string) {
			InitProxy(exportConfig, cmd, args)
		},
	}
	config.AddCommand(configExport)

	var configImport = &cobra.Command{
		Use:   "import [flags]",
		Short: "Import all relay information",
		Long:  "Import all relay information from JSON.",
		Run: func(cmd *cobra.Command, args []string) {
			InitProxy(importConfig, cmd, args)
		},
	}
	configImport.Flags().String("data", "", "JSON String")
	configImport.MarkFlagRequired("data")
	config.AddCommand(configImport)

	var configEnable = &cobra.Command{
		Use:   "enable",
		Short: "Enable relay configuration",
		Long: `Enable relay configuration.
 - person-only
	Blocking feature for service-type actor.
 - manually-accept
	Enable manually accept follow request.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(configEnable, cmd, args)
		},
	}
	config.AddCommand(configEnable)

	var configDisable = &cobra.Command{
		Use:   "disable",
		Short: "Disable relay configuration",
		Long: `Disable relay configuration.
 - person-only
	Blocking feature for service-type actor.
 - manually-accept
	Enable manually accept follow request.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(configDisable, cmd, args)
		},
	}
	config.AddCommand(configDisable)

	return config
}

func editConfig(key string, value bool) string {
	var statement string
	if value {
		statement = "Enabled"
	} else {
		statement = "Disabled"
	}
	switch key {
	case "person-only":
		RelayState.SetConfig(PersonOnly, value)
		return "Limited for Person-Type Actor is " + statement + "."
	case "manually-accept":
		RelayState.SetConfig(ManuallyAccept, value)
		return "Manually Accept Follow-Request is " + statement + "."
	}
	return "Invalid Config Provided."
}

func configEnable(cmd *cobra.Command, args []string) error {
	for _, config := range args {
		message := editConfig(config, true)
		cmd.Println(message)
	}
	return nil
}

func configDisable(cmd *cobra.Command, args []string) error {
	for _, config := range args {
		message := editConfig(config, false)
		cmd.Println(message)
	}
	return nil
}

func listConfig(cmd *cobra.Command, _ []string) {
	cmd.Println("Limited for Person-Type Actor : ", RelayState.RelayConfig.PersonOnly)
	cmd.Println("Manually Accept Follow-Request : ", RelayState.RelayConfig.ManuallyAccept)
}

func exportConfig(cmd *cobra.Command, _ []string) {
	jsonData, _ := json.Marshal(&RelayState)
	cmd.Println(string(jsonData))
}

func importConfig(cmd *cobra.Command, _ []string) {
	jsonData := cmd.Flag("data").Value.String()
	var data models.RelayState
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		logrus.Error(err)
		return
	}

	if data.RelayConfig.PersonOnly {
		RelayState.SetConfig(PersonOnly, true)
		cmd.Println("Blocking for service-type actor is Enabled.")
	}
	if data.RelayConfig.ManuallyAccept {
		RelayState.SetConfig(ManuallyAccept, true)
		cmd.Println("Manually accept follow-request is Enabled.")
	}
	for _, LimitedDomain := range data.LimitedDomains {
		RelayState.SetLimitedDomain(LimitedDomain, true)
		cmd.Println("Set [" + LimitedDomain + "] as limited domain")
	}
	for _, BlockedDomain := range data.BlockedDomains {
		RelayState.SetBlockedDomain(BlockedDomain, true)
		cmd.Println("Set [" + BlockedDomain + "] as blocked domain")
	}
	for _, Subscription := range data.Subscribers {
		RelayState.AddSubscriber(models.Subscriber{
			Domain:     Subscription.Domain,
			InboxURL:   Subscription.InboxURL,
			ActivityID: Subscription.ActivityID,
			ActorID:    Subscription.ActorID,
		})
		cmd.Println("Register [" + Subscription.Domain + "] as subscriber")
	}
}
