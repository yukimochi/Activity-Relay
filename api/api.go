package api

import (
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	metrics *defaultMetrics
)

func Entrypoint(g *models.RelayConfig, v string) error {
	var err error

	version = v
	GlobalConfig = g

	// Initialize the metrics
	metrics = newDefaultMetrics(prometheus.DefaultRegisterer, nil, nil)

	err = initialize(GlobalConfig)
	if err != nil {
		return err
	}

	handlersRegister()

	logrus.Info("Starting API Server at ", GlobalConfig.ServerBind())
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
	// Register the Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Register the new health and readiness endpoints
	http.Handle("/-/healthy", metrics.MetricsMiddleware(http.HandlerFunc(handleHealthy)))
	http.Handle("/-/ready", metrics.MetricsMiddleware(http.HandlerFunc(handleReady)))

	// Wrap handlers with the metrics middleware, preserving the URL path
	http.Handle("/.well-known/nodeinfo", metrics.MetricsMiddleware(http.HandlerFunc(handleNodeinfoLink)))
	http.Handle("/.well-known/webfinger", metrics.MetricsMiddleware(http.HandlerFunc(handleWebfinger)))
	http.Handle("/nodeinfo/2.1", metrics.MetricsMiddleware(http.HandlerFunc(handleNodeinfo)))
	http.Handle("/actor", metrics.MetricsMiddleware(http.HandlerFunc(handleRelayActor)))
	http.Handle("/inbox", metrics.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, decodeActivity)
	})))
}

// handleHealthy returns a health status message
func handleHealthy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ActivityRelay is Healthy."))
}

// handleReady returns a readiness status message
func handleReady(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ActivityRelay is Ready."))
}
