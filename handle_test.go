package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/yukimochi/Activity-Relay/ActivityPub"
	"github.com/yukimochi/Activity-Relay/RelayConf"
)

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

func TestHandleInboxNoSignature(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, decodeActivity)
	}))
	defer s.Close()

	req, _ := http.NewRequest("POST", s.URL, nil)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 400 {
		t.Fatalf("Failed - StatusCode is not 400")
	}
}

func TestHandleInboxInvalidMethod(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, decodeActivity)
	}))
	defer s.Close()

	req, _ := http.NewRequest("GET", s.URL, nil)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 404 {
		t.Fatalf("Failed - StatusCode is not 404")
	}
}

func mockActivityDecoderProvider(activity *activitypub.Activity, actor *activitypub.Actor) func(r *http.Request) (*activitypub.Activity, *activitypub.Actor, []byte, error) {
	return func(r *http.Request) (*activitypub.Activity, *activitypub.Actor, []byte, error) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		return activity, actor, body, nil
	}
}

func TestHandleInboxValidFollow(t *testing.T) {
	activity := activitypub.Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		"https://mastodon.test.yukimochi.io/c125e836-e622-478e-a22d-2d9fbf2f496f",
		"https://mastodon.test.yukimochi.io/users/yukimochi",
		"Follow",
		"https://www.w3.org/ns/activitystreams#Public",
		nil,
		nil,
	}
	actor := activitypub.Actor{
		[]string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		"https://mastodon.test.yukimochi.io/users/yukimochi",
		"Person",
		"yukimochi",
		"https://mastodon.test.yukimochi.io/users/yukimochi/inbox",
		&activitypub.Endpoints{
			"https://mastodon.test.yukimochi.io/inbox",
		},
		activitypub.PublicKey{
			"https://mastodon.test.yukimochi.io/users/yukimochi#main-key",
			"https://mastodon.test.yukimochi.io/users/yukimochi",
			"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuak6v+V5hd683ioTLSPF\nLR7CxiI1GMzmOfgaP/P37YBi8bk1aYu3pSDaSJ4889llLHOrLWnzuojHHAUTsVH3\nDG3BXUIjMdGzO6CYG0Tsk36PF7yKZ4RrIj3z03XEUogBbNN/YiqjWCiUkOLLayx5\nM/iE1VBu3zoC2cP8m+hnVdSOpTV8XcaTXMQSGnk/mKMh93CP16pMkJ3Jaw5I2tYm\nCTKVV3zPdmXwT5rCL/qstlIfDaIkKc/PL04mhA9/8+9A6HhhTsxCsgA1zJZomTBI\n4FXeu7mzFZJtZJdDwaVy2H+CKMw6HOHneEenvvCR/37kiLjk8gw+grC/G1Bw6E2h\nZwIDAQAB\n-----END PUBLIC KEY-----\n",
		},
	}
	domain, _ := url.Parse(activity.Actor)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, mockActivityDecoderProvider(&activity, &actor))
	}))
	defer s.Close()

	relayconf.SetConfig(redClient, "manually_accept", false)
	relConfig = relayconf.LoadConfig(redClient)

	req, _ := http.NewRequest("POST", s.URL, nil)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 202 {
		t.Fatalf("Failed - StatusCode is not 202")
	}
	res, _ := redClient.Exists("relay:subscription:" + domain.Host).Result()
	if res != 1 {
		t.Fatalf("Failed - Subscription not works.")
	}
	redClient.Del("relay:subscription:" + domain.Host).Result()
	redClient.Del("relay:pending:" + domain.Host).Result()

	// Switch Manually
	relayconf.SetConfig(redClient, "manually_accept", true)
	relConfig = relayconf.LoadConfig(redClient)

	req, _ = http.NewRequest("POST", s.URL, nil)
	client = new(http.Client)
	r, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 202 {
		t.Fatalf("Failed - StatusCode is not 202")
	}
	res, _ = redClient.Exists("relay:pending:" + domain.Host).Result()
	if res != 1 {
		t.Fatalf("Failed - Pending not works.")
	}
	res, _ = redClient.Exists("relay:subscription:" + domain.Host).Result()
	if res != 0 {
		t.Fatalf("Failed - Pending was skipped.")
	}
	redClient.Del("relay:subscription:" + domain.Host).Result()
	redClient.Del("relay:pending:" + domain.Host).Result()
	relayconf.SetConfig(redClient, "manually_accept", false)
	relConfig = relayconf.LoadConfig(redClient)
}

