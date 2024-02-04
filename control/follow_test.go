package control

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestListFollows(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := followCmdInit()

	buffer := new(bytes.Buffer)
	app.SetOut(buffer)

	RelayState.RedisClient.HMSet(context.TODO(), "relay:pending:example.com", map[string]interface{}{
		"inbox_url":   "https://example.com/inbox",
		"activity_id": "https://example.com/UUID",
		"type":        "Follow",
		"actor":       "https://example.com/user/example",
		"object":      "https://" + GlobalConfig.ServerHostname().Host + "/actor",
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

func TestAcceptSubscribe(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := followCmdInit()

	RelayState.RedisClient.HMSet(context.TODO(), "relay:pending:example.com", map[string]interface{}{
		"inbox_url":   "https://example.com/inbox",
		"activity_id": "https://example.com/UUID",
		"type":        "Follow",
		"actor":       "https://example.com/user/example",
		"object":      "https://www.w3.org/ns/activitystreams#Public",
	})

	app.SetArgs([]string{"accept", "example.com"})
	app.Execute()

	valid, _ := RelayState.RedisClient.Exists(context.TODO(), "relay:pending:example.com").Result()
	if valid != 0 {
		t.Fatalf("Not removed follow request.")
	}

	valid, _ = RelayState.RedisClient.Exists(context.TODO(), "relay:subscription:example.com").Result()
	if valid != 1 {
		t.Fatalf("Not created subscriber.")
	}
}

func TestAcceptFollow(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := followCmdInit()

	RelayState.RedisClient.HMSet(context.TODO(), "relay:pending:example.com", map[string]interface{}{
		"inbox_url":   "https://example.com/inbox",
		"activity_id": "https://example.com/UUID",
		"type":        "Follow",
		"actor":       "https://example.com/user/example",
		"object":      "https://" + GlobalConfig.ServerHostname().Host + "/actor",
	})

	app.SetArgs([]string{"accept", "example.com"})
	app.Execute()

	valid, _ := RelayState.RedisClient.Exists(context.TODO(), "relay:pending:example.com").Result()
	if valid != 0 {
		t.Fatalf("Not removed follow request.")
	}

	valid, _ = RelayState.RedisClient.Exists(context.TODO(), "relay:follower:example.com").Result()
	if valid != 1 {
		t.Fatalf("Not created follower.")
	}
}

func TestRejectFollow(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := followCmdInit()

	RelayState.RedisClient.HMSet(context.TODO(), "relay:pending:example.com", map[string]interface{}{
		"inbox_url":   "https://example.com/inbox",
		"activity_id": "https://example.com/UUID",
		"type":        "Follow",
		"actor":       "https://example.com/user/example",
		"object":      "https://" + GlobalConfig.ServerHostname().Host + "/actor",
	})

	app.SetArgs([]string{"reject", "example.com"})
	app.Execute()

	valid, _ := RelayState.RedisClient.Exists(context.TODO(), "relay:pending:example.com").Result()
	if valid != 0 {
		t.Fatalf("No response follow request.")
	}

	valid, _ = RelayState.RedisClient.Exists(context.TODO(), "relay:subscription:example.com").Result()
	if valid != 0 {
		t.Fatalf("Created subscription.")
	}
}

func TestInvalidFollow(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := followCmdInit()

	buffer := new(bytes.Buffer)
	app.SetOut(buffer)

	app.SetArgs([]string{"accept", "unknown.tld"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid domain [unknown.tld] given" {
		t.Fatalf("Invalid Response.")
	}
}

func TestInvalidRejectFollow(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := followCmdInit()

	buffer := new(bytes.Buffer)
	app.SetOut(buffer)

	app.SetArgs([]string{"reject", "unknown.tld"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid domain [unknown.tld] given" {
		t.Fatalf("Invalid Response.")
	}
}

func TestCreateUpdateActorActivity(t *testing.T) {
	app := configCmdInit()
	file, err := os.Open("../misc/test/exampleConfig.json")
	if err != nil {
		t.Fatalf("Test resource fetch error.")
	}
	jsonData, _ := io.ReadAll(file)

	app.SetArgs([]string{"import", "--data", string(jsonData)})
	app.Execute()

	app = followCmdInit()

	app.SetArgs([]string{"update"})
	app.Execute()
}
