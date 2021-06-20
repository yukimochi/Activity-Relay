package models

import (
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
			t.Error("Failed parse: RelayConfig.serverBind")
		}
		if relayConfig.domain.Host != "relay.toot.yukimochi.jp" {
			t.Error("Failed parse: RelayConfig.domain")
		}
		if relayConfig.serviceName != "YUKIMOCHI Toot Relay Service" {
			t.Error("Failed parse: RelayConfig.serviceName")
		}
		if relayConfig.serviceSummary != "YUKIMOCHI Toot Relay Service is Running by Activity-Relay" {
			t.Error("Failed parse: RelayConfig.serviceSummary")
		}
		if relayConfig.serviceIconURL.String() != "https://example.com/example_icon.png" {
			t.Error("Failed parse: RelayConfig.serviceIconURL")
		}
		if relayConfig.serviceImageURL.String() != "https://example.com/example_image.png" {
			t.Error("Failed parse: RelayConfig.serviceImageURL")
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
				t.Error("Failed catch error: " + key)
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
		t.Error("Failed accessor: ServerBind()")
	}
}

func TestRelayConfig_ServerHostname(t *testing.T) {
	relayConfig := createRelayConfig(t)
	if relayConfig.ServerHostname() != relayConfig.domain {
		t.Error("Failed accessor: ServerHostname()")
	}
}

func TestRelayConfig_DumpWelcomeMessage(t *testing.T) {
	relayConfig := createRelayConfig(t)
	w := relayConfig.DumpWelcomeMessage("Testing", "")

	informations := map[string]string{
		"module NAME":  "Testing",
		"RELAY NANE":   relayConfig.serviceName,
		"RELAY DOMAIN": relayConfig.domain.Host,
		"REDIS URL":    relayConfig.redisURL,
		"BIND ADDRESS": relayConfig.serverBind,
	}

	for key, information := range informations {
		if !strings.Contains(w, information) {
			t.Error("Missed welcome message information: ", key)
		}
	}
}

func TestNewMachineryServer(t *testing.T) {
	relayConfig := createRelayConfig(t)

	_, err := NewMachineryServer(relayConfig)
	if err != nil {
		t.Error("Failed create machinery server: ", err)
	}
}
