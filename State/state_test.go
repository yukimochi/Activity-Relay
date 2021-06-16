package state

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

var redisClient *redis.Client
var relayState RelayState
var ch chan bool

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
	redisClient.FlushAll().Result()

	ch = make(chan bool)
	relayState = NewState(redisClient, true)
	relayState.ListenNotify(ch)

	code := m.Run()
	redisClient.FlushAll().Result()

	os.Exit(code)
}

func TestLoadEmpty(t *testing.T) {
	redisClient.FlushAll().Result()

	if relayState.RelayConfig.BlockService != false {
		t.Fatalf("Failed read config.")
	}
	if relayState.RelayConfig.CreateAsAnnounce != false {
		t.Fatalf("Failed read config.")
	}
	if relayState.RelayConfig.ManuallyAccept != false {
		t.Fatalf("Failed read config.")
	}
}

func TestSetConfig(t *testing.T) {
	redisClient.FlushAll().Result()

	relayState.SetConfig(BlockService, true)
	<-ch
	if relayState.RelayConfig.BlockService != true {
		t.Fatalf("Failed enable config.")
	}
	relayState.SetConfig(CreateAsAnnounce, true)
	<-ch
	if relayState.RelayConfig.CreateAsAnnounce != true {
		t.Fatalf("Failed enable config.")
	}
	relayState.SetConfig(ManuallyAccept, true)
	<-ch
	if relayState.RelayConfig.ManuallyAccept != true {
		t.Fatalf("Failed enable config.")
	}

	relayState.SetConfig(BlockService, false)
	<-ch
	if relayState.RelayConfig.BlockService != false {
		t.Fatalf("Failed disable config.")
	}
	relayState.SetConfig(CreateAsAnnounce, false)
	<-ch
	if relayState.RelayConfig.CreateAsAnnounce != false {
		t.Fatalf("Failed disable config.")
	}
	relayState.SetConfig(ManuallyAccept, false)
	<-ch
	if relayState.RelayConfig.ManuallyAccept != false {
		t.Fatalf("Failed disable config.")
	}
}

func TestTreatSubscriptionNotify(t *testing.T) {
	redisClient.FlushAll().Result()

	relayState.AddSubscription(Subscription{
		Domain:   "example.com",
		InboxURL: "https://example.com/inbox",
	})
	<-ch

	valid := false
	for _, domain := range relayState.Subscriptions {
		if domain.Domain == "example.com" && domain.InboxURL == "https://example.com/inbox" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}

	relayState.DelSubscription("example.com")
	<-ch

	for _, domain := range relayState.Subscriptions {
		if domain.Domain == "example.com" {
			valid = false
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}
}

func TestSelectDomain(t *testing.T) {
	redisClient.FlushAll().Result()

	exampleSubscription := Subscription{
		Domain:   "example.com",
		InboxURL: "https://example.com/inbox",
	}

	relayState.AddSubscription(exampleSubscription)
	<-ch

	subscription := relayState.SelectSubscription("example.com")
	if *subscription != exampleSubscription {
		t.Fatalf("Failed select domain.")
	}

	subscription = relayState.SelectSubscription("example.org")
	if subscription != nil {
		t.Fatalf("Failed select domain.")
	}
}

func TestBlockedDomain(t *testing.T) {
	redisClient.FlushAll().Result()

	relayState.SetBlockedDomain("example.com", true)
	<-ch

	valid := false
	for _, domain := range relayState.BlockedDomains {
		if domain == "example.com" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}

	relayState.SetBlockedDomain("example.com", false)
	<-ch

	for _, domain := range relayState.BlockedDomains {
		if domain == "example.com" {
			valid = false
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}
}

func TestLimitedDomain(t *testing.T) {
	redisClient.FlushAll().Result()

	relayState.SetLimitedDomain("example.com", true)
	<-ch

	valid := false
	for _, domain := range relayState.LimitedDomains {
		if domain == "example.com" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}

	relayState.SetLimitedDomain("example.com", false)
	<-ch

	for _, domain := range relayState.LimitedDomains {
		if domain == "example.com" {
			valid = false
		}
	}
	if !valid {
		t.Fatalf("Failed write config.")
	}
}

func TestLoadCompatiSubscription(t *testing.T) {
	redisClient.FlushAll().Result()

	relayState.AddSubscription(Subscription{
		Domain:   "example.com",
		InboxURL: "https://example.com/inbox",
	})

	relayState.RedisClient.HDel("relay:subscription:example.com", "activity_id", "actor_id")
	relayState.Load()

	valid := false
	for _, domain := range relayState.Subscriptions {
		if domain.Domain == "example.com" && domain.InboxURL == "https://example.com/inbox" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("Failed load compati config.")
	}
}
