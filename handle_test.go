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

func mockActivityDecoderProvider(activity *activitypub.Activity, actor *activitypub.Actor) func(r *http.Request) (*activitypub.Activity, *activitypub.Actor, []byte, error) {
	return func(r *http.Request) (*activitypub.Activity, *activitypub.Actor, []byte, error) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		return activity, actor, body, nil
	}
}

func mockActivity(req string) activitypub.Activity {
	switch req {
	case "Follow":
		body := "{\"@context\":\"https://www.w3.org/ns/activitystreams\",\"id\":\"https://mastodon.test.yukimochi.io/c125e836-e622-478e-a22d-2d9fbf2f496f\",\"type\":\"Follow\",\"actor\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"object\":\"https://www.w3.org/ns/activitystreams#Public\"}"
		var activity activitypub.Activity
		json.Unmarshal([]byte(body), &activity)
		return activity
	case "Invalid-Follow":
		body := "{\"@context\":\"https://www.w3.org/ns/activitystreams\",\"id\":\"https://mastodon.test.yukimochi.io/c125e836-e622-478e-a22d-2d9fbf2f496f\",\"type\":\"Follow\",\"actor\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"object\":\"https://" + os.Getenv("RELAY_DOMAIN") + "/users/relay\"}"
		var activity activitypub.Activity
		json.Unmarshal([]byte(body), &activity)
		return activity
	case "Unfollow":
		body := "{\"@context\":\"https://www.w3.org/ns/activitystreams\",\"id\":\"https://mastodon.test.yukimochi.io/c125e836-e622-478e-a22d-2d9fbf2f496f\",\"type\":\"Undo\",\"actor\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"object\":{\"@context\":\"https://www.w3.org/ns/activitystreams\",\"id\":\"https://mastodon.test.yukimochi.io/c125e836-e622-478e-a22d-2d9fbf2f496f\",\"type\":\"Follow\",\"actor\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"object\":\"https://www.w3.org/ns/activitystreams#Public\"}}"
		var activity activitypub.Activity
		json.Unmarshal([]byte(body), &activity)
		return activity
	case "Invalid-Unfollow":
		body := "{\"@context\":\"https://www.w3.org/ns/activitystreams\",\"id\":\"https://mastodon.test.yukimochi.io/c125e836-e622-478e-a22d-2d9fbf2f496f\",\"type\":\"Undo\",\"actor\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"object\":{\"@context\":\"https://www.w3.org/ns/activitystreams\",\"id\":\"https://hacked.test.yukimochi.io/c125e836-e622-478e-a22d-2d9fbf2f496f\",\"type\":\"Follow\",\"actor\":\"https://hacked.test.yukimochi.io/users/yukimochi\",\"object\":\"https://www.w3.org/ns/activitystreams#Public\"}}"
		var activity activitypub.Activity
		json.Unmarshal([]byte(body), &activity)
		return activity
	case "Create":
		body := "{\"@context\":[\"https://www.w3.org/ns/activitystreams\",\"https://w3id.org/security/v1\",{\"manuallyApprovesFollowers\":\"as:manuallyApprovesFollowers\",\"sensitive\":\"as:sensitive\",\"movedTo\":{\"@id\":\"as:movedTo\",\"@type\":\"@id\"},\"Hashtag\":\"as:Hashtag\",\"ostatus\":\"http://ostatus.org#\",\"atomUri\":\"ostatus:atomUri\",\"inReplyToAtomUri\":\"ostatus:inReplyToAtomUri\",\"conversation\":\"ostatus:conversation\",\"toot\":\"http://joinmastodon.org/ns#\",\"Emoji\":\"toot:Emoji\",\"focalPoint\":{\"@container\":\"@list\",\"@id\":\"toot:focalPoint\"},\"featured\":{\"@id\":\"toot:featured\",\"@type\":\"@id\"},\"schema\":\"http://schema.org#\",\"PropertyValue\":\"schema:PropertyValue\",\"value\":\"schema:value\"}],\"id\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075045564444857/activity\",\"type\":\"Create\",\"actor\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"published\":\"2018-11-15T11:07:26Z\",\"to\":[\"https://www.w3.org/ns/activitystreams#Public\"],\"cc\":[\"https://mastodon.test.yukimochi.io/users/yukimochi/followers\"],\"object\":{\"id\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075045564444857\",\"type\":\"Note\",\"summary\":null,\"inReplyTo\":null,\"published\":\"2018-11-15T11:07:26Z\",\"url\":\"https://mastodon.test.yukimochi.io/@yukimochi/101075045564444857\",\"attributedTo\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"to\":[\"https://www.w3.org/ns/activitystreams#Public\"],\"cc\":[\"https://mastodon.test.yukimochi.io/users/yukimochi/followers\"],\"sensitive\":false,\"atomUri\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075045564444857\",\"inReplyToAtomUri\":null,\"conversation\":\"tag:mastodon.test.yukimochi.io,2018-11-15:objectId=68:objectType=Conversation\",\"content\":\"<p>Actvity-Relay</p>\",\"contentMap\":{\"en\":\"<p>Actvity-Relay</p>\"},\"attachment\":[],\"tag\":[]},\"signature\":{\"type\":\"RsaSignature2017\",\"creator\":\"https://mastodon.test.yukimochi.io/users/yukimochi#main-key\",\"created\":\"2018-11-15T11:07:26Z\",\"signatureValue\":\"mMgl2GgVPgb1Kw6a2iDIZc7r0j3ob+Cl9y+QkCxIe6KmnUzb15e60UuhkE5j3rJnoTwRKqOFy1PMkSxlYW6fPG/5DBxW9I4kX+8sw8iH/zpwKKUOnXUJEqfwRrNH2ix33xcs/GkKPdedY6iAPV9vGZ10MSMOdypfYgU9r+UI0sTaaC2iMXH0WPnHQuYAI+Q1JDHIbDX5FH1WlDL6+8fKAicf3spBMxDwPHGPK8W2jmDLWdN2Vz4ffsCtWs5BCuqOKZrtTW0Rdd4HWzo40MnRXvBjv7yNlnnKzokANBqiOLWT7kNfK0+Vtnt6c/bNX64KBro53KR7wL3ZBvPVuv5rdQ==\"}}"
		var activity activitypub.Activity
		json.Unmarshal([]byte(body), &activity)
		return activity
	case "Create-Article":
		body := "{\"@context\":[\"https://www.w3.org/ns/activitystreams\",\"https://w3id.org/security/v1\",{\"manuallyApprovesFollowers\":\"as:manuallyApprovesFollowers\",\"sensitive\":\"as:sensitive\",\"movedTo\":{\"@id\":\"as:movedTo\",\"@type\":\"@id\"},\"Hashtag\":\"as:Hashtag\",\"ostatus\":\"http://ostatus.org#\",\"atomUri\":\"ostatus:atomUri\",\"inReplyToAtomUri\":\"ostatus:inReplyToAtomUri\",\"conversation\":\"ostatus:conversation\",\"toot\":\"http://joinmastodon.org/ns#\",\"Emoji\":\"toot:Emoji\",\"focalPoint\":{\"@container\":\"@list\",\"@id\":\"toot:focalPoint\"},\"featured\":{\"@id\":\"toot:featured\",\"@type\":\"@id\"},\"schema\":\"http://schema.org#\",\"PropertyValue\":\"schema:PropertyValue\",\"value\":\"schema:value\"}],\"id\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075045564444857/activity\",\"type\":\"Create\",\"actor\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"published\":\"2018-11-15T11:07:26Z\",\"to\":[\"https://www.w3.org/ns/activitystreams#Public\"],\"cc\":[\"https://mastodon.test.yukimochi.io/users/yukimochi/followers\"],\"object\":{\"id\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075045564444857\",\"type\":\"Article\",\"summary\":null,\"inReplyTo\":null,\"published\":\"2018-11-15T11:07:26Z\",\"url\":\"https://mastodon.test.yukimochi.io/@yukimochi/101075045564444857\",\"attributedTo\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"to\":[\"https://www.w3.org/ns/activitystreams#Public\"],\"cc\":[\"https://mastodon.test.yukimochi.io/users/yukimochi/followers\"],\"sensitive\":false,\"atomUri\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075045564444857\",\"inReplyToAtomUri\":null,\"conversation\":\"tag:mastodon.test.yukimochi.io,2018-11-15:objectId=68:objectType=Conversation\",\"content\":\"<p>Actvity-Relay</p>\",\"contentMap\":{\"en\":\"<p>Actvity-Relay</p>\"},\"attachment\":[],\"tag\":[]},\"signature\":{\"type\":\"RsaSignature2017\",\"creator\":\"https://mastodon.test.yukimochi.io/users/yukimochi#main-key\",\"created\":\"2018-11-15T11:07:26Z\",\"signatureValue\":\"mMgl2GgVPgb1Kw6a2iDIZc7r0j3ob+Cl9y+QkCxIe6KmnUzb15e60UuhkE5j3rJnoTwRKqOFy1PMkSxlYW6fPG/5DBxW9I4kX+8sw8iH/zpwKKUOnXUJEqfwRrNH2ix33xcs/GkKPdedY6iAPV9vGZ10MSMOdypfYgU9r+UI0sTaaC2iMXH0WPnHQuYAI+Q1JDHIbDX5FH1WlDL6+8fKAicf3spBMxDwPHGPK8W2jmDLWdN2Vz4ffsCtWs5BCuqOKZrtTW0Rdd4HWzo40MnRXvBjv7yNlnnKzokANBqiOLWT7kNfK0+Vtnt6c/bNX64KBro53KR7wL3ZBvPVuv5rdQ==\"}}"
		var activity activitypub.Activity
		json.Unmarshal([]byte(body), &activity)
		return activity
	case "Announce":
		body := "{\"@context\":[\"https://www.w3.org/ns/activitystreams\",\"https://w3id.org/security/v1\",{\"manuallyApprovesFollowers\":\"as:manuallyApprovesFollowers\",\"sensitive\":\"as:sensitive\",\"movedTo\":{\"@id\":\"as:movedTo\",\"@type\":\"@id\"},\"Hashtag\":\"as:Hashtag\",\"ostatus\":\"http://ostatus.org#\",\"atomUri\":\"ostatus:atomUri\",\"inReplyToAtomUri\":\"ostatus:inReplyToAtomUri\",\"conversation\":\"ostatus:conversation\",\"toot\":\"http://joinmastodon.org/ns#\",\"Emoji\":\"toot:Emoji\",\"focalPoint\":{\"@container\":\"@list\",\"@id\":\"toot:focalPoint\"},\"featured\":{\"@id\":\"toot:featured\",\"@type\":\"@id\"},\"schema\":\"http://schema.org#\",\"PropertyValue\":\"schema:PropertyValue\",\"value\":\"schema:value\"}],\"id\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075096728565994/activity\",\"type\":\"Announce\",\"actor\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"published\":\"2018-11-15T11:20:27Z\",\"to\":[\"https://www.w3.org/ns/activitystreams#Public\"],\"cc\":[\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"https://mastodon.test.yukimochi.io/users/yukimochi/followers\"],\"object\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075042939498980\",\"atomUri\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075096728565994/activity\",\"signature\":{\"type\":\"RsaSignature2017\",\"creator\":\"https://mastodon.test.yukimochi.io/users/yukimochi#main-key\",\"created\":\"2018-11-15T11:20:27Z\",\"signatureValue\":\"HUe7M49uoEE5bsCM7rrG1ruamKVuYclKYst4OHQHBcvGSWkMTYCG5OmNQMihpFAantN1Mhz+PWKubXsWmrEnUGDNtog9XDGo2iVbYDcD1wjrDz6EuJiq3CBjLpzQ+F04EIx8LK8WSq6pec+jaIxBJghBa7BNH5i77nUdD7QLZxglqljMkf/r2s1i1eDtVJVDLzU3PW05Qu6Z+RDGZrG137ZwLZ3a5hnFyUPqw3fSgdA4n+AmxYenIHorgj45bmI4QJB8X1TPuAadB2XDvnSTTSuJQyDPyR3kCafBWmXDrqb0MRREsc99KzS9L00OiOY31v0TXr78vjSDxoGzEE81cw==\"}}"
		var activity activitypub.Activity
		json.Unmarshal([]byte(body), &activity)
		return activity
	case "Undo":
		body := "{\"@context\":[\"https://www.w3.org/ns/activitystreams\",\"https://w3id.org/security/v1\",{\"manuallyApprovesFollowers\":\"as:manuallyApprovesFollowers\",\"sensitive\":\"as:sensitive\",\"movedTo\":{\"@id\":\"as:movedTo\",\"@type\":\"@id\"},\"Hashtag\":\"as:Hashtag\",\"ostatus\":\"http://ostatus.org#\",\"atomUri\":\"ostatus:atomUri\",\"inReplyToAtomUri\":\"ostatus:inReplyToAtomUri\",\"conversation\":\"ostatus:conversation\",\"toot\":\"http://joinmastodon.org/ns#\",\"Emoji\":\"toot:Emoji\",\"focalPoint\":{\"@container\":\"@list\",\"@id\":\"toot:focalPoint\"},\"featured\":{\"@id\":\"toot:featured\",\"@type\":\"@id\"},\"schema\":\"http://schema.org#\",\"PropertyValue\":\"schema:PropertyValue\",\"value\":\"schema:value\"}],\"id\":\"https://mastodon.test.yukimochi.io/users/yukimochi#announces/101075096728565994/undo\",\"type\":\"Undo\",\"actor\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"to\":[\"https://www.w3.org/ns/activitystreams#Public\"],\"object\":{\"id\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075096728565994/activity\",\"type\":\"Announce\",\"actor\":\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"published\":\"2018-11-15T11:20:27Z\",\"to\":[\"https://www.w3.org/ns/activitystreams#Public\"],\"cc\":[\"https://mastodon.test.yukimochi.io/users/yukimochi\",\"https://mastodon.test.yukimochi.io/users/yukimochi/followers\"],\"object\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075042939498980\",\"atomUri\":\"https://mastodon.test.yukimochi.io/users/yukimochi/statuses/101075096728565994/activity\"},\"signature\":{\"type\":\"RsaSignature2017\",\"creator\":\"https://mastodon.test.yukimochi.io/users/yukimochi#main-key\",\"created\":\"2018-11-15T11:20:53Z\",\"signatureValue\":\"iJOPQXBfCSMZKEvj7dBKiwLa2Uo7u8tXYUkQVcVBaMqc1vZWWJ2x6N6H1HvUYgIIDhaPUT+ivzyhG+hi1YJzqH2zviy75EiI4O55Vl2+EjKntzasWSYB3jGb10Dzxdn0P9Ei24Z+72BKNhkLMGKWNtayA9qA1pr0gmfd5WsZ73jrlY4HGyoADsom9xRH8NYqEugv7mQr/qnkpaVB9276kMEPQRE/V0pNSX29sDOHhBmQmdLRk/WROMPn67/PDly8Oyd9kZKetFnPU9OTOXvmL4LL0xInhOLeL9WgZE3bbE0RWUFlEWC+CJPoDg3Ra2wJd2Nb12esHDJV4LN7gXUG+A==\"}}"
		var activity activitypub.Activity
		json.Unmarshal([]byte(body), &activity)
		return activity
	default:
		panic("No assined request.")
	}
}