func TestHandleInboxValidFollowBlocked(t *testing.T) {
	activity := activitypub.Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		"https://mastodon.test.yukimochi.io/c125e836-e622-478e-a22d-2d9fbf2f496f",
		"https://mastodon.test.yukimochi.io/users/yukimochi",
		"Follow",
		"https://www.w3.org/ns/activitystreams#Public",
		nil,
		nil,
	}
	actor := activitypub.Actor{
		[]string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		"https://mastodon.test.yukimochi.io/users/yukimochi",
		"Person",
		"yukimochi",
		"https://mastodon.test.yukimochi.io/users/yukimochi/inbox",
		&activitypub.Endpoints{
			"https://mastodon.test.yukimochi.io/inbox",
		},
		activitypub.PublicKey{
			"https://mastodon.test.yukimochi.io/users/yukimochi#main-key",
			"https://mastodon.test.yukimochi.io/users/yukimochi",
			"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuak6v+V5hd683ioTLSPF\nLR7CxiI1GMzmOfgaP/P37YBi8bk1aYu3pSDaSJ4889llLHOrLWnzuojHHAUTsVH3\nDG3BXUIjMdGzO6CYG0Tsk36PF7yKZ4RrIj3z03XEUogBbNN/YiqjWCiUkOLLayx5\nM/iE1VBu3zoC2cP8m+hnVdSOpTV8XcaTXMQSGnk/mKMh93CP16pMkJ3Jaw5I2tYm\nCTKVV3zPdmXwT5rCL/qstlIfDaIkKc/PL04mhA9/8+9A6HhhTsxCsgA1zJZomTBI\n4FXeu7mzFZJtZJdDwaVy2H+CKMw6HOHneEenvvCR/37kiLjk8gw+grC/G1Bw6E2h\nZwIDAQAB\n-----END PUBLIC KEY-----\n",
		},
	}
	domain, _ := url.Parse(activity.Actor)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, mockActivityDecoderProvider(&activity, &actor))
	}))
	defer s.Close()

	redClient.HSet("relay:config:blockedDomain", domain.Host, "1")

	req, _ := http.NewRequest("POST", s.URL, nil)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 202 {
		t.Fatalf("Failed - StatusCode is not 202")
	}
	res, _ := redClient.Exists("relay:subscription:" + domain.Host).Result()
	if res != 0 {
		t.Fatalf("Failed - Subscription not blocked.")
	}
	redClient.Del("relay:subscription:" + domain.Host).Result()
	redClient.Del("relay:pending:" + domain.Host).Result()
}

func TestHandleInboxValidUnfollow(t *testing.T) {
	activity := activitypub.Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		"https://mastodon.test.yukimochi.io/c125e836-e622-478e-a22d-2d9fbf2f496f",
		"https://mastodon.test.yukimochi.io/users/yukimochi",
		"Undo",
		map[string]interface{}{
			"@context": "https://www.w3.org/ns/activitystreams",
			"id":       "https://mastodon.test.yukimochi.io/c125e836-e622-478e-a22d-2d9fbf2f496f",
			"actor":    "https://mastodon.test.yukimochi.io/users/yukimochi",
			"type":     "Follow",
			"object":   "https://www.w3.org/ns/activitystreams#Public",
		},
		nil,
		nil,
	}
	actor := activitypub.Actor{
		[]string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		"https://mastodon.test.yukimochi.io/users/yukimochi",
		"Person",
		"yukimochi",
		"https://mastodon.test.yukimochi.io/users/yukimochi/inbox",
		&activitypub.Endpoints{
			"https://mastodon.test.yukimochi.io/inbox",
		},
		activitypub.PublicKey{
			"https://mastodon.test.yukimochi.io/users/yukimochi#main-key",
			"https://mastodon.test.yukimochi.io/users/yukimochi",
			"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuak6v+V5hd683ioTLSPF\nLR7CxiI1GMzmOfgaP/P37YBi8bk1aYu3pSDaSJ4889llLHOrLWnzuojHHAUTsVH3\nDG3BXUIjMdGzO6CYG0Tsk36PF7yKZ4RrIj3z03XEUogBbNN/YiqjWCiUkOLLayx5\nM/iE1VBu3zoC2cP8m+hnVdSOpTV8XcaTXMQSGnk/mKMh93CP16pMkJ3Jaw5I2tYm\nCTKVV3zPdmXwT5rCL/qstlIfDaIkKc/PL04mhA9/8+9A6HhhTsxCsgA1zJZomTBI\n4FXeu7mzFZJtZJdDwaVy2H+CKMw6HOHneEenvvCR/37kiLjk8gw+grC/G1Bw6E2h\nZwIDAQAB\n-----END PUBLIC KEY-----\n",
		},
	}
	domain, _ := url.Parse(activity.Actor)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, mockActivityDecoderProvider(&activity, &actor))
	}))
	defer s.Close()

	redClient.HSet("relay:subscription:"+domain.Host, "inbox_url", "https://mastodon.test.yukimochi.io/inbox").Result()

	req, _ := http.NewRequest("POST", s.URL, nil)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 202 {
		t.Fatalf("Failed - StatusCode is not 202")
	}
	res, _ := redClient.Exists("relay:subscription:" + domain.Host).Result()
	if res != 0 {
		t.Fatalf("Failed - Subscription not succeed.")
	}
}
