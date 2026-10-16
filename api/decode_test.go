package api

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/yukimochi/Activity-Relay/models"
)

func TestDecodeActivity(t *testing.T) {
	t.Skip("Skipping TestDecodeActivity due to external dependency issues")
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

func TestDecodeActivityWithInvalidDigest(t *testing.T) {
	originalConfig := GlobalConfig
	GlobalConfig = nil
	t.Cleanup(func() { GlobalConfig = originalConfig })
	for _, digest := range []string{"", "SHA-256=invalid"} {
		t.Run(digest, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/inbox", bytes.NewReader([]byte(`{"type":"Follow"}`)))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Digest", digest)
			req.Header.Set("Signature", `keyId="https://key.example/actor#key",algorithm="rsa-sha256",headers="(request-target) host date digest",signature="AAABBBCCC"`)
			activity, actor, body, err := decodeActivity(req)
			if err == nil || err.Error() != "digest header is mismatch" {
				t.Fatalf("expected digest mismatch before key fetching, got %v", err)
			}
			if activity != nil || actor != nil || body != nil {
				t.Fatal("rejected request returned decoded data")
			}
		})
	}
}

func TestDecodeActivityValidDigestStillRequiresSignature(t *testing.T) {
	body := []byte(`{"type":"Follow"}`)
	hash := sha256.Sum256(body)
	req, err := http.NewRequest("POST", "/inbox", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Digest", "SHA-256="+base64.StdEncoding.EncodeToString(hash[:]))
	_, _, _, err = decodeActivity(req)
	if err == nil || err.Error() != "neither \"Signature\" nor \"Authorization\" have signature parameters" {
		t.Fatalf("valid digest must not bypass signature verification: %v", err)
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
	t.Skip("Skipping TestDecodeActivityWithNotFoundKeyId due to external dependency issues")
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
