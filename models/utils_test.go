package models

import (
	"context"
	"errors"
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
			t.Error(err)
		}
		if value != "1" {
			t.Error(errors.New("value is override by redisHGetOrCreateWithDefault"))
		}

		_, err = relayConfig.redisClient.HDel(context.TODO(), "gotest:redis:hget:or:create:with:default", "exist").Result()
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Execute HGet when value not exist", func(t *testing.T) {
		_, err := redisHGetOrCreateWithDefault(relayConfig.redisClient, "gotest:redis:hget:or:create:with:default", "not_exist", "2")
		if err != nil {
			t.Error(err)
		}

		value, err := relayConfig.redisClient.HGet(context.TODO(), "gotest:redis:hget:or:create:with:default", "not_exist").Result()
		if err != nil {
			t.Error(err)
		}

		if value != "2" {
			t.Error(errors.New("redisHGetOrCreateWithDefault is not write default value successfully"))
		}

		_, err = relayConfig.redisClient.HDel(context.TODO(), "gotest:redis:hget:or:create:with:default", "not_exist").Result()
		if err != nil {
			t.Error(err)
		}
	})
}
