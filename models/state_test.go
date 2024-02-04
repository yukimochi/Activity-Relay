package models

import (
	"context"
	"testing"
)

func TestLoadEmpty(t *testing.T) {
	relayState.RedisClient.FlushAll(context.TODO()).Result()
	relayState.Load()

	if relayState.RelayConfig.PersonOnly != false {
		t.Fatalf("fail - read config")
	}
	if relayState.RelayConfig.ManuallyAccept != false {
		t.Fatalf("fail - read config")
	}
}

func TestSetConfig(t *testing.T) {
	relayState.RedisClient.FlushAll(context.TODO()).Result()

	relayState.SetConfig(PersonOnly, true)
	<-ch
	if relayState.RelayConfig.PersonOnly != true {
		t.Fatalf("fail - enable config")
	}
	relayState.SetConfig(ManuallyAccept, true)
	<-ch
	if relayState.RelayConfig.ManuallyAccept != true {
		t.Fatalf("fail - enable config")
	}

	relayState.SetConfig(PersonOnly, false)
	<-ch
	if relayState.RelayConfig.PersonOnly != false {
		t.Fatalf("fail - disable config")
	}
	relayState.SetConfig(ManuallyAccept, false)
	<-ch
	if relayState.RelayConfig.ManuallyAccept != false {
		t.Fatalf("fail - disable config")
	}
}

func TestTreatSubscriptionNotify(t *testing.T) {
	relayState.RedisClient.FlushAll(context.TODO()).Result()

	relayState.AddSubscriber(Subscriber{
		Domain:   "example.com",
		InboxURL: "https://example.com/inbox",
	})
	<-ch

	valid := false
	for _, domain := range relayState.Subscribers {
		if domain.Domain == "example.com" && domain.InboxURL == "https://example.com/inbox" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("fail - write config")
	}

	relayState.DelSubscriber("example.com")
	<-ch

	for _, domain := range relayState.Subscribers {
		if domain.Domain == "example.com" {
			valid = false
		}
	}
	if !valid {
		t.Fatalf("fail - write config")
	}
}

func TestSelectDomain(t *testing.T) {
	relayState.RedisClient.FlushAll(context.TODO()).Result()

	exampleSubscription := Subscriber{
		Domain:   "example.com",
		InboxURL: "https://example.com/inbox",
	}

	relayState.AddSubscriber(exampleSubscription)
	<-ch

	subscription := relayState.SelectSubscriber("example.com")
	if *subscription != exampleSubscription {
		t.Fatalf("fail - select domain")
	}

	subscription = relayState.SelectSubscriber("example.org")
	if subscription != nil {
		t.Fatalf("fail - select domain")
	}
}

func TestBlockedDomain(t *testing.T) {
	relayState.RedisClient.FlushAll(context.TODO()).Result()

	relayState.SetBlockedDomain("example.com", true)
	<-ch

	valid := false
	for _, domain := range relayState.BlockedDomains {
		if domain == "example.com" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("fail - write config")
	}

	relayState.SetBlockedDomain("example.com", false)
	<-ch

	for _, domain := range relayState.BlockedDomains {
		if domain == "example.com" {
			valid = false
		}
	}
	if !valid {
		t.Fatalf("fail - write config")
	}
}

func TestLimitedDomain(t *testing.T) {
	relayState.RedisClient.FlushAll(context.TODO()).Result()

	relayState.SetLimitedDomain("example.com", true)
	<-ch

	valid := false
	for _, domain := range relayState.LimitedDomains {
		if domain == "example.com" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("fail - write config")
	}

	relayState.SetLimitedDomain("example.com", false)
	<-ch

	for _, domain := range relayState.LimitedDomains {
		if domain == "example.com" {
			valid = false
		}
	}
	if !valid {
		t.Fatalf("fail - write config")
	}
}

func TestLoadCompatibleSubscription(t *testing.T) {
	relayState.RedisClient.FlushAll(context.TODO()).Result()

	relayState.AddSubscriber(Subscriber{
		Domain:   "example.com",
		InboxURL: "https://example.com/inbox",
	})

	relayState.RedisClient.HDel(context.TODO(), "relay:subscription:example.com", "activity_id", "actor_id")
	relayState.Load()

	valid := false
	for _, domain := range relayState.Subscribers {
		if domain.Domain == "example.com" && domain.InboxURL == "https://example.com/inbox" {
			valid = true
		}
	}
	if !valid {
		t.Fatalf("fail - load compati config")
	}
}
