package main

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli"
	"github.com/yukimochi/Activity-Relay/RelayConf"
)

const (
	BlockService relayconf.Config = iota
	ManuallyAccept
	CreateAsAnnounce
)

func serviceBlock(c *cli.Context) {
	if c.Bool("undo") {
		relConfig.Set(redClient, BlockService, false)
		fmt.Println("Blocking for service-type actor is Disabled.")
	} else {
		relConfig.Set(redClient, BlockService, true)
		fmt.Println("Blocking for service-type actor is Enabled.")
	}
}

func manuallyAccept(c *cli.Context) {
	if c.Bool("undo") {
		relConfig.Set(redClient, ManuallyAccept, false)
		fmt.Println("Manually accept follow-request is Disabled.")
	} else {
		relConfig.Set(redClient, ManuallyAccept, true)
		fmt.Println("Manually accept follow-request is Enabled.")
	}
}

func createAsAnnounce(c *cli.Context) {
	if c.Bool("undo") {
		relConfig.Set(redClient, CreateAsAnnounce, false)
		fmt.Println("Announce activity instead of relay create activity is Disabled.")
	} else {
		relConfig.Set(redClient, CreateAsAnnounce, true)
		fmt.Println("Announce activity instead of relay create activity is Enabled.")
	}
}

func listConfigs(c *cli.Context) {
	relConfig.Load(redClient)

	fmt.Println("Blocking for service-type actor : ", relConfig.BlockService)
	fmt.Println("Manually accept follow-request : ", relConfig.ManuallyAccept)
	fmt.Println("Announce activity instead of relay create activity : ", relConfig.CreateAsAnnounce)
}

func exportConfigs(c *cli.Context) {
	var ex relayconf.ExportConfig
	ex.Import(redClient)

	jsonData, _ := json.Marshal(&ex)
	fmt.Println(string(jsonData))
}
