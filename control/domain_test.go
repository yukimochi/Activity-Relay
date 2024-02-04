package control

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestListDomainSubscriber(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	file, err := os.Open("../misc/test/exampleConfig.json")
	if err != nil {
		t.Fatalf("Test resource fetch error.")
	}
	jsonData, _ := io.ReadAll(file)

	app.SetArgs([]string{"import", "--data", string(jsonData)})
	app.Execute()
	RelayState.Load()

	buffer := new(bytes.Buffer)

	app = domainCmdInit()
	app.SetOut(buffer)
	app.SetArgs([]string{"list"})
	app.Execute()

	output := buffer.String()
	valid := ` - Subscriber list :
[*] subscription.example.jp
 - Follower list :
Total : 1
`
	if output != valid {
		t.Fatalf("Invalid Response.")
	}
}

func TestListDomainLimited(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	file, err := os.Open("../misc/test/exampleConfig.json")
	if err != nil {
		t.Fatalf("Test resource fetch error.")
	}
	jsonData, _ := io.ReadAll(file)

	app.SetArgs([]string{"import", "--data", string(jsonData)})
	app.Execute()
	RelayState.Load()

	buffer := new(bytes.Buffer)

	app = domainCmdInit()
	app.SetOut(buffer)
	app.SetArgs([]string{"list", "-t", "limited"})
	app.Execute()

	output := buffer.String()
	valid := ` - Limited domains :
limitedDomain.example.jp
Total : 1
`
	if output != valid {
		t.Fatalf("Invalid Response.")
	}
}

func TestListDomainBlocked(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	file, err := os.Open("../misc/test/exampleConfig.json")
	if err != nil {
		t.Fatalf("Test resource fetch error.")
	}
	jsonData, _ := io.ReadAll(file)

	app.SetArgs([]string{"import", "--data", string(jsonData)})
	app.Execute()
	RelayState.Load()

	buffer := new(bytes.Buffer)

	app = domainCmdInit()
	app.SetOut(buffer)
	app.SetArgs([]string{"list", "-t", "blocked"})
	app.Execute()

	output := buffer.String()
	valid := ` - Blocked domains :
blockedDomain.example.jp
Total : 1
`
	if output != valid {
		t.Fatalf("Invalid Response.")
	}
}

func TestSetDomainBlocked(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

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
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

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
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	file, err := os.Open("../misc/test/exampleConfig.json")
	if err != nil {
		t.Fatalf("Test resource fetch error.")
	}
	jsonData, _ := io.ReadAll(file)

	app.SetArgs([]string{"import", "--data", string(jsonData)})
	app.Execute()

	app = domainCmdInit()
	app.SetArgs([]string{"unset", "-t", "blocked", "blockedDomain.example.jp"})
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
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	file, err := os.Open("../misc/test/exampleConfig.json")
	if err != nil {
		t.Fatalf("Test resource fetch error.")
	}
	jsonData, _ := io.ReadAll(file)

	app.SetArgs([]string{"import", "--data", string(jsonData)})
	app.Execute()

	app = domainCmdInit()
	app.SetArgs([]string{"unset", "-t", "limited", "limitedDomain.example.jp"})
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
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	file, err := os.Open("../misc/test/exampleConfig.json")
	if err != nil {
		t.Fatalf("Test resource fetch error.")
	}
	jsonData, _ := io.ReadAll(file)

	app.SetArgs([]string{"import", "--data", string(jsonData)})
	app.Execute()
	RelayState.Load()

	buffer := new(bytes.Buffer)

	app = domainCmdInit()
	app.SetOut(buffer)
	app.SetArgs([]string{"set", "-t", "hoge", "hoge.example.jp"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid type provided" {
		t.Fatalf("Invalid Response.")
	}
}

func TestUnfollowDomain(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	file, err := os.Open("../misc/test/exampleConfig.json")
	if err != nil {
		t.Fatalf("Test resource fetch error.")
	}
	jsonData, _ := io.ReadAll(file)

	app.SetArgs([]string{"import", "--data", string(jsonData)})
	app.Execute()

	app = domainCmdInit()
	app.SetArgs([]string{"unfollow", "subscription.example.jp"})
	app.Execute()
	RelayState.Load()

	valid := true
	for _, domain := range RelayState.Subscribers {
		if domain.Domain == "subscription.example.jp" {
			valid = false
		}
	}

	if !valid {
		t.Fatalf("Not unfollowed domain")
	}
}

func TestInvalidUnfollowDomain(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	file, err := os.Open("../misc/test/exampleConfig.json")
	if err != nil {
		t.Fatalf("Test resource fetch error.")
	}
	jsonData, _ := io.ReadAll(file)

	app.SetArgs([]string{"import", "--data", string(jsonData)})
	app.Execute()
	RelayState.Load()

	buffer := new(bytes.Buffer)

	app = domainCmdInit()
	app.SetOut(buffer)
	app.SetArgs([]string{"unfollow", "unknown.tld"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid domain [unknown.tld] provided" {
		t.Fatalf("Invalid Response.")
	}
}
