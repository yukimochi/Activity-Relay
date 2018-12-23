package activitypub

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/patrickmn/go-cache"
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

// GenerateSelfKey : Generate relay Actor from Publickey.
func (actor *Actor) GenerateSelfKey(hostname *url.URL, publickey *rsa.PublicKey) {
	actor.Context = []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"}
	actor.ID = hostname.String() + "/actor"
	actor.Type = "Service"
	actor.PreferredUsername = "relay"
	actor.Inbox = hostname.String() + "/inbox"
	actor.PublicKey = PublicKey{
		hostname.String() + "/actor#main-key",
		hostname.String() + "/actor",
		keyloader.GeneratePublicKeyPEMString(publickey),
	}
}

// RetrieveRemoteActor : Retrieve Actor from remote instance.
func (actor *Actor) RetrieveRemoteActor(url string, uaString string, cache *cache.Cache) error {
	var err error
	cacheData, found := cache.Get(url)
	if found {
		err = json.Unmarshal(cacheData.([]byte), &actor)
		if err != nil {
			cache.Delete(url)
		} else {
			return nil
		}
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/activity+json, application/ld+json")
	req.Header.Set("User-Agent", uaString)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(data, &actor)
	if err != nil {
		return err
	}
	cache.Set(url, data, 5*time.Minute)
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

// GenerateResponse : Generate activity response.
func (activity *Activity) GenerateResponse(host *url.URL, responseType string) Activity {
	return Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		host.String() + "/activities/" + uuid.NewV4().String(),
		host.String() + "/actor",
		responseType,
		&activity,
		nil,
		nil,
	}
}

// GenerateAnnounce : Generate Announce of activity.
func (activity *Activity) GenerateAnnounce(host *url.URL) Activity {
	return Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		host.String() + "/activities/" + uuid.NewV4().String(),
		host.String() + "/actor",
		"Announce",
		activity.ID,
		[]string{host.String() + "/actor/followers"},
		nil,
	}
}

// NestedActivity : Unwrap nested activity.
func (activity *Activity) NestedActivity() (*Activity, error) {
	mappedObject := activity.Object.(map[string]interface{})
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

// GenerateFromActor : Generate Webfinger resource from Actor.
func (resource *WebfingerResource) GenerateFromActor(hostname *url.URL, actor *Actor) {
	resource.Subject = "acct:" + actor.PreferredUsername + "@" + hostname.Host
	resource.Links = []WebfingerLink{
		WebfingerLink{
			"self",
			"application/activity+json",
			actor.ID,
		},
	}
}
