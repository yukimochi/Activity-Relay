package activitypub

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/satori/go.uuid"
	"github.com/yukimochi/Activity-Relay/KeyLoader"
)

// PublicKey : Activity Certificate.
type PublicKey struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	PublicKeyPem string `json:"publicKeyPem"`
}

//Endpoints : Contains SharedInbox address.
type Endpoints struct {
	SharedInbox string `json:"sharedInbox"`
}

// Actor : ActivityPub Actor.
type Actor struct {
	Context           interface{} `json:"@context"`
	ID                string      `json:"id"`
	Type              string      `json:"type"`
	PreferredUsername string      `json:"preferredUsername"`
	Inbox             string      `json:"inbox"`
	Endpoints         *Endpoints  `json:"endpoints"`
	PublicKey         PublicKey   `json:"publicKey"`
}

func (a *Actor) GenerateSelfKey(hostname *url.URL, publickey *rsa.PublicKey) {
	a.Context = []string{"https://www.w3.org/ns/activitystreams"}
	a.ID = hostname.String() + "/actor"
	a.Type = "Service"
	a.PreferredUsername = "relay"
	a.Inbox = hostname.String() + "/inbox"
	a.PublicKey = PublicKey{
		hostname.String() + "/actor#main-key",
		hostname.String() + "/actor",
		keyloader.GeneratePublicKeyPEMString(publickey),
	}
}

func (a *Actor) RetrieveRemoteActor(url string) error {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/activity+json, application/ld+json")
	req.Header.Set("User-Agent", UA_STRING)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, &a)
	if err != nil {
		return err
	}
	return nil
}

// Activity : ActivityPub Activity.
type Activity struct {
	Context interface{} `json:"@context"`
	ID      string      `json:"id"`
	Actor   string      `json:"actor"`
	Type    string      `json:"type"`
	Object  interface{} `json:"object"`
	To      []string    `json:"to"`
	Cc      []string    `json:"cc"`
}

func (a *Activity) GenerateResponse(host *url.URL, responseType string) Activity {
	return Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		host.String() + "/activities/" + uuid.NewV4().String(),
		host.String() + "/actor",
		responseType,
		&a,
		nil,
		nil,
	}
}

func (a *Activity) GenerateAnnounce(host *url.URL) Activity {
	return Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		host.String() + "/activities/" + uuid.NewV4().String(),
		host.String() + "/actor",
		"Announce",
		a.ID,
		[]string{host.String() + "/actor/followers"},
		nil,
	}
}

func (a *Activity) NestedActivity() (*Activity, error) {
	mappedObject := a.Object.(map[string]interface{})
	if id, ok := mappedObject["id"].(string); ok {
		if nestedType, ok := mappedObject["type"].(string); ok {
			actor, ok := mappedObject["actor"].(string)
			if !ok {
				actor = ""
			}
			switch object := mappedObject["object"].(type) {
			case string:
				return &Activity{
					ID:     id,
					Type:   nestedType,
					Actor:  actor,
					Object: object,
				}, nil
			default:
				return &Activity{
					ID:     id,
					Type:   nestedType,
					Actor:  actor,
					Object: mappedObject["object"],
				}, nil
			}
		}
		return nil, errors.New("Can't assart type")
	}
	return nil, errors.New("Can't assart id")
}

// Signature : ActivityPub Header Signature.
type Signature struct {
	Type           string `json:"type"`
	Creator        string `json:"creator"`
	Created        string `json:"created"`
	SignatureValue string `json:"signatureValue"`
}

// WebfingerLink : Webfinger Link Resource.
type WebfingerLink struct {
	Rel  string `json:"rel"`
	Type string `json:"type"`
	Href string `json:"href"`
}

// WebfingerResource : Webfinger Resource.
type WebfingerResource struct {
	Subject string          `json:"subject"`
	Links   []WebfingerLink `json:"links"`
}

func (a *WebfingerResource) GenerateFromActor(hostname *url.URL, actor *Actor) {
	a.Subject = "acct:" + actor.PreferredUsername + "@" + hostname.Host
	a.Links = []WebfingerLink{
		WebfingerLink{
			"self",
			"application/activity+json",
			actor.ID,
		},
	}
}