func mockActor(req string) activitypub.Actor {
	switch req {
	case "Person":
		return activitypub.Actor{
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
	case "Service":
		return activitypub.Actor{
			[]string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
			"https://mastodon.test.yukimochi.io/users/yukimochi",
			"Service",
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
	default:
		panic("No assined request.")
	}
}

func TestSuitableRelayNoBlockService(t *testing.T) {
	activity := mockActivity("Create")
	personActor := mockActor("Person")
	serviceActor := mockActor("Service")

	relayconf.SetConfig(redClient, "block_service", false)
	relConfig = relayconf.LoadConfig(redClient)

	if suitableRelay(&activity, &personActor) != true {
		t.Fatalf("Failed - Person status not relay")
	}
	if suitableRelay(&activity, &serviceActor) != true {
		t.Fatalf("Failed - Service status not relay")
	}
}

func TestSuitableRelayBlockService(t *testing.T) {
	activity := mockActivity("Create")
	personActor := mockActor("Person")
	serviceActor := mockActor("Service")

	relayconf.SetConfig(redClient, "block_service", true)
	relConfig = relayconf.LoadConfig(redClient)

	if suitableRelay(&activity, &personActor) != true {
		t.Fatalf("Failed - Person status not relay")
	}
	if suitableRelay(&activity, &serviceActor) != false {
		t.Fatalf("Failed - Service status may relay when blocking mode")
	}
	relayconf.SetConfig(redClient, "block_service", false)
	relConfig = relayconf.LoadConfig(redClient)
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

func TestHandleInboxValidFollow(t *testing.T) {
	activity := mockActivity("Follow")
	actor := mockActor("Person")
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
}

func TestHandleInboxValidManuallyFollow(t *testing.T) {
	activity := mockActivity("Follow")
	actor := mockActor("Person")
	domain, _ := url.Parse(activity.Actor)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, mockActivityDecoderProvider(&activity, &actor))
	}))
	defer s.Close()

	// Switch Manually
	relayconf.SetConfig(redClient, "manually_accept", true)
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
	res, _ := redClient.Exists("relay:pending:" + domain.Host).Result()
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

