package models

import (
	"context"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestNewRelayConfigV2(t *testing.T) {
	t.Run("success generate valid options", func(t *testing.T) {
		relayConfig, err := NewRelayConfigV2(RelayConfigV2BuilderOptions{
			WithServerConfig:   true,
			WithJobConcurrency: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		// ServerConfig
		if relayConfig.serverConfig.Domain.Host != "relay.toot.yukimochi.jp" {
			t.Error("fail - parse: RelayConfig.serverConfig.Domain")
		}
		if relayConfig.serverConfig.Bind != "0.0.0.0:8080" {
			t.Error("fail - parse: RelayConfig.serverConfig.Bind")
		}
		if relayConfig.serverConfig.PrivateKey == nil {
			t.Error("fail - parse: RelayConfig.serverConfig.PrivateKey")
		}

		// ServiceConfig
		if relayConfig.serviceConfig.Name != "YUKIMOCHI Toot Relay Service" {
			t.Error("fail - parse: RelayConfig.serviceConfig.Name")
		}
		if relayConfig.serviceConfig.Summary != "YUKIMOCHI Toot Relay Service is Running by Activity-Relay" {
			t.Error("fail - parse: RelayConfig.serviceConfig.Summary")
		}
		if relayConfig.serviceConfig.IconURL.String() != "https://example.com/example_icon.png" {
			t.Error("fail - parse: RelayConfig.serviceConfig.IconURL")
		}
		if relayConfig.serviceConfig.ImageURL.String() != "https://example.com/example_image.png" {
			t.Error("fail - parse: RelayConfig.serviceConfig.ImageURL")
		}

		// RedisOptions
		if relayConfig.redisOptions == nil {
			t.Error("fail - parse: RelayConfig.redisOptions")
		}

		// JobConcurrency
		if relayConfig.jobConcurrency != 50 {
			t.Error("fail - parse: RelayConfig.jobConcurrency")
		}
	})

	t.Run("success generate valid options without serverConfig", func(t *testing.T) {
		relayConfig, err := NewRelayConfigV2(RelayConfigV2BuilderOptions{
			WithServerConfig:   false,
			WithJobConcurrency: true,
		})
		if err != nil {
			t.Fatal(err)
		}

		// ServerConfig
		if relayConfig.serverConfig != nil {
			t.Error("fail - parse: RelayConfig.serverConfig")
		}
	})

	t.Run("success generate valid options without jobConcurrency", func(t *testing.T) {
		relayConfig, err := NewRelayConfigV2(RelayConfigV2BuilderOptions{
			WithServerConfig:   true,
			WithJobConcurrency: false,
		})
		if err != nil {
			t.Fatal(err)
		}

		// JobConcurrency
		if relayConfig.jobConcurrency != 0 {
			t.Error("fail - parse: RelayConfig.jobConcurrency")
		}
	})

	t.Run("fail invalid options", func(t *testing.T) {
		invalidExamples := map[string]string{
			"ACTOR_PEM@notFound":   "../misc/test/notfound.pem",
			"REDIS_URL@invalidURL": "",
		}

		for key, invalidValue := range invalidExamples {
			viperKey := strings.Split(key, "@")[0]
			validValue := viper.GetString(viperKey)

			viper.Set(viperKey, invalidValue)
			_, err := NewRelayConfig()
			if err == nil {
				t.Error("fail - invalid value should be raise error : " + key)
			}

			viper.Set(viperKey, validValue)
		}
	})
}

func TestNewRedisClient(t *testing.T) {
	t.Run("success create client for reachable redis serer", func(t *testing.T) {
		relayConfig, err := NewRelayConfigV2(RelayConfigV2BuilderOptions{
			WithServerConfig:   false,
			WithJobConcurrency: false,
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = relayConfig.NewRedisClient(context.Background())
		if err != nil {
			t.Error("fail - create client for reachable redis serer")
		}
	})

	t.Run("fail create client for unreachable redis serer", func(t *testing.T) {
		validURL := viper.GetString("REDIS_URL")
		viper.Set("REDIS_URL", "redis://localhost:6380")

		relayConfig, err := NewRelayConfigV2(RelayConfigV2BuilderOptions{
			WithServerConfig:   false,
			WithJobConcurrency: false,
		})
		if err != nil {
			t.Fatal(err)
		}

		_, err = relayConfig.NewRedisClient(context.Background())
		if err == nil {
			t.Error("fail - create client for unreachable redis serer")
		}

		viper.Set("REDIS_URL", validURL)
	})
}
