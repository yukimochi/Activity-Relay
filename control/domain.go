package control

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yukimochi/Activity-Relay/models"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(listDomains, cmd, args)
		},
	}
	domainList.Flags().StringP("type", "t", "subscriber", "domain type [subscriber,limited,blocked]")
	domain.AddCommand(domainList)

	var domainSet = &cobra.Command{
		Use:   "set [flags]",
		Short: "Set or unset domain as limited or blocked",
		Long:  "Set or unset domain as limited or blocked.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(setDomainType, cmd, args)
		},
	}
	domainSet.Flags().StringP("type", "t", "", "Apply domain type [limited,blocked]")
	domainSet.MarkFlagRequired("type")
	domainSet.Flags().BoolP("undo", "u", false, "Unset domain as limited or blocked")
	domain.AddCommand(domainSet)

	var domainUnfollow = &cobra.Command{
		Use:   "unfollow [flags]",
		Short: "Send Unfollow request for given domains",
		Long:  "Send unfollow request for given domains.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(unfollowDomains, cmd, args)
		},
	}
	domain.AddCommand(domainUnfollow)

	return domain
}

func createUnfollowRequestResponse(subscription models.Subscription) error {
	activity := models.Activity{
		Context: []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		ID:      subscription.ActivityID,
		Actor:   subscription.ActorID,
		Type:    "Follow",
		Object:  "https://www.w3.org/ns/activitystreams#Public",
	}

	resp := activity.GenerateResponse(GlobalConfig.ServerHostname(), "Reject")
	jsonData, _ := json.Marshal(&resp)
	pushRegisterJob(subscription.InboxURL, jsonData)

	return nil
}

func listDomains(cmd *cobra.Command, _ []string) error {
	var domains []string
	switch cmd.Flag("type").Value.String() {
	case "limited":
		cmd.Println(" - Limited domain :")
		domains = RelayState.LimitedDomains
	case "blocked":
		cmd.Println(" - Blocked domain :")
		domains = RelayState.BlockedDomains
	default:
		cmd.Println(" - Subscriber domain :")
		temp := RelayState.Subscriptions
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
			RelayState.SetLimitedDomain(domain, !undo)
			if undo {
				cmd.Println("Unset [" + domain + "] as limited domain")
			} else {
				cmd.Println("Set [" + domain + "] as limited domain")
			}
		}
	case "blocked":
		for _, domain := range args {
			RelayState.SetBlockedDomain(domain, !undo)
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
	subscriptions := RelayState.Subscriptions
	for _, domain := range args {
		if contains(subscriptions, domain) {
			subscription := *RelayState.SelectSubscription(domain)
			createUnfollowRequestResponse(subscription)
			RelayState.DelSubscription(subscription.Domain)
			cmd.Println("Unfollow [" + subscription.Domain + "]")
		} else {
			cmd.Println("Invalid domain [" + domain + "] given")
		}
	}

	return nil
}
