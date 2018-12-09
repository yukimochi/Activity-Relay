package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/RichardKnop/machinery/v1/tasks"
	"github.com/urfave/cli"
	"github.com/yukimochi/Activity-Relay/ActivityPub"
)

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

func listFollows(c *cli.Context) error {
	var err error
	var domains []string

	fmt.Println(" - Follow request :")
	follows, err := redClient.Keys("relay:pending:*").Result()
	if err != nil {
		return err
	}
	for _, follow := range follows {
		domains = append(domains, strings.Replace(follow, "relay:pending:", "", 1))
	}
	for _, domain := range domains {
		fmt.Println(domain)
	}
	return nil
}

func acceptFollow(c *cli.Context) error {
	domain := c.String("domain")
	if domain != "" {
		num, err := redClient.Exists("relay:pending:" + domain).Result()
		if err != nil {
			return err
		}
		if num == 0 {
			fmt.Println("Given domain not found.")
			return nil
		}

		fmt.Println("Accept Follow request : " + domain)
		data, err := redClient.HGetAll("relay:pending:" + domain).Result()
		if err != nil {
			return err
		}
		activity := activitypub.Activity{
			[]string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
			data["activity_id"],
			data["actor"],
			data["type"],
			data["object"],
			nil,
			nil,
		}

		resp := activity.GenerateResponse(hostname, "Accept")
		jsonData, _ := json.Marshal(&resp)

		pushRegistorJob(data["inbox_url"], jsonData)
		redClient.HSet("relay:subscription:"+domain, "inbox_url", data["inbox_url"])
		redClient.Del("relay:pending:" + domain)
		return nil
	} else {
		fmt.Println("No domain given.")
		return nil
	}
}

func rejectFollow(c *cli.Context) error {
	domain := c.String("domain")
	if domain != "" {
		num, err := redClient.Exists("relay:pending:" + domain).Result()
		if err != nil {
			return err
		}
		if num == 0 {
			fmt.Println("Given domain not found.")
			return nil
		}

		fmt.Println("Reject Follow request : " + domain)
		data, err := redClient.HGetAll("relay:pending:" + domain).Result()
		if err != nil {
			return err
		}
		activity := activitypub.Activity{
			[]string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
			data["activity_id"],
			data["actor"],
			data["type"],
			data["object"],
			nil,
			nil,
		}

		resp := activity.GenerateResponse(hostname, "Reject")
		jsonData, _ := json.Marshal(&resp)

		pushRegistorJob(data["inbox_url"], jsonData)
		redClient.Del("relay:pending:" + domain)
		return nil
	} else {
		fmt.Println("No domain given.")
		return nil
	}
}
