package models

import (
	"context"
	"testing"
)

func TestRedisHGetOrCreateWithDefault(t *testing.T) {
	relayConfig := createRelayConfig(t)

	t.Run("Execute HGet when value exist", func(t *testing.T) {
		_, err := relayConfig.redisClient.HSet(context.TODO(), "gotest:redis:hget:or:create:with:default", "exist", "1").Result()
		if err != nil {
			t.Error(err)
		}

		value, err := redisHGetOrCreateWithDefault(relayConfig.redisClient, "gotest:redis:hget:or:create:with:default", "exist", "2")
		if err != nil {
			t.Fatalf("Expected no error from redisHGetOrCreateWithDefault, but got: %v", err)
		}
		if value != "1" {
			t.Fatalf("Expected value to be '1' (not overridden), but got '%s'", value)
		}

		_, err = relayConfig.redisClient.HDel(context.TODO(), "gotest:redis:hget:or:create:with:default", "exist").Result()
		if err != nil {
			t.Fatalf("Expected no error from HDel, but got: %v", err)
		}
	})

	t.Run("Execute HGet when value not exist", func(t *testing.T) {
		_, err := redisHGetOrCreateWithDefault(relayConfig.redisClient, "gotest:redis:hget:or:create:with:default", "not_exist", "2")
		if err != nil {
			t.Fatalf("Expected no error from redisHGetOrCreateWithDefault, but got: %v", err)
		}

		value, err := relayConfig.redisClient.HGet(context.TODO(), "gotest:redis:hget:or:create:with:default", "not_exist").Result()
		if err != nil {
			t.Fatalf("Expected no error from HGet, but got: %v", err)
		}

		if value != "2" {
			t.Fatalf("Expected value to be '2' (default written), but got '%s'", value)
		}

		_, err = relayConfig.redisClient.HDel(context.TODO(), "gotest:redis:hget:or:create:with:default", "not_exist").Result()
		if err != nil {
			t.Fatalf("Expected no error from HDel, but got: %v", err)
		}
	})
}
