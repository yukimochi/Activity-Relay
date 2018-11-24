package main

import (
	"fmt"

	"github.com/urfave/cli"
	"github.com/yukimochi/Activity-Relay/RelayConf"
)

func serviceBlock(c *cli.Context) {
	if c.Bool("undo") {
		relayconf.SetConfig(redClient, "block_service", false)
		fmt.Println("Blocking for service-type actor is Disabled.")
	} else {
		relayconf.SetConfig(redClient, "block_service", true)
		fmt.Println("Blocking for service-type actor is Enabled.")
	}
}

func manuallyAccept(c *cli.Context) {
	if c.Bool("undo") {
		relayconf.SetConfig(redClient, "manually_accept", false)
		fmt.Println("Manually accept follow-request is Disabled.")
	} else {
		relayconf.SetConfig(redClient, "manually_accept", true)
		fmt.Println("Manually accept follow-request is Enabled.")
	}
}

func createAsAnnounce(c *cli.Context) {
	if c.Bool("undo") {
		relayconf.SetConfig(redClient, "create_as_announce", false)
		fmt.Println("Announce activity instead of relay create activity is Disabled.")
	} else {
		relayconf.SetConfig(redClient, "create_as_announce", true)
		fmt.Println("Announce activity instead of relay create activity is Enabled.")
	}
}

func listConfigs(c *cli.Context) {
	config := relayconf.LoadConfig(redClient)

	fmt.Println("Blocking for service-type actor : ", config.BlockService)
	fmt.Println("Manually accept follow-request : ", config.ManuallyAccept)
	fmt.Println("Announce activity instead of relay create activity : ", config.CreateAsAnnounce)
}
