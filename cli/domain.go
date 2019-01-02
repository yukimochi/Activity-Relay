package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	activitypub "github.com/yukimochi/Activity-Relay/ActivityPub"
	state "github.com/yukimochi/Activity-Relay/State"
)

func domainCmdInit() *cobra.Command {
	var domain = &cobra.Command{
		Use:   "domain",
		Short: "Manage subscriber domain",
		Long:  "List all subscriber, set/unset domain as limited or blocked and undo subscribe domain.",
	}

	var domainList = &cobra.Command{
		Use:   "list [flags]",
		Short: "List domain",
		Long:  "List domain which filtered given type.",
		RunE:  listDomains,
	}
	domainList.Flags().StringP("type", "t", "subscriber", "domain type [subscriber,limited,blocked]")
	domain.AddCommand(domainList)

	var domainSet = &cobra.Command{
		Use:   "set [flags]",
		Short: "Set or unset domain as limited or blocked",
		Long:  "Set or unset domain as limited or blocked.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  setDomainType,
	}
	domainSet.Flags().StringP("type", "t", "", "Apply domain type [limited,blocked]")
	domainSet.MarkFlagRequired("type")
	domainSet.Flags().BoolP("undo", "u", false, "Unset domain as limited or blocked")
	domain.AddCommand(domainSet)

	var domainUnfollow = &cobra.Command{
		Use:   "unfollow [flags]",
		Short: "Send Unfollow request for given domains",
		Long:  "Send unfollow request for given domains.",
		RunE:  unfollowDomains,
	}
	domain.AddCommand(domainUnfollow)

	return domain
}

func createUnfollowRequestResponse(subscription state.Subscription) error {
	activity := activitypub.Activity{
		Context: []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		ID:      subscription.ActivityID,
		Actor:   subscription.ActorID,
		Type:    "Follow",
		Object:  "https://www.w3.org/ns/activitystreams#Public",
	}

	resp := activity.GenerateResponse(hostname, "Reject")
	jsonData, _ := json.Marshal(&resp)
	pushRegistorJob(subscription.InboxURL, jsonData)

	return nil
}

func listDomains(cmd *cobra.Command, args []string) error {
	var domains []string
	switch cmd.Flag("type").Value.String() {
	case "limited":
		cmd.Println(" - Limited domain :")
		domains = relayState.LimitedDomains
	case "blocked":
		cmd.Println(" - Blocked domain :")
		domains = relayState.BlockedDomains
	default:
		cmd.Println(" - Subscriber domain :")
		temp := relayState.Subscriptions
		for _, domain := range temp {
			domains = append(domains, domain.Domain)
		}
	}
	for _, domain := range domains {
		cmd.Println(domain)
	}
	cmd.Println(fmt.Sprintf("Total : %d", len(domains)))

	return nil
}

func setDomainType(cmd *cobra.Command, args []string) error {
	undo := cmd.Flag("undo").Value.String() == "true"
	switch cmd.Flag("type").Value.String() {
	case "limited":
		for _, domain := range args {
			relayState.SetLimitedDomain(domain, !undo)
			if undo {
				cmd.Println("Unset [" + domain + "] as limited domain")
			} else {
				cmd.Println("Set [" + domain + "] as limited domain")
			}
		}
	case "blocked":
		for _, domain := range args {
			relayState.SetBlockedDomain(domain, !undo)
			if undo {
				cmd.Println("Unset [" + domain + "] as blocked domain")
			} else {
				cmd.Println("Set [" + domain + "] as blocked domain")
			}
		}
	default:
		cmd.Println("Invalid type given")
	}

	return nil
}

func unfollowDomains(cmd *cobra.Command, args []string) error {
	subscriptions := relayState.Subscriptions
	for _, domain := range args {
		for _, subscription := range subscriptions {
			if domain == subscription.Domain {
				cmd.Println("Unfollow [" + domain + "]")
				createUnfollowRequestResponse(subscription)
				relayState.DelSubscription(subscription.Domain)
				break
			}
			fmt.Println("Invalid domain given")
		}
	}

	return nil
}
