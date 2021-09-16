package api

import (
	"net/http"
	"time"

	"github.com/RichardKnop/machinery/v1"
	cache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/yukimochi/Activity-Relay/models"
)

var (
	version      string
	globalConfig *models.RelayConfig

	// Actor : Relay's Actor
	Actor models.Actor

	// WebfingerResource : Relay's Webfinger resource
	WebfingerResource models.WebfingerResource

	// Nodeinfo : Relay's Nodeinfo
	Nodeinfo models.NodeinfoResources

	relayState      models.RelayState
	machineryServer *machinery.Server
	actorCache      *cache.Cache
)

func Entrypoint(g *models.RelayConfig, v string) error {
	var err error
	globalConfig = g
	version = v

	err = initialize(globalConfig)
	if err != nil {
		return err
	}

	registResourceHandlers()

	logrus.Info("Staring API Server at ", globalConfig.ServerBind())
	err = http.ListenAndServe(globalConfig.ServerBind(), nil)
	if err != nil {
		return err
	}

	return nil
}

func initialize(globalConfig *models.RelayConfig) error {
	var err error

	redisClient := globalConfig.RedisClient()
	relayState = models.NewState(redisClient, true)
	relayState.ListenNotify(nil)

	machineryServer, err = models.NewMachineryServer(globalConfig)
	if err != nil {
		return err
	}

	Actor = models.NewActivityPubActorFromSelfKey(globalConfig)
	actorCache = cache.New(5*time.Minute, 10*time.Minute)

	WebfingerResource.GenerateFromActor(globalConfig.ServerHostname(), &Actor)
	Nodeinfo.GenerateFromActor(globalConfig.ServerHostname(), &Actor, version)

	return nil
}

func registResourceHandlers() {
	http.HandleFunc("/.well-known/nodeinfo", handleNodeinfoLink)
	http.HandleFunc("/.well-known/webfinger", handleWebfinger)
	http.HandleFunc("/nodeinfo/2.1", handleNodeinfo)
	http.HandleFunc("/actor", handleActor)
	http.HandleFunc("/inbox", func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, decodeActivity)
	})
}
