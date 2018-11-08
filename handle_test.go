package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/yukimochi/Activity-Relay/ActivityPub"
	keyloader "github.com/yukimochi/Activity-Relay/KeyLoader"
)

func TestMain(m *testing.M) {
	os.Setenv("ACTOR_PEM", "misc/testKey.pem")
	os.Setenv("RELAY_DOMAIN", "relay.yukimochi.example.org")
	pemPath := os.Getenv("ACTOR_PEM")
	relayDomain := os.Getenv("RELAY_DOMAIN")
	hostkey, _ = keyloader.ReadPrivateKeyRSAfromPath(pemPath)
	hostname, _ = url.Parse("https://" + relayDomain)
	Actor = activitypub.GenerateActor(hostname, &hostkey.PublicKey)
	WebfingerResource = activitypub.GenerateWebfingerResource(hostname, &Actor)

	code := m.Run()
	os.Exit(code)
}

func TestHandleWebfingerGet(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handleWebfinger))
	defer s.Close()

	req, _ := http.NewRequest("GET", s.URL, nil)
	q := req.URL.Query()
	q.Add("resource", "acct:relay@"+os.Getenv("RELAY_DOMAIN"))
	req.URL.RawQuery = q.Encode()
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.Header.Get("Content-Type") != "application/json" {
		t.Fatalf("Failed - Content-Type not match.")
	}
	if r.StatusCode != 200 {
		t.Fatalf("Failed - StatusCode is not 200.")
	}
	defer r.Body.Close()

	data, _ := ioutil.ReadAll(r.Body)
	var wfresource activitypub.WebfingerResource
	err = json.Unmarshal(data, &wfresource)
	if err != nil {
		t.Fatalf("WebfingerResource responce is not valid.")
	}

	domain, _ := url.Parse(wfresource.Links[0].Href)
	if domain.Host != os.Getenv("RELAY_DOMAIN") {
		t.Fatalf("WebfingerResource's Host not valid.")
	}
}
func TestHandleWebfingerGetBadResource(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handleWebfinger))
	defer s.Close()

	req, _ := http.NewRequest("GET", s.URL, nil)
	q := req.URL.Query()
	q.Add("resource", "acct:yukimochi@"+os.Getenv("RELAY_DOMAIN"))
	req.URL.RawQuery = q.Encode()
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 404 {
		t.Fatalf("Failed - StatusCode is not 404.")
	}
}
func TestHandleWebfingerInvalidMethod(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handleWebfinger))
	defer s.Close()

	req, _ := http.NewRequest("POST", s.URL, nil)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 400 {
		t.Fatalf("Failed - StatusCode is not 400.")
	}
}
func TestHandleActorGet(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handleActor))
	defer s.Close()

	r, err := http.Get(s.URL)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.Header.Get("Content-Type") != "application/activity+json" {
		t.Fatalf("Failed - Content-Type not match.")
	}
	if r.StatusCode != 200 {
		t.Fatalf("Failed - StatusCode is not 200.")
	}
	defer r.Body.Close()

	data, _ := ioutil.ReadAll(r.Body)
	var actor activitypub.Actor
	err = json.Unmarshal(data, &actor)
	if err != nil {
		t.Fatalf("Actor responce is not valid.")
	}

	domain, _ := url.Parse(actor.ID)
	if domain.Host != os.Getenv("RELAY_DOMAIN") {
		t.Fatalf("Actor's Host not valid.")
	}
}

func TestHandleActorInvalidMethod(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handleActor))
	defer s.Close()

	r, err := http.Post(s.URL, "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 400 {
		t.Fatalf("Failed - StatusCode is not 400.")
	}
}

func TestContains(t *testing.T) {
	data := "nil"
	sData := []string{
		"no",
		"nil",
	}
	badData := 0
	result := contains(data, "true")
	if result != false {
		t.Fatalf("Failed - no contain but true.")
	}
	result = contains(data, "nil")
	if result != true {
		t.Fatalf("Failed - contain but false.")
	}
	result = contains(sData, "true")
	if result != false {
		t.Fatalf("Failed - no contain but true. (slice)")
	}
	result = contains(sData, "nil")
	if result != true {
		t.Fatalf("Failed - contain but false. (slice)")
	}
	result = contains(badData, "hoge")
	if result != false {
		t.Fatalf("Failed - input bad data but true. (slice)")
	}
}
