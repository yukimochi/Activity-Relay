package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-redis/redis"
	"github.com/urfave/cli"
)

var redClient *redis.Client

func listDomain(c *cli.Context) error {
	var err error
	var domains []string
	var message string
	switch c.String("type") {
	case "limited":
		message = " - Limited domain :"
		domains, err = redClient.HKeys("relay:config:limitedDomain").Result()
		if err != nil {
			return err
		}
	case "blocked":
		message = " - Blocked domain :"
		domains, err = redClient.HKeys("relay:config:blockedDomain").Result()
		if err != nil {
			return err
		}
	default:
		message = " - Subscribed domain :"
		temp, err := redClient.Keys("relay:subscription:*").Result()
		if err != nil {
			return err
		}
		for _, domain := range temp {
			domains = append(domains, strings.Replace(domain, "relay:subscription:", "", 1))
		}
	}
	fmt.Println(message)
	for _, domain := range domains {
		fmt.Println(domain)
	}
	fmt.Println(fmt.Sprintf("Total : %d", len(domains)))
	return nil
}

func manageDomain(c *cli.Context) error {
	if c.String("domain") == "" {
		fmt.Println("No domain given.")
		return nil
	}
	switch c.String("type") {
	case "limited":
		if c.Bool("undo") {
			redClient.HDel("relay:config:limitedDomain", c.String("domain"))
			fmt.Println("Unregistrate [" + c.String("domain") + "] from Limited domain.")
		} else {
			redClient.HSet("relay:config:limitedDomain", c.String("domain"), "1")
			fmt.Println("Registrate [" + c.String("domain") + "] as Limited domain.")
		}
	case "blocked":
		if c.Bool("undo") {
			redClient.HDel("relay:config:blockedDomain", c.String("domain"))
			fmt.Println("Unregistrate [" + c.String("domain") + "] from Blocked domain.")
		} else {
			redClient.HSet("relay:config:blockedDomain", c.String("domain"), "1")
			fmt.Println("Registrate [" + c.String("domain") + "] as Blocked domain.")
		}
	default:
		fmt.Println("No type given.")
	}
	return nil
}

func main() {
	redClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})

	app := cli.NewApp()
	app.Name = "Activity Relay Extarnal CLI"
	app.Usage = "Control Relay configration"
	app.Version = "0.0.2"
	app.Commands = []cli.Command{
		{
			Name:    "list-domain",
			Aliases: []string{"ld"},
			Usage:   "List {subscribed,limited,blocked} domains",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type, t",
					Value: "subscribed",
					Usage: "Registrate type [subscribed,limited,blocked]",
				},
			},
			Action: listDomain,
		},
		{
			Name:    "manage-domain",
			Aliases: []string{"md"},
			Usage:   "Manage {limited,blocked} domains",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "type, t",
					Usage: "Registrate type [limited,blocked]",
				},
				cli.StringFlag{
					Name:  "domain, d",
					Usage: "Registrate domain",
				},
				cli.BoolFlag{
					Name:  "undo, u",
					Usage: "Undo registrate",
				},
			},
			Action: manageDomain,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
