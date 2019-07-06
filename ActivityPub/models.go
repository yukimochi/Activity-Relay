package activitypub

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	cache "github.com/patrickmn/go-cache"
	uuid "github.com/satori/go.uuid"
	keyloader "github.com/yukimochi/Activity-Relay/KeyLoader"
)

// PublicKey : Activity Certificate.
type PublicKey struct {
	ID           string `json:"id,omitempty"`
	Owner        string `json:"owner,omitempty"`
	PublicKeyPem string `json:"publicKeyPem,omitempty"`
}

//Endpoints : Contains SharedInbox address.
type Endpoints struct {
	SharedInbox string `json:"sharedInbox,omitempty"`
}

// Image : Image Object.
type Image struct {
	URL string `json:"url,omitempty"`
}

// Actor : ActivityPub Actor.
type Actor struct {
	Context           interface{} `json:"@context,omitempty"`
	ID                string      `json:"id,omitempty"`
	Type              string      `json:"type,omitempty"`
	Name              string      `json:"name,omitempty"`
	PreferredUsername string      `json:"preferredUsername,omitempty"`
	Summary           string      `json:"summary,omitempty"`
	Inbox             string      `json:"inbox,omitempty"`
	Endpoints         *Endpoints  `json:"endpoints,omitempty"`
	PublicKey         PublicKey   `json:"publicKey,omitempty"`
	Icon              Image       `json:"icon,omitempty"`
	Image             Image       `json:"image,omitempty"`
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
	req.Header.Set("Accept", "application/activity+json")
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
	Context interface{} `json:"@context,omitempty"`
	ID      string      `json:"id,omitempty"`
	Actor   string      `json:"actor,omitempty"`
	Type    string      `json:"type,omitempty"`
	Object  interface{} `json:"object,omitempty"`
	To      []string    `json:"to,omitempty"`
	Cc      []string    `json:"cc,omitempty"`
}

// GenerateFollowbackRequest : Generate follow response.
func (activity *Activity) GenerateFollowbackRequest(host *url.URL) Activity {
	return Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		host.String() + "/activities/" + uuid.NewV4().String(),
		host.String() + "/actor",
		"Follow",
		activity.Actor,
		[]string{activity.Actor},
		nil,
	}
}

// GenerateResponse : Generate activity response.
func (activity *Activity) GenerateResponse(host *url.URL, responseType string) Activity {
	return Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		host.String() + "/activities/" + uuid.NewV4().String(),
		host.String() + "/actor",
		responseType,
		&activity,
		[]string{activity.Actor},
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

// ActivityObject : ActivityPub Activity.
type ActivityObject struct {
	ID      string   `json:"id,omitempty"`
	Type    string   `json:"type,omitempty"`
	Name    string   `json:"name,omitempty"`
	Content string   `json:"content,omitempty"`
	To      []string `json:"to,omitempty"`
	Cc      []string `json:"cc,omitempty"`
}

// Signature : ActivityPub Header Signature.
type Signature struct {
	Type           string `json:"type,omitempty"`
	Creator        string `json:"creator,omitempty"`
	Created        string `json:"created,omitempty"`
	SignatureValue string `json:"signatureValue,omitempty"`
}

// WebfingerLink : Webfinger Link Resource.
type WebfingerLink struct {
	Rel  string `json:"rel,omitempty"`
	Type string `json:"type,omitempty"`
	Href string `json:"href,omitempty"`
}

// WebfingerResource : Webfinger Resource.
type WebfingerResource struct {
	Subject string          `json:"subject,omitempty"`
	Links   []WebfingerLink `json:"links,omitempty"`
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
