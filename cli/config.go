package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	state "github.com/yukimochi/Activity-Relay/State"
)

const (
	BlockService state.Config = iota
	ManuallyAccept
	CreateAsAnnounce
)

func configCmdInit() *cobra.Command {
	var config = &cobra.Command{
		Use:   "config",
		Short: "Manage configuration for relay",
		Long:  "Enable/disable relay costomize and import/export relay database.",
	}

	var configList = &cobra.Command{
		Use:   "list",
		Short: "List all relay configration",
		Long:  "List all relay configration.",
		Run:   listConfig,
	}
	config.AddCommand(configList)

	var configExport = &cobra.Command{
		Use:   "export",
		Short: "Export all relay information",
		Long:  "Export all relay information by JSON format.",
		Run:   exportConfig,
	}
	config.AddCommand(configExport)

	var configImport = &cobra.Command{
		Use:   "import [flags]",
		Short: "Import all relay information",
		Long:  "Import all relay information from JSON file.",
		Run:   importConfig,
	}
	configImport.Flags().String("json", "", "JSON file-path")
	configImport.MarkFlagRequired("json")
	config.AddCommand(configImport)

	var configEnable = &cobra.Command{
		Use:   "enable",
		Short: "Enable/disable relay configration",
		Long: `Enable or disable relay configration.
 - service-block
	Blocking feature for service-type actor.
 - manually-accept
	Enable manually accept follow request.
 - create-as-announce
	Enable announce activity instead of relay create activity (not recommend)`,
		Args: cobra.MinimumNArgs(1),
		RunE: configEnable,
	}
	configEnable.Flags().BoolP("disable", "d", false, "Disable configration instead of Enable")
	config.AddCommand(configEnable)

	return config
}

func configEnable(cmd *cobra.Command, args []string) error {
	disable := cmd.Flag("disable").Value.String() == "true"
	for _, config := range args {
		switch config {
		case "service-block":
			if disable {
				relayState.SetConfig(BlockService, false)
				cmd.Println("Blocking for service-type actor is Disabled.")
			} else {
				relayState.SetConfig(BlockService, true)
				cmd.Println("Blocking for service-type actor is Enabled.")
			}
		case "manually-accept":
			if disable {
				relayState.SetConfig(ManuallyAccept, false)
				cmd.Println("Manually accept follow-request is Disabled.")
			} else {
				relayState.SetConfig(ManuallyAccept, true)
				cmd.Println("Manually accept follow-request is Enabled.")
			}
		case "create-as-announce":
			if disable {
				relayState.SetConfig(CreateAsAnnounce, false)
				cmd.Println("Announce activity instead of relay create activity is Disabled.")
			} else {
				relayState.SetConfig(CreateAsAnnounce, true)
				cmd.Println("Announce activity instead of relay create activity is Enabled.")
			}
		default:
			cmd.Println("Invalid config given")
		}
	}

	return nil
}

func listConfig(cmd *cobra.Command, args []string) {
	cmd.Println("Blocking for service-type actor : ", relayState.RelayConfig.BlockService)
	cmd.Println("Manually accept follow-request : ", relayState.RelayConfig.ManuallyAccept)
	cmd.Println("Announce activity instead of relay create activity : ", relayState.RelayConfig.CreateAsAnnounce)
}

func exportConfig(cmd *cobra.Command, args []string) {
	jsonData, _ := json.Marshal(&relayState)
	cmd.Println(string(jsonData))
}

func importConfig(cmd *cobra.Command, args []string) {
	file, err := os.Open(cmd.Flag("json").Value.String())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	jsonData, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	var data state.RelayState
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if data.RelayConfig.BlockService {
		relayState.SetConfig(BlockService, true)
		cmd.Println("Blocking for service-type actor is Enabled.")
	}
	if data.RelayConfig.ManuallyAccept {
		relayState.SetConfig(ManuallyAccept, true)
		cmd.Println("Manually accept follow-request is Enabled.")
	}
	if data.RelayConfig.CreateAsAnnounce {
		relayState.SetConfig(CreateAsAnnounce, true)
		cmd.Println("Announce activity instead of relay create activity is Enabled.")
	}
	for _, LimitedDomain := range data.LimitedDomains {
		relayState.SetLimitedDomain(LimitedDomain, true)
		cmd.Println("Set [" + LimitedDomain + "] as limited domain")
	}
	for _, BlockedDomain := range data.BlockedDomains {
		relayState.SetBlockedDomain(BlockedDomain, true)
		cmd.Println("Set [" + BlockedDomain + "] as blocked domain")
	}
	for _, Subscription := range data.Subscriptions {
		relayState.AddSubscription(state.Subscription{
			Domain:     Subscription.Domain,
			InboxURL:   Subscription.InboxURL,
			ActivityID: Subscription.ActivityID,
			ActorID:    Subscription.ActorID,
		})
		cmd.Println("Regist [" + Subscription.Domain + "] as subscriber")
	}
}