func TestHandleInboxInvalidFollow(t *testing.T) {
	activity := mockActivity("Invalid-Follow")
	actor := mockActor("Person")
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
	if r.StatusCode != 400 {
		t.Fatalf("Failed - StatusCode is not 400")
	}
	res, _ := redClient.Exists("relay:subscription:" + domain.Host).Result()
	if res != 0 {
		t.Fatalf("Failed - Subscription not blocked.")
	}
}

func TestHandleInboxValidFollowBlocked(t *testing.T) {
	activity := mockActivity("Follow")
	actor := mockActor("Person")
	domain, _ := url.Parse(activity.Actor)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, mockActivityDecoderProvider(&activity, &actor))
	}))
	defer s.Close()

	redClient.HSet("relay:config:blockedDomain", domain.Host, "1").Result()

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
	redClient.Del("relay:config:blockedDomain", domain.Host).Result()
}

func TestHandleInboxValidUnfollow(t *testing.T) {
	activity := mockActivity("Unfollow")
	actor := mockActor("Person")
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
	redClient.Del("relay:subscription:" + domain.Host).Result()
}

func TestHandleInboxInvalidUnfollow(t *testing.T) {
	activity := mockActivity("Invalid-Unfollow")
	actor := mockActor("Person")
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
	if r.StatusCode != 400 {
		t.Fatalf("Failed - StatusCode is not 400")
	}
	res, _ := redClient.Exists("relay:subscription:" + domain.Host).Result()
	if res != 1 {
		t.Fatalf("Failed - Block hacked unfollow not succeed.")
	}
	redClient.Del("relay:subscription:" + domain.Host).Result()
}

