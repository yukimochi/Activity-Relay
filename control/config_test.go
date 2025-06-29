package control

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestPersonOnlyConfiguration(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()

	t.Run("Enable person-only configuration", func(t *testing.T) {
		app.SetArgs([]string{"enable", "person-only"})
		app.Execute()
		RelayState.Load()
		if !RelayState.RelayConfig.PersonOnly {
			t.Fatalf("Expected PersonOnly to be enabled, but it was not")
		}
	})

	t.Run("Disable person-only configuration", func(t *testing.T) {
		app.SetArgs([]string{"disable", "person-only"})
		app.Execute()
		RelayState.Load()
		if RelayState.RelayConfig.PersonOnly {
			t.Fatalf("Expected PersonOnly to be disabled, but it was not")
		}
	})
}

func TestManuallyAcceptConfiguration(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()

	t.Run("Enable manually-accept configuration", func(t *testing.T) {
		app.SetArgs([]string{"enable", "manually-accept"})
		app.Execute()
		RelayState.Load()
		if !RelayState.RelayConfig.ManuallyAccept {
			t.Fatalf("Expected ManuallyAccept to be enabled, but it was not")
		}
	})

	t.Run("Disable manually-accept configuration", func(t *testing.T) {
		app.SetArgs([]string{"disable", "manually-accept"})
		app.Execute()
		RelayState.Load()
		if RelayState.RelayConfig.ManuallyAccept {
			t.Fatalf("Expected ManuallyAccept to be disabled, but it was not")
		}
	})
}

func TestInvalidConfig(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	buffer := new(bytes.Buffer)
	app.SetOut(buffer)

	app.SetArgs([]string{"enable", "hoge"})
	app.Execute()

	output := buffer.String()
	if strings.Split(output, "\n")[0] != "Invalid configuration provided: hoge" {
		t.Fatalf("Expected output to be 'Invalid configuration provided: hoge', but got '%s'", strings.Split(output, "\n")[0])
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
		case "Person-Type Actor limitation":
			if strings.Split(row, ":")[1] == " true" {
				t.Fatalf("Expected 'Person-Type Actor limitation' to be false, but got true")
			}
		case "Manual follow request acceptance":
			if strings.Split(row, ":")[1] == " true" {
				t.Fatalf("Expected 'Manual follow request acceptance' to be false, but got true")
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
		t.Fatalf("Failed to open test resource file: %v", err)
	}
	jsonData, _ := io.ReadAll(file)
	output := buffer.String()
	if strings.Split(output, "\n")[0] != string(jsonData) {
		t.Fatalf("Expected exported config to be '%s', but got '%s'", string(jsonData), strings.Split(output, "\n")[0])
	}
}

func TestImportConfig(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	app := configCmdInit()
	file, err := os.Open("../misc/test/exampleConfig.json")
	if err != nil {
		t.Fatalf("Failed to open test resource file: %v", err)
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
		t.Fatalf("Expected exported config to be '%s', but got '%s'", string(jsonData), strings.Split(output, "\n")[0])
	}
}
