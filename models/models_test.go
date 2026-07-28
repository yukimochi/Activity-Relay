package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/viper"
)

var globalConfig *RelayConfig
var relayState RelayState
var ch chan bool

func TestMain(m *testing.M) {
	var err error

	testConfigPath := "../misc/test/config.yml"
	file, _ := os.Open(testConfigPath)
	defer file.Close()

	viper.SetConfigType("yaml")
	viper.ReadConfig(file)
	viper.Set("ACTOR_PEM", "../misc/test/testKey.pem")
	viper.BindEnv("REDIS_URL")

	globalConfig, err = NewRelayConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	relayState = NewState(globalConfig.RedisClient(), true)
	ch = make(chan bool)
	relayState.ListenNotify(ch)
	relayState.RedisClient.FlushAll(context.TODO()).Result()
	code := m.Run()
	os.Exit(code)
}

func TestActivityUnmarshalToAndCc(t *testing.T) {
	tests := []struct {
		name   string
		json   string
		wantTo []string
		wantCc []string
	}{
		{
			name:   "to as string",
			json:   `{"type":"Follow","actor":"https://gts.example/users/host","object":"https://relay.example/actor","to":"https://relay.example/actor"}`,
			wantTo: []string{"https://relay.example/actor"},
			wantCc: nil,
		},
		{
			name:   "to as array",
			json:   `{"type":"Follow","actor":"https://gts.example/users/host","object":"https://relay.example/actor","to":["https://relay.example/actor"]}`,
			wantTo: []string{"https://relay.example/actor"},
			wantCc: nil,
		},
		{
			name:   "to and cc as arrays",
			json:   `{"type":"Create","actor":"https://gts.example/users/host","to":["https://www.w3.org/ns/activitystreams#Public"],"cc":["https://relay.example/actor"]}`,
			wantTo: []string{"https://www.w3.org/ns/activitystreams#Public"},
			wantCc: []string{"https://relay.example/actor"},
		},
		{
			name:   "to and cc as strings",
			json:   `{"type":"Create","actor":"https://gts.example/users/host","to":"https://www.w3.org/ns/activitystreams#Public","cc":"https://relay.example/actor"}`,
			wantTo: []string{"https://www.w3.org/ns/activitystreams#Public"},
			wantCc: []string{"https://relay.example/actor"},
		},
		{
			name:   "to and cc absent",
			json:   `{"type":"Follow","actor":"https://gts.example/users/host","object":"https://relay.example/actor"}`,
			wantTo: nil,
			wantCc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var a Activity
			if err := json.Unmarshal([]byte(tt.json), &a); err != nil {
				t.Fatalf("UnmarshalJSON error: %v", err)
			}
			if !sliceEqual(a.To, tt.wantTo) {
				t.Errorf("To = %v, want %v", a.To, tt.wantTo)
			}
			if !sliceEqual(a.Cc, tt.wantCc) {
				t.Errorf("Cc = %v, want %v", a.Cc, tt.wantCc)
			}
		})
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
