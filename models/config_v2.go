package models

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/yukimochi/machinery-v1/v1"
	machineryConfig "github.com/yukimochi/machinery-v1/v1/config"
)

type ServerConfig struct {
	Domain     *url.URL
	Bind       string
	PrivateKey *rsa.PrivateKey
}

type ServiceConfig struct {
	Name     string
	Summary  string
	IconURL  *url.URL
	ImageURL *url.URL
}

type RelayConfigV2 struct {
	serverConfig   *ServerConfig
	serviceConfig  *ServiceConfig
	redisOptions   *redis.Options
	jobConcurrency int
}

type RelayConfigV2BuilderOptions struct {
	WithServerConfig   bool
	WithJobConcurrency bool
}

func buildServerConfig() (*ServerConfig, error) {
	domain, err := url.ParseRequestURI("https://" + viper.GetString("RELAY_DOMAIN"))
	if err != nil {
		return nil, errors.New("RELAY_DOMAIN: " + err.Error())
	}

	privateKey, err := readPrivateKeyRSA(viper.GetString("ACTOR_PEM"))
	if err != nil {
		return nil, errors.New("ACTOR_PEM: " + err.Error())
	}

	return &ServerConfig{
		Domain:     domain,
		Bind:       viper.GetString("RELAY_BIND"),
		PrivateKey: privateKey,
	}, nil
}

func buildServiceConfig() (*ServiceConfig, error) {
	// Required fields
	name := viper.GetString("RELAY_SERVICENAME")
	summary := viper.GetString("RELAY_SUMMARY")

	// Optional fields
	iconURL, err := url.ParseRequestURI(viper.GetString("RELAY_ICON"))
	if err != nil {
		logrus.Warn("RELAY_ICON: INVALID OR EMPTY. THIS COLUMN IS DISABLED.")
		iconURL = nil
	}
	imageURL, err := url.ParseRequestURI(viper.GetString("RELAY_IMAGE"))
	if err != nil {
		logrus.Warn("RELAY_IMAGE: INVALID OR EMPTY. THIS COLUMN IS DISABLED.")
		imageURL = nil
	}

	return &ServiceConfig{
		Name:     name,
		Summary:  summary,
		IconURL:  iconURL,
		ImageURL: imageURL,
	}, nil
}

func NewRelayConfigV2(options RelayConfigV2BuilderOptions) (*RelayConfigV2, error) {
	result := RelayConfigV2{}

	serviceOptions, err := buildServiceConfig()
	if err != nil {
		return nil, err
	}
	result.serviceConfig = serviceOptions

	redisOptions, err := redis.ParseURL(viper.GetString("REDIS_URL"))
	if err != nil {
		return nil, errors.New("REDIS_URL: " + err.Error())
	}
	result.redisOptions = redisOptions

	// Works with API Server
	if options.WithServerConfig {
		serverOptions, err := buildServerConfig()
		if err != nil {
			return nil, err
		}
		result.serverConfig = serverOptions
	} else {
		result.serverConfig = nil
	}

	// Works with Job Worker
	if options.WithJobConcurrency {
		viper.SetDefault("JOB_CONCURRENCY", 10)
		jobConcurrency := viper.GetInt("JOB_CONCURRENCY")
		if jobConcurrency < 1 {
			return nil, errors.New("JOB_CONCURRENCY: Invalid Value")
		}
		result.jobConcurrency = jobConcurrency
	} else {
		result.jobConcurrency = 0
	}

	return &result, nil
}

// ServerConfig is API Server options.
func (config *RelayConfigV2) ServerConfig() (*ServerConfig, error) {
	if config.serverConfig != nil {
		return config.serverConfig, nil
	}
	return nil, errors.New("this configuration does not have ServerConfig")
}

// ServiceConfig is Relay Service options.
func (config *RelayConfigV2) ServiceConfig() *ServiceConfig {
	return config.serviceConfig
}

// RedisOptions is Redis options.
func (config *RelayConfigV2) RedisOptions() *redis.Options {
	return config.redisOptions
}

// JobConcurrency is Job concurrency.
func (config *RelayConfigV2) JobConcurrency() (int, error) {
	if config.jobConcurrency == 0 {
		return 0, errors.New("this configuration does not have JobConcurrency")
	}
	return config.jobConcurrency, nil
}

func (config *RelayConfigV2) DumpWelcomeMessage(moduleName string, version string) string {
	message := fmt.Sprintf(`Welcome to Activity-Relay %s - %s\n`, version, moduleName)
	message += fmt.Sprintf(`- Configuration\n`)
	message += fmt.Sprintf(`RELAY NAME      : %s\n`, config.serviceConfig.Name)
	message += fmt.Sprintf(`REDIS URL       : %s\n`, config.redisOptions.Addr)
	if config.serverConfig != nil {
		message += fmt.Sprintf(`RELAY DOMAIN    : %s\n`, config.serverConfig.Domain.Host)
		message += fmt.Sprintf(`BIND ADDRESS    : %s\n`, config.serverConfig.Bind)
	}
	if config.jobConcurrency != 0 {
		message += fmt.Sprintf(`JOB_CONCURRENCY : %d\n`, config.jobConcurrency)
	}

	return message
}

func (config *RelayConfigV2) NewRedisClient(ctx context.Context) (*redis.Client, error) {
	redisClient := redis.NewClient(config.redisOptions)

	pCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(pCtx).Result()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
}

func (config *RelayConfigV2) NewMachineryServer() (*machinery.Server, error) {
	cnf := &machineryConfig.Config{
		Broker:          config.redisOptions.Addr,
		DefaultQueue:    "relay",
		ResultBackend:   config.redisOptions.Addr,
		ResultsExpireIn: 1,
	}
	server, err := machinery.NewServer(cnf)
	if err != nil {
		return nil, err
	}
	return server, nil
}
