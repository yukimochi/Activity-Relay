package main

import (
	"fmt"

	"github.com/spf13/cobra"
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

	return domain
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
