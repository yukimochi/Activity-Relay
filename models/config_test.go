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
			t.Error("fail - parse: RelayConfig.serverBind")
		}
		if relayConfig.domain.Host != "relay.toot.yukimochi.jp" {
			t.Error("fail - parse: RelayConfig.domain")
		}
		if relayConfig.serviceName != "YUKIMOCHI Toot Relay Service" {
			t.Error("fail - parse: RelayConfig.serviceName")
		}
		if relayConfig.serviceSummary != "YUKIMOCHI Toot Relay Service is Running by Activity-Relay" {
			t.Error("fail - parse: RelayConfig.serviceSummary")
		}
		if relayConfig.serviceIconURL.String() != "https://example.com/example_icon.png" {
			t.Error("fail - parse: RelayConfig.serviceIconURL")
		}
		if relayConfig.serviceImageURL.String() != "https://example.com/example_image.png" {
			t.Error("fail - parse: RelayConfig.serviceImageURL")
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
				t.Error("fail - invalid key should be raise error : " + key)
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
		t.Error("fail - accessor: ServerBind()")
	}
}

func TestRelayConfig_ServerHostname(t *testing.T) {
	relayConfig := createRelayConfig(t)
	if relayConfig.ServerHostname() != relayConfig.domain {
		t.Error("fail - accessor: ServerHostname()")
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
			t.Error("fail - lack welcome message information : ", key)
		}
	}
}

func TestNewMachineryServer(t *testing.T) {
	relayConfig := createRelayConfig(t)

	_, err := NewMachineryServer(relayConfig)
	if err != nil {
		t.Error("fail - machinery server can't create : ", err)
	}
}
