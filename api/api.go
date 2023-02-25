package api

import (
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/yukimochi/Activity-Relay/models"
	"github.com/yukimochi/machinery-v1/v1"
)

var (
	version      string
	GlobalConfig *models.RelayConfig

	// RelayActor : Relay's Actor
	RelayActor models.Actor
	// Nodeinfo : Relay's Nodeinfo
	Nodeinfo models.NodeinfoResources
	// WebfingerResources : Relay's Webfinger Resources
	WebfingerResources []models.WebfingerResource

	ActorCache      *cache.Cache
	MachineryServer *machinery.Server
	RelayState      models.RelayState
)

func Entrypoint(g *models.RelayConfig, v string) error {
	var err error

	version = v
	GlobalConfig = g

	err = initialize(GlobalConfig)
	if err != nil {
		return err
	}

	handlersRegister()

	logrus.Info("Staring API Server at ", GlobalConfig.ServerBind())
	err = http.ListenAndServe(GlobalConfig.ServerBind(), nil)
	if err != nil {
		return err
	}

	return nil
}

func initialize(globalConfig *models.RelayConfig) error {
	var err error

	redisClient := globalConfig.RedisClient()
	RelayState = models.NewState(redisClient, true)
	RelayState.ListenNotify(nil)

	MachineryServer, err = models.NewMachineryServer(globalConfig)
	if err != nil {
		return err
	}

	RelayActor = models.NewActivityPubActorFromRelayConfig(globalConfig)
	ActorCache = cache.New(5*time.Minute, 10*time.Minute)

	Nodeinfo = models.GenerateNodeinfoResources(globalConfig.ServerHostname(), version)
	WebfingerResources = append(WebfingerResources, RelayActor.GenerateWebfingerResource(globalConfig.ServerHostname()))

	return nil
}

func handlersRegister() {
	http.HandleFunc("/.well-known/nodeinfo", handleNodeinfoLink)
	http.HandleFunc("/.well-known/webfinger", handleWebfinger)
	http.HandleFunc("/nodeinfo/2.1", handleNodeinfo)
	http.HandleFunc("/actor", handleRelayActor)
	http.HandleFunc("/inbox", func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, decodeActivity)
	})
}
