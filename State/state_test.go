package state

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

var redisClient *redis.Client

func TestMain(m *testing.M) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Config file is not exists. Use environment variables.")
		viper.BindEnv("redis_url")
	}
	redisOption, err := redis.ParseURL(viper.GetString("redis_url"))
	if err != nil {
		panic(err)
	}
	redisClient = redis.NewClient(redisOption)

	code := m.Run()
	os.Exit(code)
	redisClient.FlushAll().Result()
}

func TestLoadEmpty(t *testing.T) {
	redisClient.FlushAll().Result()
	testState := NewState(redisClient, false)

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

func TestSetConfig(t *testing.T) {
	ch := make(chan bool)
	redisClient.FlushAll().Result()
	testState := NewState(redisClient, true)
	testState.ListenNotify(ch)

	testState.SetConfig(BlockService, true)
	<-ch
	if testState.RelayConfig.BlockService != true {
		t.Fatalf("Failed enable config.")
	}
	testState.SetConfig(CreateAsAnnounce, true)
	<-ch
	if testState.RelayConfig.CreateAsAnnounce != true {
		t.Fatalf("Failed enable config.")
	}
	testState.SetConfig(ManuallyAccept, true)
	<-ch
	if testState.RelayConfig.ManuallyAccept != true {
		t.Fatalf("Failed enable config.")
	}

	testState.SetConfig(BlockService, false)
	<-ch
	if testState.RelayConfig.BlockService != false {
		t.Fatalf("Failed disable config.")
	}
	testState.SetConfig(CreateAsAnnounce, false)
	<-ch
	if testState.RelayConfig.CreateAsAnnounce != false {
		t.Fatalf("Failed disable config.")
	}
	testState.SetConfig(ManuallyAccept, false)
	<-ch
	if testState.RelayConfig.ManuallyAccept != false {
		t.Fatalf("Failed disable config.")
	}

	redisClient.FlushAll().Result()
}

func TestTreatSubscriptionNotify(t *testing.T) {
	ch := make(chan bool)
	redisClient.FlushAll().Result()
	testState := NewState(redisClient, true)
	testState.ListenNotify(ch)

	testState.AddSubscription(Subscription{
		Domain:   "example.com",
		InboxURL: "https://example.com/inbox",
	})
	<-ch

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
	<-ch

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

func TestSelectDomain(t *testing.T) {
	ch := make(chan bool)
	redisClient.FlushAll().Result()
	testState := NewState(redisClient, true)
	testState.ListenNotify(ch)

	exampleSubscription := Subscription{
		Domain:   "example.com",
		InboxURL: "https://example.com/inbox",
	}

	testState.AddSubscription(exampleSubscription)
	<-ch

	subscription := testState.SelectSubscription("example.com")
	if *subscription != exampleSubscription {
		t.Fatalf("Failed select domain.")
	}

	subscription = testState.SelectSubscription("example.org")
	if subscription != nil {
		t.Fatalf("Failed select domain.")
	}

	redisClient.FlushAll().Result()
}

func TestBlockedDomain(t *testing.T) {
	ch := make(chan bool)
	redisClient.FlushAll().Result()
	testState := NewState(redisClient, true)
	testState.ListenNotify(ch)

	testState.SetBlockedDomain("example.com", true)
	<-ch

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
	<-ch

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

func TestLimitedDomain(t *testing.T) {
	ch := make(chan bool)
	redisClient.FlushAll().Result()
	testState := NewState(redisClient, true)
	testState.ListenNotify(ch)

	testState.SetLimitedDomain("example.com", true)
	<-ch

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
	<-ch

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

func TestLoadCompatiSubscription(t *testing.T) {
	redisClient.FlushAll().Result()
	testState := NewState(redisClient, false)

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
