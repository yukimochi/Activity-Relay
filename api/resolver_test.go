package api

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPushActivityScriptReturnsSuccess(t *testing.T) {
	ctx := context.Background()
	client := RelayState.RedisClient
	key := "relay:activity:" + uuid.NewString()
	t.Cleanup(func() { client.Del(ctx, key) })

	result, err := client.Eval(ctx, pushActivityScript, []string{key}, "test-body", 2, 120).Int64()
	if err != nil {
		t.Fatalf("successful storage must not return redis.Nil: %v", err)
	}
	if result != 1 {
		t.Fatalf("expiry result = %d, want 1", result)
	}
	values, err := client.HGetAll(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	if values["body"] != "test-body" || values["remain_count"] != "2" {
		t.Fatalf("unexpected stored activity: %v", values)
	}
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	if ttl <= 0 || ttl > 120*time.Second {
		t.Fatalf("unexpected activity TTL: %v", ttl)
	}
}

func TestPushActivityScriptPreservesRedisErrors(t *testing.T) {
	ctx := context.Background()
	client := RelayState.RedisClient
	key := "relay:activity:" + uuid.NewString()
	t.Cleanup(func() { client.Del(ctx, key) })
	if err := client.Set(ctx, key, "not-a-hash", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	if err := client.Eval(ctx, pushActivityScript, []string{key}, "test-body", 2, 120).Err(); err == nil {
		t.Fatal("expected storage failure to remain an error")
	}
}
