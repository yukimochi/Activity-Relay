package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestListFollows(t *testing.T) {
	app := buildNewCmd()

	buffer := new(bytes.Buffer)
	app.SetOutput(buffer)

	relayState.RedisClient.HMSet("relay:pending:example.com", map[string]interface{}{
		"inbox_url":   "https://example.com/inbox",
		"activity_id": "https://example.com/UUID",
		"type":        "Follow",
		"actor":       "https://example.com/user/example",
		"object":      "https://" + hostname.Host + "/actor",
	})

	app.SetArgs([]string{"follow", "list"})
	app.Execute()

	output := buffer.String()
	valid := ` - Follow request :
example.com
Total : 1
`
	if output != valid {
		t.Fatalf("Invalid Responce.")
	}

	relayState.RedisClient.FlushAll().Result()
	relayState.Load()
}

func TestAcceptFollow(t *testing.T) {
	app := buildNewCmd()

	relayState.RedisClient.HMSet("relay:pending:example.com", map[string]interface{}{
		"inbox_url":   "https://example.com/inbox",
		"activity_id": "https://example.com/UUID",
		"type":        "Follow",
		"actor":       "https://example.com/user/example",
		"object":      "https://" + hostname.Host + "/actor",
	})

	app.SetArgs([]string{"follow", "accept", "example.com"})
	app.Execute()

	valid, _ := relayState.RedisClient.Exists("relay:pending:example.com").Result()
	if valid != 0 {
		t.Fatalf("Not removed follow request.")
	}

	valid, _ = relayState.RedisClient.Exists("relay:subscription:example.com").Result()
	if valid != 1 {
		t.Fatalf("Not created subscription.")
	}

	relayState.RedisClient.FlushAll().Result()
	relayState.Load()
}

func TestRejectFollow(t *testing.T) {
	app := buildNewCmd()

	relayState.RedisClient.HMSet("relay:pending:example.com", map[string]interface{}{
		"inbox_url":   "https://example.com/inbox",
		"activity_id": "https://example.com/UUID",
		"type":        "Follow",
		"actor":       "https://example.com/user/example",
		"object":      "https://" + hostname.Host + "/actor",
	})

	app.SetArgs([]string{"follow", "reject", "example.com"})
	app.Execute()

	valid, _ := relayState.RedisClient.Exists("relay:pending:example.com").Result()
	if valid != 0 {
		t.Fatalf("No response follow request.")
	}

	valid, _ = relayState.RedisClient.Exists("relay:subscription:example.com").Result()
	if valid != 0 {
		t.Fatalf("Created subscription.")
	}

	relayState.RedisClient.FlushAll().Result()
	relayState.Load()
}

func TestInvalidFollow(t *testing.T) {
	app := buildNewCmd()

	buffer := new(bytes.Buffer)
	app.SetOutput(buffer)

	app.SetArgs([]string{"follow", "accept", "unknown.tld"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid domain given" {
		t.Fatalf("Invalid Responce.")
	}

	relayState.RedisClient.FlushAll().Result()
	relayState.Load()
}
