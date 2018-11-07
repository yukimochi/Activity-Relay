package activitypub

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/yukimochi/Activity-Relay/KeyLoader"
	"github.com/yukimochi/httpsig"
)

var UA_STRING = "YUKIMOCHI Toot Relay Service (golang net/http; Activity-Relay v0.0.2; " + os.Getenv("RELAY_DOMAIN") + ")"

func appendSignature(r *http.Request, body *[]byte, KeyID string, pKey *rsa.PrivateKey) error {
	hash := sha256.New()
	hash.Write(*body)
	b := hash.Sum(nil)
	r.Header.Set("Digest", "SHA-256="+base64.StdEncoding.EncodeToString(b))
	r.Header.Set("Host", r.Host)

	signer, _, err := httpsig.NewSigner([]httpsig.Algorithm{httpsig.RSA_SHA256}, []string{httpsig.RequestTarget, "Host", "Date", "Digest", "Content-Type"}, httpsig.Signature)
	if err != nil {
		return err
	}
	err = signer.SignRequest(pKey, KeyID, r)
	if err != nil {
		return err
	}
	return nil
}

// SendActivity : Send ActivityPub Activity.
func SendActivity(inboxURL string, KeyID string, refBytes []byte, pKey *rsa.PrivateKey) error {
	req, _ := http.NewRequest("POST", inboxURL, bytes.NewBuffer(refBytes))
	req.Header.Set("Content-Type", "application/activity+json, application/ld+json")
	req.Header.Set("User-Agent", UA_STRING)
	req.Header.Set("Date", httpdate.Time2Str(time.Now()))
	appendSignature(req, &refBytes, KeyID, pKey)
	client := &http.Client{Timeout: time.Duration(5) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	fmt.Println(inboxURL, resp.StatusCode)
	if resp.StatusCode/100 != 2 {
		return errors.New(resp.Status)
	}

	return nil
}

// RetrieveActor : Retrieve Remote actor
func RetrieveActor(url string) (*Actor, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/activity+json, application/ld+json")
	req.Header.Set("User-Agent", UA_STRING)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var actor Actor
	err = json.Unmarshal(data, &actor)
	if err != nil {
		return nil, err
	}
	return &actor, nil
}

// DescribeNestedActivity : Descrive Nested Activity Series
func DescribeNestedActivity(nestedActivity interface{}) (*Activity, error) {
	mappedObject := nestedActivity.(map[string]interface{})

	return &Activity{
		ID:     mappedObject["id"].(string),
		Type:   mappedObject["type"].(string),
		Actor:  mappedObject["actor"].(string),
		Object: mappedObject["object"],
	}, nil
}

// GenerateActor : Generate Actor by hostname and publickey
func GenerateActor(hostname *url.URL, publickey *rsa.PublicKey) Actor {
	return Actor{
		[]string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		hostname.String() + "/actor",
		"Service",
		"relay",
		hostname.String() + "/inbox",
		nil,
		PublicKey{
			hostname.String() + "/actor#main-key",
			hostname.String() + "/actor",
			keyloader.GeneratePublicKeyPEMString(publickey),
		},
	}
}

// GenerateWebfingerResource : Generate Webfinger Resource
func GenerateWebfingerResource(hostname *url.URL, actor *Actor) WebfingerResource {
	return WebfingerResource{
		"acct:" + actor.PreferredUsername + "@" + hostname.Host,
		[]WebfingerLink{
			WebfingerLink{
				"self",
				"application/activity+json",
				actor.ID,
			},
		},
	}
}

// GenerateActivityResponse : Generate Responce Activity to Activity
func GenerateActivityResponse(host *url.URL, to *url.URL, responseType string, activity Activity) Activity {
	return Activity{
		[]string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		host.String() + "/actor#accepts/follows/" + to.Host,
		host.String() + "/actor",
		responseType,
		&activity,
		nil,
		nil,
	}
}
