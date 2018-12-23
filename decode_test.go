package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	state "github.com/yukimochi/Activity-Relay/State"
)

func TestHandleInboxNoSignure(t *testing.T) {
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

func TestHandleInboxWithRemoteActor(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, decodeActivity)
	}))
	defer s.Close()

	relayState.AddSubscription(state.Subscription{
		Domain:   "innocent.yukimochi.io",
		InboxURL: "https://innocent.yukimochi.io/inbox",
	})

	file, _ := os.Open("./misc/create.json")
	body, _ := ioutil.ReadAll(file)
	req, _ := http.NewRequest("POST", s.URL+"/inbox", bytes.NewReader(body))
	req.Host = "relay.01.cloudgarage.yukimochi.io"
	req.Header.Set("date", "Sun, 23 Dec 2018 07:39:37 GMT")
	req.Header.Set("digest", "SHA-256=mxgIzbPwBuNYxmjhQeH0vWeEedQGqR1R7zMwR/XTfX8=")
	req.Header.Set("content-type", "application/activity+json")
	req.Header.Set("signature", `keyId="https://innocent.yukimochi.io/users/YUKIMOCHI#main-key",algorithm="rsa-sha256",headers="(request-target) host date digest content-type",signature="MhxXhL21RVp8VmALER2U/oJlWldJAB2COiU2QmwGopLD2pw1c32gQvg0PaBRHfMBBOsidZuRRnj43Kn488zW2xV3n3DYWcGscSh527/hhRzcpLVX2kBqbf/WeQzJmfJVuOX4SzivVhnnUB8PvlPj5LRHpw4n/ctMTq37strKDl9iZg9rej1op1YFJagDxm3iPzAhnv8lzO4RI9dstt2i/sN5EfjXai97oS7EgI//Kj1wJCRk9Pw1iTsGfPTkbk/aVZwDt7QGGvGDdO0JJjsCqtIyjojoyD9hFY9GzMqvTwVIYJrh54AUHq2i80veybaOBbCFcEaK0RpKoLs101r5Uw=="`)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 202 {
		t.Fatalf("Failed - StatusCode is not 202")
	}

	relayState.DelSubscription("innocent.yukimochi.io")
}

func TestHandleInboxWithNotfoundRemoteActor(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, decodeActivity)
	}))
	defer s.Close()

	relayState.AddSubscription(state.Subscription{
		Domain:   "innocent.yukimochi.io",
		InboxURL: "https://innocent.yukimochi.io/inbox",
	})

	file, _ := os.Open("./misc/create.json")
	body, _ := ioutil.ReadAll(file)
	req, _ := http.NewRequest("POST", s.URL+"/inbox", bytes.NewReader(body))
	req.Host = "relay.01.cloudgarage.yukimochi.io"
	req.Header.Set("date", "Sun, 23 Dec 2018 07:39:37 GMT")
	req.Header.Set("digest", "SHA-256=mxgIzbPwBuNYxmjhQeH0vWeEedQGqR1R7zMwR/XTfX8=")
	req.Header.Set("content-type", "application/activity+json")
	req.Header.Set("signature", `keyId="https://innocent.yukimochi.io/users/admin#main-key",algorithm="rsa-sha256",headers="(request-target) host date digest content-type",signature="MhxXhL21RVp8VmALER2U/oJlWldJAB2COiU2QmwGopLD2pw1c32gQvg0PaBRHfMBBOsidZuRRnj43Kn488zW2xV3n3DYWcGscSh527/hhRzcpLVX2kBqbf/WeQzJmfJVuOX4SzivVhnnUB8PvlPj5LRHpw4n/ctMTq37strKDl9iZg9rej1op1YFJagDxm3iPzAhnv8lzO4RI9dstt2i/sN5EfjXai97oS7EgI//Kj1wJCRk9Pw1iTsGfPTkbk/aVZwDt7QGGvGDdO0JJjsCqtIyjojoyD9hFY9GzMqvTwVIYJrh54AUHq2i80veybaOBbCFcEaK0RpKoLs101r5Uw=="`)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 400 {
		t.Fatalf("Failed - StatusCode is not 400")
	}

	relayState.DelSubscription("innocent.yukimochi.io")
}

func TestHandleInboxInvalidDigestWithRemoteActor(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleInbox(w, r, decodeActivity)
	}))
	defer s.Close()

	relayState.AddSubscription(state.Subscription{
		Domain:   "innocent.yukimochi.io",
		InboxURL: "https://innocent.yukimochi.io/inbox",
	})

	file, _ := os.Open("./misc/create.json")
	body, _ := ioutil.ReadAll(file)
	req, _ := http.NewRequest("POST", s.URL+"/inbox", bytes.NewReader(body))
	req.Host = "relay.01.cloudgarage.yukimochi.io"
	req.Header.Set("date", "Sun, 23 Dec 2018 07:39:37 GMT")
	req.Header.Set("digest", "SHA-256=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	req.Header.Set("content-type", "application/activity+json")
	req.Header.Set("signature", `keyId="https://innocent.yukimochi.io/users/YUKIMOCHI#main-key",algorithm="rsa-sha256",headers="(request-target) host date digest content-type",signature="MhxXhL21RVp8VmALER2U/oJlWldJAB2COiU2QmwGopLD2pw1c32gQvg0PaBRHfMBBOsidZuRRnj43Kn488zW2xV3n3DYWcGscSh527/hhRzcpLVX2kBqbf/WeQzJmfJVuOX4SzivVhnnUB8PvlPj5LRHpw4n/ctMTq37strKDl9iZg9rej1op1YFJagDxm3iPzAhnv8lzO4RI9dstt2i/sN5EfjXai97oS7EgI//Kj1wJCRk9Pw1iTsGfPTkbk/aVZwDt7QGGvGDdO0JJjsCqtIyjojoyD9hFY9GzMqvTwVIYJrh54AUHq2i80veybaOBbCFcEaK0RpKoLs101r5Uw=="`)
	client := new(http.Client)
	r, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	if r.StatusCode != 400 {
		t.Fatalf("Failed - StatusCode is not 400")
	}

	relayState.DelSubscription("innocent.yukimochi.io")
}
