package state

import (
	"os"
	"testing"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

var redisClient *redis.Client

func TestMain(m *testing.M) {
	viper.BindEnv("redis_url")
	redisOption, err := redis.ParseURL(viper.GetString("redis_url"))
	if err != nil {
		panic(err)
	}
	redisClient = redis.NewClient(redisOption)

	code := m.Run()
	os.Exit(code)
	redisClient.FlushAll().Result()
}

func TestInitialLoad(t *testing.T) {
	redisClient.FlushAll().Result()
	testState := NewState(redisClient)

	if testState.RelayConfig.BlockService != false {
		t.Fatalf("Failed read config.")
	}
	if testState.RelayConfig.CreateAsAnnounce != false {
		t.Fatalf("Failed read config.")
	}
	if testState.RelayConfig.ManuallyAccept != false {
		t.Fatalf("Failed read config.")
	}

	redisClient.FlushAll().Result()
}

func TestAddLimited(t *testing.T) {
	redisClient.FlushAll().Result()
	testState := NewState(redisClient)

	testState.SetLimitedDomain("example.com", true)

	valid := false
	for _, domain := range testState.LimitedDomains {
		if domain == "example.com" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}

	testState.SetLimitedDomain("example.com", false)

	for _, domain := range testState.LimitedDomains {
		if domain == "example.com" {
			valid = false
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}

	redisClient.FlushAll().Result()
}

func TestAddBlocked(t *testing.T) {
	redisClient.FlushAll().Result()
	testState := NewState(redisClient)

	testState.SetBlockedDomain("example.com", true)

	valid := false
	for _, domain := range testState.BlockedDomains {
		if domain == "example.com" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}

	testState.SetBlockedDomain("example.com", false)

	for _, domain := range testState.BlockedDomains {
		if domain == "example.com" {
			valid = false
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}

	redisClient.FlushAll().Result()
}

func TestAddSubscription(t *testing.T) {
	redisClient.FlushAll().Result()
	testState := NewState(redisClient)

	testState.AddSubscription(Subscription{
		Domain:   "example.com",
		InboxURL: "https://example.com/inbox",
	})

	valid := false
	for _, domain := range testState.Subscriptions {
		if domain.Domain == "example.com" && domain.InboxURL == "https://example.com/inbox" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}

	testState.DelSubscription("example.com")

	for _, domain := range testState.Subscriptions {
		if domain.Domain == "example.com" {
			valid = false
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}

	redisClient.FlushAll().Result()
}

func TestLoadCompatiSubscription(t *testing.T) {
	redisClient.FlushAll().Result()
	testState := NewState(redisClient)

	testState.AddSubscription(Subscription{
		Domain:   "example.com",
		InboxURL: "https://example.com/inbox",
	})

	testState.RedisClient.HDel("relay:subscription:example.com", "activity_id", "actor_id")
	testState.Load()

	valid := false
	for _, domain := range testState.Subscriptions {
		if domain.Domain == "example.com" && domain.InboxURL == "https://example.com/inbox" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("Failed load compati config.")
	}

	redisClient.FlushAll().Result()
}

func TestSetConfig(t *testing.T) {
	redisClient.FlushAll().Result()
	testState := NewState(redisClient)

	testState.SetConfig(BlockService, true)
	if testState.RelayConfig.BlockService != true {
		t.Fatalf("Failed enable config.")
	}
	testState.SetConfig(CreateAsAnnounce, true)
	if testState.RelayConfig.CreateAsAnnounce != true {
		t.Fatalf("Failed enable config.")
	}
	testState.SetConfig(ManuallyAccept, true)
	if testState.RelayConfig.ManuallyAccept != true {
		t.Fatalf("Failed enable config.")
	}

	testState.SetConfig(BlockService, false)
	if testState.RelayConfig.BlockService != false {
		t.Fatalf("Failed disable config.")
	}
	testState.SetConfig(CreateAsAnnounce, false)
	if testState.RelayConfig.CreateAsAnnounce != false {
		t.Fatalf("Failed disable config.")
	}
	testState.SetConfig(ManuallyAccept, false)
	if testState.RelayConfig.ManuallyAccept != false {
		t.Fatalf("Failed disable config.")
	}

	redisClient.FlushAll().Result()
}
