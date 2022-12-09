package control

import (
	"encoding/json"
	"fmt"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yukimochi/Activity-Relay/models"
	"github.com/yukimochi/machinery-v1/v1/tasks"
)

func followCmdInit() *cobra.Command {
	var follow = &cobra.Command{
		Use:   "follow",
		Short: "Manage follow request for relay",
		Long:  "List all follow request and accept/reject follow requests.",
	}

	var followList = &cobra.Command{
		Use:   "list",
		Short: "List follow request",
		Long:  "List follow request.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(listFollows, cmd, args)
		},
	}
	follow.AddCommand(followList)

	var followAccept = &cobra.Command{
		Use:   "accept",
		Short: "Accept follow request",
		Long:  "Accept follow request by domain.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(acceptFollow, cmd, args)
		},
	}
	follow.AddCommand(followAccept)

	var followReject = &cobra.Command{
		Use:   "reject",
		Short: "Reject follow request",
		Long:  "Reject follow request by domain.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(rejectFollow, cmd, args)
		},
	}
	follow.AddCommand(followReject)

	var updateActor = &cobra.Command{
		Use:   "update",
		Short: "Update actor object",
		Long:  "Update actor object for whole subscribers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return InitProxyE(updateActor, cmd, args)
		},
	}
	follow.AddCommand(updateActor)

	return follow
}

func pushRegisterJob(inboxURL string, body []byte) {
	job := &tasks.Signature{
		Name:       "register",
		RetryCount: 25,
		Args: []tasks.Arg{
			{
				Name:  "inboxURL",
				Type:  "string",
				Value: inboxURL,
			},
			{
				Name:  "body",
				Type:  "string",
				Value: string(body),
			},
		},
	}
	_, err := MachineryServer.SendTask(job)
	if err != nil {
		logrus.Error(err)
	}
}

func createFollowRequestResponse(domain string, response string) error {
	data, err := RelayState.RedisClient.HGetAll("relay:pending:" + domain).Result()
	if err != nil {
		return err
	}
	activity := models.Activity{
		Context: []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		ID:      data["activity_id"],
		Actor:   data["actor"],
		Type:    data["type"],
		Object:  data["object"],
	}

	resp := activity.GenerateResponse(GlobalConfig.ServerHostname(), response)
	jsonData, err := json.Marshal(&resp)
	if err != nil {
		return err
	}
	pushRegisterJob(data["inbox_url"], jsonData)
	RelayState.RedisClient.Del("relay:pending:" + domain)
	if response == "Accept" {
		RelayState.AddSubscription(models.Subscription{
			Domain:     domain,
			InboxURL:   data["inbox_url"],
			ActivityID: data["activity_id"],
			ActorID:    data["actor"],
		})
	}

	return nil
}

func createUpdateActorActivity(subscription models.Subscription) error {
	activity := models.Activity{
		Context: []string{"https://www.w3.org/ns/activitystreams"},
		ID:      GlobalConfig.ServerHostname().String() + "/activities/" + uuid.NewV4().String(),
		Actor:   GlobalConfig.ServerHostname().String() + "/actor",
		Type:    "Update",
		To:      []string{"https://www.w3.org/ns/activitystreams#Public"},
		Object:  Actor,
	}

	jsonData, err := json.Marshal(&activity)
	if err != nil {
		return err
	}
	pushRegisterJob(subscription.InboxURL, jsonData)

	return nil
}

func listFollows(cmd *cobra.Command, _ []string) error {
	var domains []string
	cmd.Println(" - Follow request :")
	follows, err := RelayState.RedisClient.Keys("relay:pending:*").Result()
	if err != nil {
		return err
	}
	for _, follow := range follows {
		domains = append(domains, strings.Replace(follow, "relay:pending:", "", 1))
	}
	for _, domain := range domains {
		cmd.Println(domain)
	}
	cmd.Println(fmt.Sprintf("Total : %d", len(domains)))

	return nil
}

func acceptFollow(cmd *cobra.Command, args []string) error {
	var err error
	var domains []string
	follows, err := RelayState.RedisClient.Keys("relay:pending:*").Result()
	if err != nil {
		return err
	}
	for _, follow := range follows {
		domains = append(domains, strings.Replace(follow, "relay:pending:", "", 1))
	}

	for _, domain := range args {
		if contains(domains, domain) {
			cmd.Println("Accept [" + domain + "] follow request")
			createFollowRequestResponse(domain, "Accept")
		} else {
			cmd.Println("Invalid domain [" + domain + "] given")
		}
	}

	return nil
}

func rejectFollow(cmd *cobra.Command, args []string) error {
	var err error
	var domains []string
	follows, err := RelayState.RedisClient.Keys("relay:pending:*").Result()
	if err != nil {
		return err
	}
	for _, follow := range follows {
		domains = append(domains, strings.Replace(follow, "relay:pending:", "", 1))
	}

	for _, domain := range args {
		if contains(domains, domain) {
			cmd.Println("Reject [" + domain + "] follow request")
			createFollowRequestResponse(domain, "Reject")
		} else {
			cmd.Println("Invalid domain [" + domain + "] given")
		}
	}

	return nil
}

func updateActor(cmd *cobra.Command, _ []string) error {
	for _, subscription := range RelayState.Subscriptions {
		err := createUpdateActorActivity(subscription)
		if err != nil {
			cmd.Println("Failed Update Actor for " + subscription.Domain)
		}
	}
	return nil
}
