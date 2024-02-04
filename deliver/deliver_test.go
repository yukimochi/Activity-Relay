package deliver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"github.com/yukimochi/Activity-Relay/models"
)

func TestMain(m *testing.M) {
	var err error

	testConfigPath := "../misc/test/config.yml"
	file, _ := os.Open(testConfigPath)
	defer file.Close()

	viper.SetConfigType("yaml")
	viper.ReadConfig(file)
	viper.Set("ACTOR_PEM", "../misc/test/testKey.pem")
	viper.BindEnv("REDIS_URL")

	GlobalConfig, err = models.NewRelayConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = initialize(GlobalConfig)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	RedisClient.FlushAll(context.TODO()).Result()
	code := m.Run()
	os.Exit(code)
}

func TestRelayActivity(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		if string(data) != "ExampleData" || r.Header.Get("Content-Type") != "application/activity+json" {
			w.WriteHeader(500)
			w.Write(nil)
		} else {
			w.WriteHeader(202)
			w.Write(nil)
		}
	}))
	defer s.Close()

	activityID := uuid.New()
	remainCount := 1

	pushActivityScript := "redis.call('HSET',KEYS[1], 'body', ARGV[1], 'remain_count', ARGV[2]); redis.call('EXPIRE', KEYS[1], ARGV[3]);"
	RedisClient.Eval(context.TODO(), pushActivityScript, []string{"relay:activity:" + activityID.String()}, "ExampleData", remainCount, 10).Result()

	err := relayActivityV2(s.URL, activityID.String())
	if err != nil {
		t.Fatal(err)
	}
}

func TestRelayActivityNoHost(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	defer s.Close()

	activityID := uuid.New()
	remainCount := 1

	pushActivityScript := "redis.call('HSET',KEYS[1], 'body', ARGV[1], 'remain_count', ARGV[2]); redis.call('EXPIRE', KEYS[1], ARGV[3]);"
	RedisClient.Eval(context.TODO(), pushActivityScript, []string{"relay:activity:" + activityID.String()}, "ExampleData", remainCount, 10).Result()

	err := relayActivityV2("http://nohost.example.jp", activityID.String())
	if err == nil {
		t.Fatal("fail - Error not reported")
	}
	domain, _ := url.Parse("http://nohost.example.jp")
	data, _ := RedisClient.HGet(context.TODO(), "relay:statistics:"+domain.Host, "last_error").Result()
	if data == "" {
		t.Fatal("fail - Error not saved")
	}
}

func TestRelayActivityResp500(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write(nil)
	}))
	defer s.Close()

	activityID := uuid.New()
	remainCount := 1

	pushActivityScript := "redis.call('HSET',KEYS[1], 'body', ARGV[1], 'remain_count', ARGV[2]); redis.call('EXPIRE', KEYS[1], ARGV[3]);"
	RedisClient.Eval(context.TODO(), pushActivityScript, []string{"relay:activity:" + activityID.String()}, "ExampleData", remainCount, 10).Result()

	err := relayActivityV2(s.URL, activityID.String())
	if err == nil {
		t.Fatal("fail - Error not reported")
	}
	domain, _ := url.Parse(s.URL)
	data, _ := RedisClient.HGet(context.TODO(), "relay:statistics:"+domain.Host, "last_error").Result()
	if data == "" {
		t.Fatal("fail - Error not saved")
	}
}

func TestRegisterActivity(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		if string(data) != "data" || r.Header.Get("Content-Type") != "application/activity+json" {
			w.WriteHeader(500)
			w.Write(nil)
		} else {
			w.WriteHeader(202)
			w.Write(nil)
		}
	}))
	defer s.Close()

	err := registerActivity(s.URL, "data")
	if err != nil {
		t.Fatal("fail - Data transfer not collect")
	}
}

func TestRegisterActivityNoHost(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	defer s.Close()

	err := registerActivity("http://nohost.example.jp", "data")
	if err == nil {
		t.Fatal("fail - Error not reported.")
	}
}

func TestRegisterActivityResp500(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write(nil)
	}))
	defer s.Close()

	err := registerActivity(s.URL, "data")
	if err == nil {
		t.Fatal("fail - Error not reported")
	}
}
