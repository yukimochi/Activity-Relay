package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/urfave/cli"
	"github.com/yukimochi/Activity-Relay/State"
)

const (
	BlockService state.Config = iota
	ManuallyAccept
	CreateAsAnnounce
)

func serviceBlock(c *cli.Context) {
	if c.Bool("undo") {
		relayState.SetConfig(BlockService, false)
		fmt.Println("Blocking for service-type actor is Disabled.")
	} else {
		relayState.SetConfig(BlockService, true)
		fmt.Println("Blocking for service-type actor is Enabled.")
	}
}

func manuallyAccept(c *cli.Context) {
	if c.Bool("undo") {
		relayState.SetConfig(ManuallyAccept, false)
		fmt.Println("Manually accept follow-request is Disabled.")
	} else {
		relayState.SetConfig(ManuallyAccept, true)
		fmt.Println("Manually accept follow-request is Enabled.")
	}
}

func createAsAnnounce(c *cli.Context) {
	if c.Bool("undo") {
		relayState.SetConfig(CreateAsAnnounce, false)
		fmt.Println("Announce activity instead of relay create activity is Disabled.")
	} else {
		relayState.SetConfig(CreateAsAnnounce, true)
		fmt.Println("Announce activity instead of relay create activity is Enabled.")
	}
}

func listConfigs(c *cli.Context) {
	fmt.Println("Blocking for service-type actor : ", relayState.RelayConfig.BlockService)
	fmt.Println("Manually accept follow-request : ", relayState.RelayConfig.ManuallyAccept)
	fmt.Println("Announce activity instead of relay create activity : ", relayState.RelayConfig.CreateAsAnnounce)
}

func exportConfigs(c *cli.Context) {
	jsonData, _ := json.Marshal(&relayState)
	fmt.Println(string(jsonData))
}

func importConfigs(c *cli.Context) {
	file, err := os.Open(c.String("json"))
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
	}
	if data.RelayConfig.ManuallyAccept {
		relayState.SetConfig(ManuallyAccept, true)
	}
	if data.RelayConfig.CreateAsAnnounce {
		relayState.SetConfig(CreateAsAnnounce, true)
	}
	for _, LimitedDomain := range data.LimitedDomains {
		relayState.SetLimitedDomain(LimitedDomain, true)
		redClient.HSet("relay:config:limitedDomain", LimitedDomain, "1")
	}
	for _, BlockedDomain := range data.BlockedDomains {
		relayState.SetLimitedDomain(BlockedDomain, true)
		redClient.HSet("relay:config:blockedDomain", BlockedDomain, "1")
	}
	for _, Subscription := range data.Subscriptions {
		relayState.AddSubscription(state.Subscription{
			Domain:     Subscription.Domain,
			InboxURL:   Subscription.InboxURL,
			ActivityID: Subscription.ActivityID,
			ActorID:    Subscription.ActorID,
		})
	}
}