func TestHandleInboxValidCreate(t *testing.T) {
	activity := mockActivity("Create")
	actor := mockActor("Person")
	domain, _ := url.Parse(activity.Actor)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, mockActivityDecoderProvider(&activity, &actor))
	}))
	defer s.Close()

	redClient.HSet("relay:subscription:"+domain.Host, "inbox_url", "https://mastodon.test.yukimochi.io/inbox").Result()
	redClient.HSet("relay:subscription:example.org", "inbox_url", "https://example.org/inbox").Result()

	req, _ := http.NewRequest("POST", s.URL, nil)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 202 {
		t.Fatalf("Failed - StatusCode is not 202")
	}
	redClient.Del("relay:subscription:" + domain.Host).Result()
	redClient.Del("relay:subscription:example.org").Result()
}

func TestHandleInboxlimitedCreate(t *testing.T) {
	activity := mockActivity("Create")
	actor := mockActor("Person")
	domain, _ := url.Parse(activity.Actor)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, mockActivityDecoderProvider(&activity, &actor))
	}))
	defer s.Close()

	redClient.HSet("relay:subscription:"+domain.Host, "inbox_url", "https://mastodon.test.yukimochi.io/inbox").Result()
	redClient.HSet("relay:config:limitedDomain", domain.Host, "1").Result()

	req, _ := http.NewRequest("POST", s.URL, nil)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 202 {
		t.Fatalf("Failed - StatusCode is not 202")
	}
	redClient.Del("relay:subscription:" + domain.Host).Result()
	redClient.Del("relay:config:limitedDomain", domain.Host).Result()
}

