package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

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
		exportConfig.SetConfig(BlockService, false)
		fmt.Println("Blocking for service-type actor is Disabled.")
	} else {
		exportConfig.SetConfig(BlockService, true)
		fmt.Println("Blocking for service-type actor is Enabled.")
	}
}

func manuallyAccept(c *cli.Context) {
	if c.Bool("undo") {
		exportConfig.SetConfig(ManuallyAccept, false)
		fmt.Println("Manually accept follow-request is Disabled.")
	} else {
		exportConfig.SetConfig(ManuallyAccept, true)
		fmt.Println("Manually accept follow-request is Enabled.")
	}
}

func createAsAnnounce(c *cli.Context) {
	if c.Bool("undo") {
		exportConfig.SetConfig(CreateAsAnnounce, false)
		fmt.Println("Announce activity instead of relay create activity is Disabled.")
	} else {
		exportConfig.SetConfig(CreateAsAnnounce, true)
		fmt.Println("Announce activity instead of relay create activity is Enabled.")
	}
}

func listConfigs(c *cli.Context) {
	fmt.Println("Blocking for service-type actor : ", exportConfig.RelayConfig.BlockService)
	fmt.Println("Manually accept follow-request : ", exportConfig.RelayConfig.ManuallyAccept)
	fmt.Println("Announce activity instead of relay create activity : ", exportConfig.RelayConfig.CreateAsAnnounce)
}

func exportConfigs(c *cli.Context) {
	jsonData, _ := json.Marshal(&exportConfig)
	fmt.Println(string(jsonData))
}

func importConfigs(c *cli.Context) {
	file, err := os.Open(c.String("json"))
	if err != nil {
		fmt.Println(err)
		return
	}
	jsonData, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		return
	}
	var data relayconf.ExportConfig
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		fmt.Println(err)
		return
	}

	if data.RelayConfig.BlockService {
		exportConfig.SetConfig(BlockService, true)
	}
	if data.RelayConfig.ManuallyAccept {
		exportConfig.SetConfig(ManuallyAccept, true)
	}
	if data.RelayConfig.CreateAsAnnounce {
		exportConfig.SetConfig(CreateAsAnnounce, true)
	}
	for _, LimitedDomain := range data.LimitedDomains {
		exportConfig.SetLimitedDomain(LimitedDomain, true)
		redClient.HSet("relay:config:limitedDomain", LimitedDomain, "1")
	}
	for _, BlockedDomain := range data.BlockedDomains {
		exportConfig.SetLimitedDomain(BlockedDomain, true)
		redClient.HSet("relay:config:blockedDomain", BlockedDomain, "1")
	}
	for _, Subscription := range data.Subscriptions {
		exportConfig.AddSubscription(relayconf.Subscription{
			Domain:     Subscription.Domain,
			InboxURL:   Subscription.InboxURL,
			ActivityID: Subscription.ActivityID,
			ActorID:    Subscription.ActorID,
		})
	}
}
