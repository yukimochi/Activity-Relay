package models

import (
	"context"
	"fmt"
	"testing"
)

func TestRefresh(t *testing.T) {
	relayConfig, err := NewRelayConfigV2(RelayConfigV2BuilderOptions{
		WithServerConfig:   true,
		WithJobConcurrency: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	redisClient, _ := relayConfig.NewRedisClient(context.TODO())
	refreshResult := make(chan bool)

	NewStateV2(context.TODO(), relayConfig, RelayStateV2BuilderOptions{Refreshable: true, RefreshResult: refreshResult})

	redisClient.Publish(context.TODO(), "relay_refresh_v2", nil)
	<-refreshResult
	fmt.Println("-> State refreshed")

}
