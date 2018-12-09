package main

import (
	"strings"
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/urfave/cli"
)

func TestServiceBlock(t *testing.T) {
	app := cli.NewApp()
	fooCmd := cli.Command{
		Name:  "service-block",
		Usage: "Enable blocking for service-type actor",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "undo, u",
				Usage: "Undo block",
			},
		},
		Action: serviceBlock,
	}
	app.Commands = []cli.Command{
		fooCmd,
	}

	exportConfig.SetConfig(BlockService, false)
	app.Run([]string{"", "service-block"})
	if !exportConfig.RelayConfig.BlockService {
		t.Fatalf("Not Enabled ServiceBlock feature,")
	}

	app.Run([]string{"", "service-block", "-u"})
	if exportConfig.RelayConfig.BlockService {
		t.Fatalf("Not Disabled ServiceBlock feature,")
	}
}

func TestManuallyAccept(t *testing.T) {
	app := cli.NewApp()
	fooCmd := cli.Command{
		Name:  "manually-accept",
		Usage: "Enable Manually accept follow-request",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "undo, u",
				Usage: "Undo block",
			},
		},
		Action: manuallyAccept,
	}
	app.Commands = []cli.Command{
		fooCmd,
	}

	exportConfig.SetConfig(ManuallyAccept, false)
	app.Run([]string{"", "manually-accept"})
	if !exportConfig.RelayConfig.ManuallyAccept {
		t.Fatalf("Not Enabled Manually accept follow-request feature,")
	}

	app.Run([]string{"", "manually-accept", "-u"})
	if exportConfig.RelayConfig.ManuallyAccept {
		t.Fatalf("Not Disabled Manually accept follow-request feature,")
	}
}

func TestCreateAsAnnounce(t *testing.T) {
	app := cli.NewApp()
	fooCmd := cli.Command{
		Name:  "create-as-announce",
		Usage: "Enable Announce activity instead of relay create activity (Not recommended)",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "undo, u",
				Usage: "Undo block",
			},
		},
		Action: createAsAnnounce,
	}
	app.Commands = []cli.Command{
		fooCmd,
	}

	exportConfig.SetConfig(CreateAsAnnounce, false)
	app.Run([]string{"", "create-as-announce"})
	if !exportConfig.RelayConfig.CreateAsAnnounce {
		t.Fatalf("Not Enabled Announce activity instead of relay create activity feature,")
	}

	app.Run([]string{"", "create-as-announce", "-u"})
	if exportConfig.RelayConfig.CreateAsAnnounce {
		t.Fatalf("Not Disabled Announce activity instead of relay create activity feature,")
	}
}

func TestListConfigs(t *testing.T) {
	app := cli.NewApp()
	fooCmd := cli.Command{
		Name:   "show",
		Usage:  "Show all relay configrations",
		Action: listConfigs,
	}
	app.Commands = []cli.Command{
		fooCmd,
	}

	exportConfig.SetConfig(BlockService, true)
	exportConfig.SetConfig(ManuallyAccept, true)
	exportConfig.SetConfig(CreateAsAnnounce, true)
	out := capturer.CaptureStdout(func() {
		app.Run([]string{"", "show"})
	})

	for _, row := range strings.Split(out, "\n") {
		switch strings.Split(row, ":")[0] {
		case "Blocking for service-type actor ":
			if !(strings.Split(row, ":")[1] == "  true") {
				t.Fatalf(strings.Split(row, ":")[1])
			}
		case "Manually accept follow-request ":
			if !(strings.Split(row, ":")[1] == "  true") {
				t.Fatalf("Invalid Responce.")
			}
		case "Announce activity instead of relay create activity ":
			if !(strings.Split(row, ":")[1] == "  true") {
				t.Fatalf("Invalid Responce.")
			}
		}
	}

	exportConfig.SetConfig(BlockService, false)
	exportConfig.SetConfig(ManuallyAccept, false)
	exportConfig.SetConfig(CreateAsAnnounce, false)
	out = capturer.CaptureStdout(func() {
		app.Run([]string{"", "show"})
	})

	for _, row := range strings.Split(out, "\n") {
		switch strings.Split(row, ":")[0] {
		case "Blocking for service-type actor ":
			if !(strings.Split(row, ":")[1] == "  false") {
				t.Fatalf("Invalid Responce.")
			}
		case "Manually accept follow-request ":
			if !(strings.Split(row, ":")[1] == "  false") {
				t.Fatalf("Invalid Responce.")
			}
		case "Announce activity instead of relay create activity ":
			if !(strings.Split(row, ":")[1] == "  false") {
				t.Fatalf("Invalid Responce.")
			}
		}
	}
}
