package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/url"
	"os"

	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/go-redis/redis"
	"github.com/urfave/cli"
	"github.com/yukimochi/Activity-Relay/KeyLoader"
	"github.com/yukimochi/Activity-Relay/RelayConf"
)

var hostname *url.URL
var hostkey *rsa.PrivateKey
var redClient *redis.Client
var macServer *machinery.Server
var relConfig relayconf.RelayConfig

func main() {
	pemPath := os.Getenv("ACTOR_PEM")
	if pemPath == "" {
		panic("Require ACTOR_PEM environment variable.")
	}
	relayDomain := os.Getenv("RELAY_DOMAIN")
	if relayDomain == "" {
		panic("Require RELAY_DOMAIN environment variable.")
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "127.0.0.1:6379"
	}

	var err error
	hostkey, err = keyloader.ReadPrivateKeyRSAfromPath(pemPath)
	if err != nil {
		panic("Can't read Hostkey Pemfile")
	}
	hostname, err = url.Parse("https://" + relayDomain)
	if err != nil {
		panic("Can't parse Relay Domain")
	}
	redClient = redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	var macConfig = &config.Config{
		Broker:          "redis://" + redisURL,
		DefaultQueue:    "relay",
		ResultBackend:   "redis://" + redisURL,
		ResultsExpireIn: 5,
	}

	macServer, err = machinery.NewServer(macConfig)
	if err != nil {
		fmt.Println(err)
	}

	app := cli.NewApp()
	app.Name = "Activity Relay Extarnal CLI"
	app.Usage = "Control Relay configration"
	app.Version = "0.1.1"
	app.Commands = []cli.Command{
		{
			Name:  "domain",
			Usage: "Management domains",
			Subcommands: []cli.Command{
				{
					Name:  "list",
					Usage: "List {subscribed,limited,blocked} domains",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "type, t",
							Value: "subscribed",
							Usage: "Domain type [subscribed,limited,blocked]",
						},
					},
					Action: listDomains,
				},
				{
					Name:  "set",
					Usage: "set domain type [limited,blocked]",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "type, t",
							Usage: "Domain type [limited,blocked]",
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
					Action: setDomainType,
				},
			},
		},
		{
			Name:  "config",
			Usage: "Management relay config",
			Subcommands: []cli.Command{
				{
					Name:   "show",
					Usage:  "Show all relay configrations",
					Action: listConfigs,
				},
				{
					Name:   "export",
					Usage:  "Export all relay information (json)",
					Action: exportConfigs,
				},
				{
					Name:  "service-block",
					Usage: "Enable blocking for service-type actor",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "undo, u",
							Usage: "Undo block",
						},
					},
					Action: serviceBlock,
				},
				{
					Name:  "manually-accept",
					Usage: "Enable Manually accept follow-request",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "undo, u",
							Usage: "Undo block",
						},
					},
					Action: manuallyAccept,
				},
				{
					Name:  "create-as-announce",
					Usage: "Enable Announce activity instead of relay create activity (Not recommended)",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "undo, u",
							Usage: "Undo block",
						},
					},
					Action: createAsAnnounce,
				},
			},
		},
		{
			Name:  "follow-request",
			Usage: "Management follow-request",
			Subcommands: []cli.Command{
				{
					Name:   "show",
					Usage:  "Show all follow-request",
					Action: listFollows,
				},
				{
					Name:  "reject",
					Usage: "Reject follow-request",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "domain, d",
							Usage: "domain name",
						},
					},
					Action: rejectFollow,
				},
				{
					Name:  "accept",
					Usage: "Accept follow-request",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "domain, d",
							Usage: "domain name",
						},
					},
					Action: acceptFollow,
				},
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
