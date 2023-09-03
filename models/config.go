package models

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yukimochi/machinery-v1/v1"
	"github.com/yukimochi/machinery-v1/v1/config"
)

// RelayConfig contains valid configuration.
type RelayConfig struct {
	actorKey        *rsa.PrivateKey
	domain          *url.URL
	redisClient     *redis.Client
	redisURL        string
	serverBind      string
	serviceName     string
	serviceSummary  string
	serviceIconURL  *url.URL
	serviceImageURL *url.URL
	serviceLanding  string
	jobConcurrency  int
}

// NewRelayConfig create valid RelayConfig from viper configuration.
func NewRelayConfig() (*RelayConfig, error) {
	domain, err := url.ParseRequestURI("https://" + viper.GetString("RELAY_DOMAIN"))
	if err != nil {
		return nil, errors.New("RELAY_DOMAIN: " + err.Error())
	}

	iconURL, err := url.ParseRequestURI(viper.GetString("RELAY_ICON"))
	if err != nil {
		logrus.Warn("RELAY_ICON: INVALID OR EMPTY. THIS COLUMN IS DISABLED.")
		iconURL = nil
	}

	serviceLanding := viper.GetString("RELAY_LANDING")

	imageURL, err := url.ParseRequestURI(viper.GetString("RELAY_IMAGE"))
	if err != nil {
		logrus.Warn("RELAY_IMAGE: INVALID OR EMPTY. THIS COLUMN IS DISABLED.")
		imageURL = nil
	}

	jobConcurrency := viper.GetInt("JOB_CONCURRENCY")
	if jobConcurrency < 1 {
		return nil, errors.New("JOB_CONCURRENCY IS 0 OR EMPTY. SHOULD BE SET MORE THAN 1")
	}

	privateKey, err := readPrivateKeyRSA(viper.GetString("ACTOR_PEM"))
	if err != nil {
		return nil, errors.New("ACTOR_PEM: " + err.Error())
	}

	redisURL := viper.GetString("REDIS_URL")
	redisOption, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, errors.New("REDIS_URL: " + err.Error())
	}
	redisClient := redis.NewClient(redisOption)
	err = redisClient.Ping().Err()
	if err != nil {
		return nil, errors.New("REDIS_URL: " + err.Error())
	}

	serverBind := viper.GetString("RELAY_BIND")

	return &RelayConfig{
		actorKey:        privateKey,
		domain:          domain,
		redisClient:     redisClient,
		redisURL:        redisURL,
		serverBind:      serverBind,
		serviceName:     viper.GetString("RELAY_SERVICENAME"),
		serviceSummary:  viper.GetString("RELAY_SUMMARY"),
		serviceIconURL:  iconURL,
		serviceImageURL: imageURL,
		serviceLanding:  serviceLanding,
		jobConcurrency:  jobConcurrency,
	}, nil
}

// ServerBind is API Server's bind interface definition.
func (relayConfig *RelayConfig) ServerBind() string {
	return relayConfig.serverBind
}

// ServerHostname is API Server's hostname definition.
func (relayConfig *RelayConfig) ServerHostname() *url.URL {
	return relayConfig.domain
}

// ServerServiceSummary is API Server's service summary.
func (relayConfig *RelayConfig) ServerServiceSummary() string {
	return relayConfig.serviceSummary
}

// ServerServiceIcon is API Server's icon URL.
func (relayConfig *RelayConfig) ServerServiceIcon() *url.URL {
	return relayConfig.serviceIconURL
}

// ServerServiceImage is API Server's image URL.
func (relayConfig *RelayConfig) ServerServiceImage() *url.URL {
	return relayConfig.serviceImageURL
}

// ServerServiceName is API Server's servername definition.
func (relayConfig *RelayConfig) ServerServiceName() string {
	return relayConfig.serviceName
}

// ServerServiceLanding is API Server's service landing definition.
func (relayConfig *RelayConfig) ServerServiceLanding() string {
	return relayConfig.serviceLanding
}

// JobConcurrency is API Worker's jobConcurrency definition.
func (relayConfig *RelayConfig) JobConcurrency() int {
	return relayConfig.jobConcurrency
}

// ActorKey is API Worker's HTTPSignature private key.
func (relayConfig *RelayConfig) ActorKey() *rsa.PrivateKey {
	return relayConfig.actorKey
}

// RedisClient is return redis client from RelayConfig.
func (relayConfig *RelayConfig) RedisClient() *redis.Client {
	return relayConfig.redisClient
}

// DumpWelcomeMessage provide build and config information string.
func (relayConfig *RelayConfig) DumpWelcomeMessage(moduleName string, version string) string {
	return fmt.Sprintf(`Welcome to Activity-Relay %s - %s
 - Configuration
RELAY NAME      : %s
RELAY DOMAIN    : %s
REDIS URL       : %s
BIND ADDRESS    : %s
JOB_CONCURRENCY : %s
`, version, moduleName, relayConfig.serviceName, relayConfig.domain.Host, relayConfig.redisURL, relayConfig.serverBind, strconv.Itoa(relayConfig.jobConcurrency))
}

// NewMachineryServer create Redis backed Machinery Server from RelayConfig.
func NewMachineryServer(globalConfig *RelayConfig) (*machinery.Server, error) {
	cnf := &config.Config{
		Broker:          globalConfig.redisURL,
		DefaultQueue:    "relay",
		ResultBackend:   globalConfig.redisURL,
		ResultsExpireIn: 1,
	}
	newServer, err := machinery.NewServer(cnf)

	return newServer, err
}
