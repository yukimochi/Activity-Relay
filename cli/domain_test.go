package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestListDomainSubscriber(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := buildNewCmd()
	app.SetArgs([]string{"config", "import", "--json", "../misc/exampleConfig.json"})
	app.Execute()
	relayState.Load()

	buffer := new(bytes.Buffer)
	app.SetOutput(buffer)

	app.SetArgs([]string{"domain", "list"})
	app.Execute()

	output := buffer.String()
	valid := ` - Subscriber domain :
subscription.example.jp
Total : 1
`
	if output != valid {
		t.Fatalf("Invalid Response.")
	}
}

func TestListDomainLimited(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := buildNewCmd()

	app.SetArgs([]string{"config", "import", "--json", "../misc/exampleConfig.json"})
	app.Execute()
	relayState.Load()

	buffer := new(bytes.Buffer)
	app.SetOutput(buffer)

	app.SetArgs([]string{"domain", "list", "-t", "limited"})
	app.Execute()

	output := buffer.String()
	valid := ` - Limited domain :
limitedDomain.example.jp
Total : 1
`
	if output != valid {
		t.Fatalf("Invalid Response.")
	}
}

func TestListDomainBlocked(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := buildNewCmd()

	app.SetArgs([]string{"config", "import", "--json", "../misc/exampleConfig.json"})
	app.Execute()
	relayState.Load()

	buffer := new(bytes.Buffer)
	app.SetOutput(buffer)

	app.SetArgs([]string{"domain", "list", "-t", "blocked"})
	app.Execute()

	output := buffer.String()
	valid := ` - Blocked domain :
blockedDomain.example.jp
Total : 1
`
	if output != valid {
		t.Fatalf("Invalid Response.")
	}
}

func TestSetDomainBlocked(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := buildNewCmd()

	app.SetArgs([]string{"domain", "set", "-t", "blocked", "testdomain.example.jp"})
	app.Execute()
	relayState.Load()

	valid := false
	for _, domain := range relayState.BlockedDomains {
		if domain == "testdomain.example.jp" {
			valid = true
		}
	}

	if !valid {
		t.Fatalf("Not set blocked domain")
	}
}

func TestSetDomainLimited(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := buildNewCmd()

	app.SetArgs([]string{"domain", "set", "-t", "limited", "testdomain.example.jp"})
	app.Execute()
	relayState.Load()

	valid := false
	for _, domain := range relayState.LimitedDomains {
		if domain == "testdomain.example.jp" {
			valid = true
		}
	}

	if !valid {
		t.Fatalf("Not set limited domain")
	}
}

func TestUnsetDomainBlocked(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := buildNewCmd()

	app.SetArgs([]string{"config", "import", "--json", "../misc/exampleConfig.json"})
	app.Execute()

	app.SetArgs([]string{"domain", "set", "-t", "blocked", "-u", "blockedDomain.example.jp"})
	app.Execute()
	relayState.Load()

	valid := true
	for _, domain := range relayState.BlockedDomains {
		if domain == "blockedDomain.example.jp" {
			valid = false
		}
	}

	if !valid {
		t.Fatalf("Not unset blocked domain")
	}
}

func TestUnsetDomainLimited(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := buildNewCmd()

	app.SetArgs([]string{"config", "import", "--json", "../misc/exampleConfig.json"})
	app.Execute()

	app.SetArgs([]string{"domain", "set", "-t", "limited", "-u", "limitedDomain.example.jp"})
	app.Execute()
	relayState.Load()

	valid := true
	for _, domain := range relayState.LimitedDomains {
		if domain == "limitedDomain.example.jp" {
			valid = false
		}
	}

	if !valid {
		t.Fatalf("Not unset blocked domain")
	}
}

func TestSetDomainInvalid(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := buildNewCmd()

	app.SetArgs([]string{"config", "import", "--json", "../misc/exampleConfig.json"})
	app.Execute()
	relayState.Load()

	buffer := new(bytes.Buffer)
	app.SetOutput(buffer)

	app.SetArgs([]string{"domain", "set", "-t", "hoge", "hoge.example.jp"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid type given" {
		t.Fatalf("Invalid Response.")
	}
}

func TestUnfollowDomain(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := buildNewCmd()

	app.SetArgs([]string{"config", "import", "--json", "../misc/exampleConfig.json"})
	app.Execute()

	app.SetArgs([]string{"domain", "unfollow", "subscription.example.jp"})
	app.Execute()
	relayState.Load()

	valid := true
	for _, domain := range relayState.Subscriptions {
		if domain.Domain == "subscription.example.jp" {
			valid = false
		}
	}

	if !valid {
		t.Fatalf("Not unfollowed domain")
	}
}

func TestInvalidUnfollowDomain(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := buildNewCmd()

	app.SetArgs([]string{"config", "import", "--json", "../misc/exampleConfig.json"})
	app.Execute()
	relayState.Load()

	buffer := new(bytes.Buffer)
	app.SetOutput(buffer)

	app.SetArgs([]string{"domain", "unfollow", "unknown.tld"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid domain [unknown.tld] given" {
		t.Fatalf("Invalid Response.")
	}
}
