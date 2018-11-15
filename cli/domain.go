package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli"
)

func listDomains(c *cli.Context) error {
	var err error
	var domains []string
	var message string
	switch c.String("type") {
	case "limited":
		fmt.Println(" - Limited domain :")
		domains, err = redClient.HKeys("relay:config:limitedDomain").Result()
		if err != nil {
			return err
		}
	case "blocked":
		fmt.Println(" - Blocked domain :")
		domains, err = redClient.HKeys("relay:config:blockedDomain").Result()
		if err != nil {
			return err
		}
	default:
		fmt.Println(" - Subscribed domain :")
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

func setDomainType(c *cli.Context) error {
	if c.String("domain") == "" {
		fmt.Println("No domain given.")
		return nil
	}
	switch c.String("type") {
	case "limited":
		if c.Bool("undo") {
			redClient.HDel("relay:config:limitedDomain", c.String("domain"))
			fmt.Println("Unset [" + c.String("domain") + "] as Limited domain.")
		} else {
			redClient.HSet("relay:config:limitedDomain", c.String("domain"), "1")
			fmt.Println("Set [" + c.String("domain") + "] as Limited domain.")
		}
	case "blocked":
		if c.Bool("undo") {
			redClient.HDel("relay:config:blockedDomain", c.String("domain"))
			fmt.Println("Unset [" + c.String("domain") + "] as Blocked domain.")
		} else {
			redClient.HSet("relay:config:blockedDomain", c.String("domain"), "1")
			fmt.Println("Set [" + c.String("domain") + "] as Blocked domain.")
		}
	default:
		fmt.Println("No type given.")
	}
	return nil
}
