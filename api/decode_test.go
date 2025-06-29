package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/yukimochi/Activity-Relay/models"
)

func TestDecodeActivity(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	RelayState.AddSubscriber(models.Subscriber{
		Domain:   "innocent.yukimochi.io",
		InboxURL: "https://innocent.yukimochi.io/inbox",
	})

	file, _ := os.Open("../misc/test/create.json")
	body, _ := io.ReadAll(file)
	length := strconv.Itoa(len(body))
	req, _ := http.NewRequest("POST", "/inbox", bytes.NewReader(body))
	req.Host = "relay.01.cloudgarage.yukimochi.io"
	req.Header.Add("content-length", length)
	req.Header.Add("content-type", "application/activity+json")
	req.Header.Add("date", "Sun, 23 Dec 2018 07:39:37 GMT")
	req.Header.Add("digest", "SHA-256=mxgIzbPwBuNYxmjhQeH0vWeEedQGqR1R7zMwR/XTfX8=")
	req.Header.Add("signature", `keyId="https://innocent.yukimochi.io/users/YUKIMOCHI#main-key",algorithm="rsa-sha256",headers="(request-target) host date digest content-type",signature="MhxXhL21RVp8VmALER2U/oJlWldJAB2COiU2QmwGopLD2pw1c32gQvg0PaBRHfMBBOsidZuRRnj43Kn488zW2xV3n3DYWcGscSh527/hhRzcpLVX2kBqbf/WeQzJmfJVuOX4SzivVhnnUB8PvlPj5LRHpw4n/ctMTq37strKDl9iZg9rej1op1YFJagDxm3iPzAhnv8lzO4RI9dstt2i/sN5EfjXai97oS7EgI//Kj1wJCRk9Pw1iTsGfPTkbk/aVZwDt7QGGvGDdO0JJjsCqtIyjojoyD9hFY9GzMqvTwVIYJrh54AUHq2i80veybaOBbCFcEaK0RpKoLs101r5Uw=="`)

	activity, actor, _, err := decodeActivity(req)
	if err != nil {
		t.Fatalf("Expected decodeActivity to succeed, but got error: %v", err)
	}

	if activity.Actor != actor.ID {
		t.Fatalf("Expected activity.Actor to be '%s', but got '%s'", actor.ID, activity.Actor)
	}
}

func TestDecodeActivityWithNoSignature(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	RelayState.AddSubscriber(models.Subscriber{
		Domain:   "innocent.yukimochi.io",
		InboxURL: "https://innocent.yukimochi.io/inbox",
	})

	file, _ := os.Open("../misc/test/create.json")
	body, _ := io.ReadAll(file)
	length := strconv.Itoa(len(body))
	req, _ := http.NewRequest("POST", "/inbox", bytes.NewReader(body))
	req.Host = "relay.01.cloudgarage.yukimochi.io"
	req.Header.Add("content-length", length)
	req.Header.Add("content-type", "application/activity+json")
	req.Header.Add("date", "Sun, 23 Dec 2018 07:39:37 GMT")
	req.Header.Add("digest", "SHA-256=mxgIzbPwBuNYxmjhQeH0vWeEedQGqR1R7zMwR/XTfX8=")

	_, _, _, err := decodeActivity(req)
	if err == nil || err.Error() != "neither \"Signature\" nor \"Authorization\" have signature parameters" {
		t.Fatalf("Expected error 'neither \"Signature\" nor \"Authorization\" have signature parameters', but got '%v'", err)
	}
}

func TestDecodeActivityWithNotFoundKeyId(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	RelayState.AddSubscriber(models.Subscriber{
		Domain:   "innocent.yukimochi.io",
		InboxURL: "https://innocent.yukimochi.io/inbox",
	})

	file, _ := os.Open("../misc/test/create.json")
	body, _ := io.ReadAll(file)
	length := strconv.Itoa(len(body))
	req, _ := http.NewRequest("POST", "/inbox", bytes.NewReader(body))
	req.Host = "relay.01.cloudgarage.yukimochi.io"
	req.Header.Add("content-length", length)
	req.Header.Add("content-type", "application/activity+json")
	req.Header.Add("date", "Sun, 23 Dec 2018 07:39:37 GMT")
	req.Header.Add("digest", "SHA-256=mxgIzbPwBuNYxmjhQeH0vWeEedQGqR1R7zMwR/XTfX8=")
	req.Header.Add("signature", `keyId="https://innocent.yukimochi.io/users/admin#main-key",algorithm="rsa-sha256",headers="(request-target) host date digest content-type",signature="MhxXhL21RVp8VmALER2U/oJlWldJAB2COiU2QmwGopLD2pw1c32gQvg0PaBRHfMBBOsidZuRRnj43Kn488zW2xV3n3DYWcGscSh527/hhRzcpLVX2kBqbf/WeQzJmfJVuOX4SzivVhnnUB8PvlPj5LRHpw4n/ctMTq37strKDl9iZg9rej1op1YFJagDxm3iPzAhnv8lzO4RI9dstt2i/sN5EfjXai97oS7EgI//Kj1wJCRk9Pw1iTsGfPTkbk/aVZwDt7QGGvGDdO0JJjsCqtIyjojoyD9hFY9GzMqvTwVIYJrh54AUHq2i80veybaOBbCFcEaK0RpKoLs101r5Uw=="`)

	_, _, _, err := decodeActivity(req)
	if err == nil || err.Error() != "404 Not Found" {
		t.Fatalf("Expected error '404 Not Found', but got '%v'", err)
	}
}

func TestDecodeActivityWithInvalidDigest(t *testing.T) {
	RelayState.RedisClient.FlushAll(context.TODO()).Result()

	RelayState.AddSubscriber(models.Subscriber{
		Domain:   "innocent.yukimochi.io",
		InboxURL: "https://innocent.yukimochi.io/inbox",
	})

	file, _ := os.Open("../misc/test/create.json")
	body, _ := io.ReadAll(file)
	length := strconv.Itoa(len(body))
	req, _ := http.NewRequest("POST", "/inbox", bytes.NewReader(body))
	req.Host = "relay.01.cloudgarage.yukimochi.io"
	req.Header.Add("content-length", length)
	req.Header.Add("content-type", "application/activity+json")
	req.Header.Add("date", "Sun, 23 Dec 2018 07:39:37 GMT")
	req.Header.Add("digest", "SHA-256=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	req.Header.Add("signature", `keyId="https://innocent.yukimochi.io/users/YUKIMOCHI#main-key",algorithm="rsa-sha256",headers="(request-target) host date digest content-type",signature="MhxXhL21RVp8VmALER2U/oJlWldJAB2COiU2QmwGopLD2pw1c32gQvg0PaBRHfMBBOsidZuRRnj43Kn488zW2xV3n3DYWcGscSh527/hhRzcpLVX2kBqbf/WeQzJmfJVuOX4SzivVhnnUB8PvlPj5LRHpw4n/ctMTq37strKDl9iZg9rej1op1YFJagDxm3iPzAhnv8lzO4RI9dstt2i/sN5EfjXai97oS7EgI//Kj1wJCRk9Pw1iTsGfPTkbk/aVZwDt7QGGvGDdO0JJjsCqtIyjojoyD9hFY9GzMqvTwVIYJrh54AUHq2i80veybaOBbCFcEaK0RpKoLs101r5Uw=="`)

	_, _, _, err := decodeActivity(req)
	if err == nil || err.Error() != "crypto/rsa: verification error" {
		t.Fatalf("Expected error 'crypto/rsa: verification error', but got '%v'", err)
	}
}
