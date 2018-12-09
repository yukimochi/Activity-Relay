package main

import (
	"fmt"

	"github.com/urfave/cli"
)

func listDomains(c *cli.Context) error {
	var domains []string
	switch c.String("type") {
	case "limited":
		fmt.Println(" - Limited domain :")
		domains = exportConfig.LimitedDomains
	case "blocked":
		fmt.Println(" - Blocked domain :")
		domains = exportConfig.BlockedDomains
	default:
		fmt.Println(" - Subscribed domain :")
		temp := exportConfig.Subscriptions
		for _, domain := range temp {
			domains = append(domains, domain.Domain)
		}
	}
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
			exportConfig.SetLimitedDomain(c.String("domain"), false)
			fmt.Println("Unset [" + c.String("domain") + "] as Limited domain.")
		} else {
			exportConfig.SetLimitedDomain(c.String("domain"), true)
			fmt.Println("Set [" + c.String("domain") + "] as Limited domain.")
		}
	case "blocked":
		if c.Bool("undo") {
			exportConfig.SetBlockedDomain(c.String("domain"), false)
			fmt.Println("Unset [" + c.String("domain") + "] as Blocked domain.")
		} else {
			exportConfig.SetBlockedDomain(c.String("domain"), true)
			fmt.Println("Set [" + c.String("domain") + "] as Blocked domain.")
		}
	default:
		fmt.Println("No type given.")
	}
	return nil
}
