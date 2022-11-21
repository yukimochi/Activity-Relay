package control

import (
	"bytes"
	"strings"
	"testing"
)

func TestListDomainSubscriber(t *testing.T) {
	RelayState.RedisClient.FlushAll().Result()

	app := configCmdInit()
	app.SetArgs([]string{"import", "--json", "../misc/test/exampleConfig.json"})
	app.Execute()
	RelayState.Load()

	buffer := new(bytes.Buffer)

	app = domainCmdInit()
	app.SetOut(buffer)
	app.SetArgs([]string{"list"})
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
	RelayState.RedisClient.FlushAll().Result()

	app := configCmdInit()

	app.SetArgs([]string{"import", "--json", "../misc/test/exampleConfig.json"})
	app.Execute()
	RelayState.Load()

	buffer := new(bytes.Buffer)

	app = domainCmdInit()
	app.SetOut(buffer)
	app.SetArgs([]string{"list", "-t", "limited"})
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
	RelayState.RedisClient.FlushAll().Result()

	app := configCmdInit()

	app.SetArgs([]string{"import", "--json", "../misc/test/exampleConfig.json"})
	app.Execute()
	RelayState.Load()

	buffer := new(bytes.Buffer)

	app = domainCmdInit()
	app.SetOut(buffer)
	app.SetArgs([]string{"list", "-t", "blocked"})
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
	RelayState.RedisClient.FlushAll().Result()

	app := domainCmdInit()

	app.SetArgs([]string{"set", "-t", "blocked", "testdomain.example.jp"})
	app.Execute()
	RelayState.Load()

	valid := false
	for _, domain := range RelayState.BlockedDomains {
		if domain == "testdomain.example.jp" {
			valid = true
		}
	}

	if !valid {
		t.Fatalf("Not set blocked domain")
	}
}

func TestSetDomainLimited(t *testing.T) {
	RelayState.RedisClient.FlushAll().Result()

	app := domainCmdInit()

	app.SetArgs([]string{"set", "-t", "limited", "testdomain.example.jp"})
	app.Execute()
	RelayState.Load()

	valid := false
	for _, domain := range RelayState.LimitedDomains {
		if domain == "testdomain.example.jp" {
			valid = true
		}
	}

	if !valid {
		t.Fatalf("Not set limited domain")
	}
}

func TestUnsetDomainBlocked(t *testing.T) {
	RelayState.RedisClient.FlushAll().Result()

	app := configCmdInit()

	app.SetArgs([]string{"import", "--json", "../misc/test/exampleConfig.json"})
	app.Execute()

	app = domainCmdInit()
	app.SetArgs([]string{"set", "-t", "blocked", "-u", "blockedDomain.example.jp"})
	app.Execute()
	RelayState.Load()

	valid := true
	for _, domain := range RelayState.BlockedDomains {
		if domain == "blockedDomain.example.jp" {
			valid = false
		}
	}

	if !valid {
		t.Fatalf("Not unset blocked domain")
	}
}

func TestUnsetDomainLimited(t *testing.T) {
	RelayState.RedisClient.FlushAll().Result()

	app := configCmdInit()

	app.SetArgs([]string{"import", "--json", "../misc/test/exampleConfig.json"})
	app.Execute()

	app = domainCmdInit()
	app.SetArgs([]string{"set", "-t", "limited", "-u", "limitedDomain.example.jp"})
	app.Execute()
	RelayState.Load()

	valid := true
	for _, domain := range RelayState.LimitedDomains {
		if domain == "limitedDomain.example.jp" {
			valid = false
		}
	}

	if !valid {
		t.Fatalf("Not unset blocked domain")
	}
}

func TestSetDomainInvalid(t *testing.T) {
	RelayState.RedisClient.FlushAll().Result()

	app := configCmdInit()

	app.SetArgs([]string{"import", "--json", "../misc/test/exampleConfig.json"})
	app.Execute()
	RelayState.Load()

	buffer := new(bytes.Buffer)

	app = domainCmdInit()
	app.SetOut(buffer)
	app.SetArgs([]string{"set", "-t", "hoge", "hoge.example.jp"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid type given" {
		t.Fatalf("Invalid Response.")
	}
}

func TestUnfollowDomain(t *testing.T) {
	RelayState.RedisClient.FlushAll().Result()

	app := configCmdInit()

	app.SetArgs([]string{"import", "--json", "../misc/test/exampleConfig.json"})
	app.Execute()

	app = domainCmdInit()
	app.SetArgs([]string{"unfollow", "subscription.example.jp"})
	app.Execute()
	RelayState.Load()

	valid := true
	for _, domain := range RelayState.Subscriptions {
		if domain.Domain == "subscription.example.jp" {
			valid = false
		}
	}

	if !valid {
		t.Fatalf("Not unfollowed domain")
	}
}

func TestInvalidUnfollowDomain(t *testing.T) {
	RelayState.RedisClient.FlushAll().Result()

	app := configCmdInit()

	app.SetArgs([]string{"import", "--json", "../misc/test/exampleConfig.json"})
	app.Execute()
	RelayState.Load()

	buffer := new(bytes.Buffer)

	app = domainCmdInit()
	app.SetOut(buffer)
	app.SetArgs([]string{"unfollow", "unknown.tld"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid domain [unknown.tld] given" {
		t.Fatalf("Invalid Response.")
	}
}
