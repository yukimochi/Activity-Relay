package control

import (
	"encoding/json"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yukimochi/Activity-Relay/models"
)

const (
	BlockService models.Config = iota
	ManuallyAccept
	CreateAsAnnounce
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
		Long:  "Import all relay information from JSON file.",
		Run: func(cmd *cobra.Command, args []string) {
			InitProxy(importConfig, cmd, args)
		},
	}
	configImport.Flags().String("json", "", "JSON file-path")
	configImport.MarkFlagRequired("json")
	config.AddCommand(configImport)

	var configEnable = &cobra.Command{
		Use:   "enable",
		Short: "Enable/disable relay configuration",
		Long: `Enable or disable relay configuration.
 - service-block
	Blocking feature for service-type actor.
 - manually-accept
	Enable manually accept follow request.
 - create-as-announce
	Enable announce activity instead of relay create activity (not recommend)`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(configEnable, cmd, args)
		},
	}
	configEnable.Flags().BoolP("disable", "d", false, "Disable configuration instead of Enable")
	config.AddCommand(configEnable)

	return config
}

func configEnable(cmd *cobra.Command, args []string) error {
	disable := cmd.Flag("disable").Value.String() == "true"
	for _, config := range args {
		switch config {
		case "service-block":
			if disable {
				RelayState.SetConfig(BlockService, false)
				cmd.Println("Blocking for service-type actor is Disabled.")
			} else {
				RelayState.SetConfig(BlockService, true)
				cmd.Println("Blocking for service-type actor is Enabled.")
			}
		case "manually-accept":
			if disable {
				RelayState.SetConfig(ManuallyAccept, false)
				cmd.Println("Manually accept follow-request is Disabled.")
			} else {
				RelayState.SetConfig(ManuallyAccept, true)
				cmd.Println("Manually accept follow-request is Enabled.")
			}
		case "create-as-announce":
			if disable {
				RelayState.SetConfig(CreateAsAnnounce, false)
				cmd.Println("Announce activity instead of relay create activity is Disabled.")
			} else {
				RelayState.SetConfig(CreateAsAnnounce, true)
				cmd.Println("Announce activity instead of relay create activity is Enabled.")
			}
		default:
			cmd.Println("Invalid config given")
		}
	}

	return nil
}

func listConfig(cmd *cobra.Command, _ []string) {
	cmd.Println("Blocking for service-type actor : ", RelayState.RelayConfig.BlockService)
	cmd.Println("Manually accept follow-request : ", RelayState.RelayConfig.ManuallyAccept)
	cmd.Println("Announce activity instead of relay create activity : ", RelayState.RelayConfig.CreateAsAnnounce)
}

func exportConfig(cmd *cobra.Command, _ []string) {
	jsonData, _ := json.Marshal(&RelayState)
	cmd.Println(string(jsonData))
}

func importConfig(cmd *cobra.Command, _ []string) {
	file, err := os.Open(cmd.Flag("json").Value.String())
	if err != nil {
		logrus.Error(err)
		return
	}
	jsonData, err := io.ReadAll(file)
	if err != nil {
		logrus.Error(err)
		return
	}
	var data models.RelayState
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		logrus.Error(err)
		return
	}

	if data.RelayConfig.BlockService {
		RelayState.SetConfig(BlockService, true)
		cmd.Println("Blocking for service-type actor is Enabled.")
	}
	if data.RelayConfig.ManuallyAccept {
		RelayState.SetConfig(ManuallyAccept, true)
		cmd.Println("Manually accept follow-request is Enabled.")
	}
	if data.RelayConfig.CreateAsAnnounce {
		RelayState.SetConfig(CreateAsAnnounce, true)
		cmd.Println("Announce activity instead of relay create activity is Enabled.")
	}
	for _, LimitedDomain := range data.LimitedDomains {
		RelayState.SetLimitedDomain(LimitedDomain, true)
		cmd.Println("Set [" + LimitedDomain + "] as limited domain")
	}
	for _, BlockedDomain := range data.BlockedDomains {
		RelayState.SetBlockedDomain(BlockedDomain, true)
		cmd.Println("Set [" + BlockedDomain + "] as blocked domain")
	}
	for _, Subscription := range data.Subscriptions {
		RelayState.AddSubscription(models.Subscription{
			Domain:     Subscription.Domain,
			InboxURL:   Subscription.InboxURL,
			ActivityID: Subscription.ActivityID,
			ActorID:    Subscription.ActorID,
		})
		cmd.Println("Register [" + Subscription.Domain + "] as subscriber")
	}
}