func TestHandleInboxValidCreateAsAnnounceNote(t *testing.T) {
	activity := mockActivity("Create")
	actor := mockActor("Person")
	domain, _ := url.Parse(activity.Actor)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, mockActivityDecoderProvider(&activity, &actor))
	}))
	defer s.Close()

	redClient.HSet("relay:subscription:"+domain.Host, "inbox_url", "https://mastodon.test.yukimochi.io/inbox").Result()
	redClient.HSet("relay:subscription:example.org", "inbox_url", "https://example.org/inbox").Result()
	redClient.HSet("relay:config", "create_as_announce", "1").Result()
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
	redClient.Del("relay:subscription:" + domain.Host).Result()
	redClient.Del("relay:subscription:example.org").Result()
	redClient.HSet("relay:config", "create_as_announce", "0").Result()
	relConfig = relayconf.LoadConfig(redClient)
}

func TestHandleInboxValidCreateAsAnnounceNoNote(t *testing.T) {
	activity := mockActivity("Create-Article")
	actor := mockActor("Person")
	domain, _ := url.Parse(activity.Actor)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, mockActivityDecoderProvider(&activity, &actor))
	}))
	defer s.Close()

	redClient.HSet("relay:subscription:"+domain.Host, "inbox_url", "https://mastodon.test.yukimochi.io/inbox").Result()
	redClient.HSet("relay:subscription:example.org", "inbox_url", "https://example.org/inbox").Result()
	redClient.HSet("relay:config", "create_as_announce", "1").Result()
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
	redClient.Del("relay:subscription:" + domain.Host).Result()
	redClient.Del("relay:subscription:example.org").Result()
	redClient.HSet("relay:config", "create_as_announce", "0").Result()
	relConfig = relayconf.LoadConfig(redClient)
}

func TestHandleInboxUnsubscriptionCreate(t *testing.T) {
	activity := mockActivity("Create")
	actor := mockActor("Person")
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, mockActivityDecoderProvider(&activity, &actor))
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

func TestHandleInboxUndo(t *testing.T) {
	activity := mockActivity("Undo")
	actor := mockActor("Person")
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
	if res != 1 {
		t.Fatalf("Failed - Missing unsubscribed.")
	}
	redClient.Del("relay:subscription:" + domain.Host).Result()
}
