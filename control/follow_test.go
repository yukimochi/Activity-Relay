package control

import (
	"bytes"
	"strings"
	"testing"
)

func TestListFollows(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := followCmdInit()

	buffer := new(bytes.Buffer)
	app.SetOutput(buffer)

	relayState.RedisClient.HMSet("relay:pending:example.com", map[string]interface{}{
		"inbox_url":   "https://example.com/inbox",
		"activity_id": "https://example.com/UUID",
		"type":        "Follow",
		"actor":       "https://example.com/user/example",
		"object":      "https://" + globalConfig.ServerHostname().Host + "/actor",
	})

	app.SetArgs([]string{"list"})
	app.Execute()

	output := buffer.String()
	valid := ` - Follow request :
example.com
Total : 1
`
	if output != valid {
		t.Fatalf("Invalid Response.")
	}
}

func TestAcceptFollow(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := followCmdInit()

	relayState.RedisClient.HMSet("relay:pending:example.com", map[string]interface{}{
		"inbox_url":   "https://example.com/inbox",
		"activity_id": "https://example.com/UUID",
		"type":        "Follow",
		"actor":       "https://example.com/user/example",
		"object":      "https://" + globalConfig.ServerHostname().Host + "/actor",
	})

	app.SetArgs([]string{"accept", "example.com"})
	app.Execute()

	valid, _ := relayState.RedisClient.Exists("relay:pending:example.com").Result()
	if valid != 0 {
		t.Fatalf("Not removed follow request.")
	}

	valid, _ = relayState.RedisClient.Exists("relay:subscription:example.com").Result()
	if valid != 1 {
		t.Fatalf("Not created subscription.")
	}
}

func TestRejectFollow(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := followCmdInit()

	relayState.RedisClient.HMSet("relay:pending:example.com", map[string]interface{}{
		"inbox_url":   "https://example.com/inbox",
		"activity_id": "https://example.com/UUID",
		"type":        "Follow",
		"actor":       "https://example.com/user/example",
		"object":      "https://" + globalConfig.ServerHostname().Host + "/actor",
	})

	app.SetArgs([]string{"reject", "example.com"})
	app.Execute()

	valid, _ := relayState.RedisClient.Exists("relay:pending:example.com").Result()
	if valid != 0 {
		t.Fatalf("No response follow request.")
	}

	valid, _ = relayState.RedisClient.Exists("relay:subscription:example.com").Result()
	if valid != 0 {
		t.Fatalf("Created subscription.")
	}
}

func TestInvalidFollow(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := followCmdInit()

	buffer := new(bytes.Buffer)
	app.SetOutput(buffer)

	app.SetArgs([]string{"accept", "unknown.tld"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid domain [unknown.tld] given" {
		t.Fatalf("Invalid Response.")
	}
}

func TestInvalidRejectFollow(t *testing.T) {
	relayState.RedisClient.FlushAll().Result()

	app := followCmdInit()

	buffer := new(bytes.Buffer)
	app.SetOutput(buffer)

	app.SetArgs([]string{"reject", "unknown.tld"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid domain [unknown.tld] given" {
		t.Fatalf("Invalid Response.")
	}
}

func TestCreateUpdateActorActivity(t *testing.T) {
	app := configCmdInit()

	app.SetArgs([]string{"import", "--json", "../misc/exampleConfig.json"})
	app.Execute()

	app = followCmdInit()

	app.SetArgs([]string{"update"})
	app.Execute()
}
