package models

import (
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestNewRelayConfig(t *testing.T) {
	t.Run("success valid configuration", func(t *testing.T) {
		relayConfig, err := NewRelayConfig()
		if err != nil {
			t.Fatal(err)
		}

		if relayConfig.serverBind != "0.0.0.0:8080" {
			t.Errorf("Expected RelayConfig.serverBind to be '0.0.0.0:8080', but got '%s'", relayConfig.serverBind)
		}
		if relayConfig.domain.Host != "relay.toot.yukimochi.jp" {
			t.Errorf("Expected RelayConfig.domain.Host to be 'relay.toot.yukimochi.jp', but got '%s'", relayConfig.domain.Host)
		}
		if relayConfig.serviceName != "YUKIMOCHI Toot Relay Service" {
			t.Errorf("Expected RelayConfig.serviceName to be 'YUKIMOCHI Toot Relay Service', but got '%s'", relayConfig.serviceName)
		}
		if relayConfig.serviceSummary != "YUKIMOCHI Toot Relay Service is Running by Activity-Relay" {
			t.Errorf("Expected RelayConfig.serviceSummary to be 'YUKIMOCHI Toot Relay Service is Running by Activity-Relay', but got '%s'", relayConfig.serviceSummary)
		}
		if relayConfig.serviceIconURL.String() != "https://example.com/example_icon.png" {
			t.Errorf("Expected RelayConfig.serviceIconURL to be 'https://example.com/example_icon.png', but got '%s'", relayConfig.serviceIconURL.String())
		}
		if relayConfig.serviceImageURL.String() != "https://example.com/example_image.png" {
			t.Errorf("Expected RelayConfig.serviceImageURL to be 'https://example.com/example_image.png', but got '%s'", relayConfig.serviceImageURL.String())
		}
	})

	t.Run("fail invalid configuration", func(t *testing.T) {
		invalidConfig := map[string]string{
			"ACTOR_PEM@notFound":        "../misc/test/notfound.pem",
			"ACTOR_PEM@invalidKey":      "../misc/test/actor.dh.pem",
			"REDIS_URL@invalidURL":      "",
			"REDIS_URL@unreachableHost": "redis://localhost:6380",
		}

		for key, value := range invalidConfig {
			viperKey := strings.Split(key, "@")[0]
			valid := viper.GetString(viperKey)

			viper.Set(viperKey, value)
			_, err := NewRelayConfig()
			if err == nil {
				t.Errorf("Expected error for invalid key '%s', but got nil", key)
			}

			viper.Set(viperKey, valid)
		}
	})
}

func createRelayConfig(t *testing.T) *RelayConfig {
	relayConfig, err := NewRelayConfig()
	if err != nil {
		t.Fatal(err)
	}

	return relayConfig
}

func TestRelayConfig_ServerBind(t *testing.T) {
	relayConfig := createRelayConfig(t)
	if relayConfig.ServerBind() != relayConfig.serverBind {
		t.Errorf("Expected ServerBind() to return '%s', but got '%s'", relayConfig.serverBind, relayConfig.ServerBind())
	}
}

func TestRelayConfig_ServerHostname(t *testing.T) {
	relayConfig := createRelayConfig(t)
	if relayConfig.ServerHostname() != relayConfig.domain {
		t.Errorf("Expected ServerHostname() to return '%v', but got '%v'", relayConfig.domain, relayConfig.ServerHostname())
	}
}

func TestRelayConfig_DumpWelcomeMessage(t *testing.T) {
	relayConfig := createRelayConfig(t)
	w := relayConfig.DumpWelcomeMessage("Testing", "")

	informations := map[string]string{
		"module NAME":     "Testing",
		"RELAY NAME":      relayConfig.serviceName,
		"RELAY DOMAIN":    relayConfig.domain.Host,
		"REDIS URL":       relayConfig.redisURL,
		"BIND ADDRESS":    relayConfig.serverBind,
		"JOB_CONCURRENCY": strconv.Itoa(relayConfig.jobConcurrency),
	}

	for key, information := range informations {
		if !strings.Contains(w, information) {
			t.Errorf("Expected welcome message to contain '%s' for key '%s', but not found", information, key)
		}
	}
}

func TestNewMachineryServer(t *testing.T) {
	relayConfig := createRelayConfig(t)

	_, err := NewMachineryServer(relayConfig)
	if err != nil {
		t.Errorf("Expected NewMachineryServer to succeed, but got error: %v", err)
	}
}
