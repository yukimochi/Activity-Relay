package control

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPersonOnly(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()

	app.SetArgs([]string{"enable", "person-only"})
	app.Execute()
	RelayState.Load()
	if !RelayState.RelayConfig.PersonOnly {
		t.Fatalf("Not Enabled Limited for Person-Type Actor")
	}

	app.SetArgs([]string{"disable", "person-only"})
	app.Execute()
	RelayState.Load()
	if RelayState.RelayConfig.PersonOnly {
		t.Fatalf("Not Disabled Limited for Person-Type Actor")
	}
}

func TestManuallyAccept(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()

	app.SetArgs([]string{"enable", "manually-accept"})
	app.Execute()
	RelayState.Load()
	if !RelayState.RelayConfig.ManuallyAccept {
		t.Fatalf("Not Enabled Manually Accept Follow-Request")
	}

	app.SetArgs([]string{"disable", "manually-accept"})
	app.Execute()
	RelayState.Load()
	if RelayState.RelayConfig.ManuallyAccept {
		t.Fatalf("Not Disabled Manually Accept Follow-Request")
	}
}

func TestInvalidConfig(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	buffer := new(bytes.Buffer)
	app.SetOut(buffer)

	app.SetArgs([]string{"enable", "hoge"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid Config Provided." {
		t.Fatalf("Invalid Response.")
	}
}

func TestListConfig(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	buffer := new(bytes.Buffer)
	app.SetOut(buffer)

	app.SetArgs([]string{"list"})
	app.Execute()

	output := buffer.String()
	for _, row := range strings.Split(output, "\n") {
		switch strings.Split(row, ":")[0] {
		case "Blocking for service-type actor ":
			if strings.Split(row, ":")[1] == "  true" {
				t.Fatalf("Invalid Response.")
			}
		case "Manually accept follow-request ":
			if strings.Split(row, ":")[1] == "  true" {
				t.Fatalf("Invalid Response.")
			}
		case "Announce activity instead of relay create activity ":
			if strings.Split(row, ":")[1] == "  true" {
				t.Fatalf("Invalid Response.")
			}
		}
	}
}

func TestExportConfig(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	buffer := new(bytes.Buffer)
	app.SetOut(buffer)

	app.SetArgs([]string{"export"})
	app.Execute()

	file, err := os.Open("../misc/test/blankConfig.json")
	if err != nil {
		t.Fatalf("Test resource fetch error.")
	}
	jsonData, _ := io.ReadAll(file)
	output := buffer.String()
	if strings.Split(output, "\n")[0] != string(jsonData) {
		t.Fatalf("Invalid Response.")
	}
}

func TestImportConfig(t *testing.T) {
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
	app.SetOut(buffer)

	app.SetArgs([]string{"export"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != string(jsonData) {
		t.Fatalf("Invalid Response.")
	}
}
