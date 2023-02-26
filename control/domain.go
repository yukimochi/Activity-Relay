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
		Long:  "List domain which filtered provided type.",
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
		Short: "Send Unfollow request for provided domains",
		Long:  "Send unfollow request for provided domains.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(unfollowDomains, cmd, args)
		},
	}
	domain.AddCommand(domainUnfollow)

	return domain
}

func createUnfollowToSubscriberRequest(subscriber models.Subscriber) error {
	activity := models.Activity{
		Context: []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		ID:      subscriber.ActivityID,
		Actor:   subscriber.ActorID,
		Type:    "Follow",
		Object:  "https://www.w3.org/ns/activitystreams#Public",
	}

	resp := activity.GenerateReply(RelayActor, activity, "Reject")
	jsonData, _ := json.Marshal(&resp)
	enqueueRegisterActivity(subscriber.InboxURL, jsonData)

	return nil
}

func createUnfollowToFollowerRequest(follower models.Follower) error {
	activity := models.Activity{
		Context: []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		ID:      follower.ActivityID,
		Actor:   follower.ActorID,
		Type:    "Follow",
		Object:  RelayActor.ID,
	}

	resp := activity.GenerateReply(RelayActor, activity, "Reject")
	jsonData, _ := json.Marshal(&resp)
	enqueueRegisterActivity(follower.InboxURL, jsonData)

	return nil
}

func listDomains(cmd *cobra.Command, _ []string) error {
	var count int
	switch cmd.Flag("type").Value.String() {
	case "limited":
		cmd.Println(" - Limited domains :")
		for _, domain := range RelayState.LimitedDomains {
			count = count + 1
			cmd.Println(domain)
		}
	case "blocked":
		cmd.Println(" - Blocked domains :")
		for _, domain := range RelayState.BlockedDomains {
			count = count + 1
			cmd.Println(domain)
		}
	default:
		cmd.Println(" - Subscriber list :")
		subscribers := RelayState.Subscribers
		for _, subscriber := range subscribers {
			count = count + 1
			cmd.Println("[*] " + subscriber.Domain)
		}
		cmd.Println(" - Follower list :")
		followers := RelayState.Followers
		for _, follower := range followers {
			count = count + 1
			if follower.MutuallyFollow {
				cmd.Println("[*] " + follower.Domain)
			} else {
				cmd.Println("[-] " + follower.Domain)
			}
		}
	}
	cmd.Println(fmt.Sprintf("Total : %d", count))

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
		cmd.Println("Invalid type provided")
	}

	return nil
}

func unfollowDomains(cmd *cobra.Command, args []string) error {
	subscriptions := RelayState.Subscribers
	followers := RelayState.Followers
	for _, domain := range args {
		switch {
		case contains(subscriptions, domain):
			subscription := *RelayState.SelectSubscriber(domain)
			createUnfollowToSubscriberRequest(subscription)
			RelayState.DelSubscriber(subscription.Domain)
			cmd.Println("Unfollow [" + subscription.Domain + "]")
		case contains(followers, domain):
			follower := *RelayState.SelectFollower(domain)
			createUnfollowToFollowerRequest(follower)
			RelayState.DelFollower(follower.Domain)
			cmd.Println("Unfollow [" + follower.Domain + "]")
		default:
			cmd.Println("Invalid domain [" + domain + "] provided")
		}
	}
	return nil
}
