package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/spf13/cobra"
	"github.com/yukimochi/Activity-Relay/ActivityPub"
	"github.com/yukimochi/Activity-Relay/State"
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
		RunE:  listFollows,
	}
	follow.AddCommand(followList)

	var followAccept = &cobra.Command{
		Use:   "accept",
		Short: "Accept follow request",
		Long:  "Accept follow request by domain.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  acceptFollow,
	}
	follow.AddCommand(followAccept)

	var followReject = &cobra.Command{
		Use:   "reject",
		Short: "Reject follow request",
		Long:  "Reject follow request by domain.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  rejectFollow,
	}
	follow.AddCommand(followReject)

	return follow
}

func pushRegistorJob(inboxURL string, body []byte) {
	job := &tasks.Signature{
		Name:       "registor",
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
	_, err := macServer.SendTask(job)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func createFollowRequestResponse(domain string, response string) error {
	data, err := relayState.RedisClient.HGetAll("relay:pending:" + domain).Result()
	if err != nil {
		return err
	}
	activity := activitypub.Activity{
		Context: []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		ID:      data["activity_id"],
		Actor:   data["actor"],
		Type:    data["type"],
		Object:  data["object"],
	}

	resp := activity.GenerateResponse(hostname, response)
	jsonData, _ := json.Marshal(&resp)
	pushRegistorJob(data["inbox_url"], jsonData)
	relayState.RedisClient.Del("relay:pending:" + domain)
	if response == "Accept" {
		relayState.AddSubscription(state.Subscription{
			Domain:     domain,
			InboxURL:   data["inbox_url"],
			ActivityID: data["activity_id"],
			ActorID:    data["actor"],
		})
	}

	return nil
}

func listFollows(cmd *cobra.Command, args []string) error {
	var domains []string
	cmd.Println(" - Follow request :")
	follows, err := relayState.RedisClient.Keys("relay:pending:*").Result()
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
	follows, err := relayState.RedisClient.Keys("relay:pending:*").Result()
	if err != nil {
		return err
	}
	for _, follow := range follows {
		domains = append(domains, strings.Replace(follow, "relay:pending:", "", 1))
	}

	for _, domain := range args {
		for _, request := range domains {
			if domain == request {
				cmd.Println("Accept [" + domain + "] follow request")
				createFollowRequestResponse(domain, "Accept")
				break
			}
		}
		cmd.Println("Invalid domain given")
	}

	return nil
}

func rejectFollow(cmd *cobra.Command, args []string) error {
	var err error
	var domains []string
	follows, err := relayState.RedisClient.Keys("relay:pending:*").Result()
	if err != nil {
		return err
	}
	for _, follow := range follows {
		domains = append(domains, strings.Replace(follow, "relay:pending:", "", 1))
	}

	for _, domain := range args {
		for _, request := range domains {
			if domain == request {
				cmd.Println("Reject [" + domain + "] follow request")
				createFollowRequestResponse(domain, "Reject")
				break
			}
		}
		cmd.Println("Invalid domain given")
	}

	return nil
}
