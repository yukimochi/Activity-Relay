package deliver

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/yukimochi/Activity-Relay/models"
)

func TestMain(m *testing.M) {
	var err error

	testConfigPath := "../misc/config.yml"
	file, _ := os.Open(testConfigPath)
	defer file.Close()

	viper.SetConfigType("yaml")
	viper.ReadConfig(file)
	viper.Set("ACTOR_PEM", "../misc/testKey.pem")
	viper.BindEnv("REDIS_URL")

	globalConfig, err = models.NewRelayConfig()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = initialize(globalConfig)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	redisClient.FlushAll().Result()
	code := m.Run()
	os.Exit(code)
}

func TestRelayActivity(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadAll(r.Body)
		if string(data) != "data" || r.Header.Get("Content-Type") != "application/activity+json" {
			w.WriteHeader(500)
			w.Write(nil)
		} else {
			w.WriteHeader(202)
			w.Write(nil)
		}
	}))
	defer s.Close()

	err := relayActivity(s.URL, "data")
	if err != nil {
		t.Fatal("Failed - Data transfar not collect")
	}
}

func TestRelayActivityNoHost(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	defer s.Close()

	err := relayActivity("http://nohost.example.jp", "data")
	if err == nil {
		t.Fatal("Failed - Error not reported.")
	}
	domain, _ := url.Parse("http://nohost.example.jp")
	data, _ := redisClient.HGet("relay:statistics:"+domain.Host, "last_error").Result()
	if data == "" {
		t.Fatal("Failed - Error not cached.")
	}
}

func TestRelayActivityResp500(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write(nil)
	}))
	defer s.Close()

	err := relayActivity(s.URL, "data")
	if err == nil {
		t.Fatal("Failed - Error not reported.")
	}
	domain, _ := url.Parse(s.URL)
	data, _ := redisClient.HGet("relay:statistics:"+domain.Host, "last_error").Result()
	if data == "" {
		t.Fatal("Failed - Error not cached.")
	}
}

func TestRegistorActivity(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadAll(r.Body)
		if string(data) != "data" || r.Header.Get("Content-Type") != "application/activity+json" {
			w.WriteHeader(500)
			w.Write(nil)
		} else {
			w.WriteHeader(202)
			w.Write(nil)
		}
	}))
	defer s.Close()

	err := registorActivity(s.URL, "data")
	if err != nil {
		t.Fatal("Failed - Data transfar not collect")
	}
}

func TestRegistorActivityNoHost(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	defer s.Close()

	err := registorActivity("http://nohost.example.jp", "data")
	if err == nil {
		t.Fatal("Failed - Error not reported.")
	}
}

func TestRegistorActivityResp500(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write(nil)
	}))
	defer s.Close()

	err := registorActivity(s.URL, "data")
	if err == nil {
		t.Fatal("Failed - Error not reported.")
	}
}
