package models

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	cache "github.com/patrickmn/go-cache"
	uuid "github.com/satori/go.uuid"
)

// PublicKey : Activity Certificate.
type PublicKey struct {
	ID           string `json:"id,omitempty"`
	Owner        string `json:"owner,omitempty"`
	PublicKeyPem string `json:"publicKeyPem,omitempty"`
}

// Endpoints : Contains SharedInbox address.
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
	Icon              *Image      `json:"icon,omitempty"`
	Image             *Image      `json:"image,omitempty"`
}

func (actor *Actor) Followers() string {
	return actor.ID + "/followers"
}

func NewActivityPubActorFromRelayConfig(globalConfig *RelayConfig) Actor {
	hostname := globalConfig.domain.String()
	publicKey := &globalConfig.actorKey.PublicKey
	publicKeyPemString := generatePublicKeyPEMString(publicKey)

	newActor := Actor{
		Context:           []string{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
		ID:                hostname + "/actor",
		Type:              "Service",
		Name:              globalConfig.serviceName,
		PreferredUsername: "relay",
		Summary:           globalConfig.serviceSummary,
		Inbox:             hostname + "/inbox",
		PublicKey: PublicKey{
			ID:           hostname + "/actor#main-key",
			Owner:        hostname + "/actor",
			PublicKeyPem: publicKeyPemString,
		},
	}

	if globalConfig.serviceIconURL != nil {
		newActor.Icon = &Image{
			URL: globalConfig.serviceIconURL.String(),
		}
	}
	if globalConfig.serviceImageURL != nil {
		newActor.Image = &Image{
			URL: globalConfig.serviceImageURL.String(),
		}
	}

	return newActor
}

// NewActivityPubActorFromRemoteActor : Retrieve Actor from remote instance.
func NewActivityPubActorFromRemoteActor(url string, uaString string, cache *cache.Cache) (Actor, error) {
	var actor = new(Actor)
	var err error
	cacheData, found := cache.Get(url)
	if found {
		err = json.Unmarshal(cacheData.([]byte), &actor)
		if err != nil {
			cache.Delete(url)
		} else {
			return *actor, nil
		}
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/activity+json")
	req.Header.Set("User-Agent", uaString)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return *actor, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return *actor, errors.New(resp.Status)
	}

	data, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(data, &actor)
	if err != nil {
		return *actor, err
	}
	cache.Set(url, data, 5*time.Minute)
	return *actor, nil
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

// GenerateReply : Generate activity to activity's actor.
func (activity *Activity) GenerateReply(actor Actor, object interface{}, activityType string) Activity {
	return Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		actor.ID + "/activities/" + uuid.NewV4().String(),
		actor.ID,
		activityType,
		object,
		[]string{activity.Actor},
		nil,
	}
}

// UnwrapInnerActivity : Unwrap inner activity.
func (activity *Activity) UnwrapInnerActivity() (*Activity, error) {
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
		return nil, errors.New("unwrap type failed")
	}
	return nil, errors.New("unwrap id failed")
}

// NewActivityPubActivity : Generate activity.
func NewActivityPubActivity(actor Actor, to []string, object interface{}, activityType string) Activity {
	return Activity{
		[]string{"https://www.w3.org/ns/activitystreams"},
		actor.ID + "/activities/" + uuid.NewV4().String(),
		actor.ID,
		activityType,
		object,
		to,
		nil,
	}
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

// GenerateWebfingerResource : Generate Webfinger resource.
func (actor *Actor) GenerateWebfingerResource(hostname *url.URL) WebfingerResource {
	resource := new(WebfingerResource)

	resource.Subject = "acct:" + actor.PreferredUsername + "@" + hostname.Host
	resource.Links = []WebfingerLink{
		{
			"self",
			"application/activity+json",
			actor.ID,
		},
	}
	return *resource
}

// NodeinfoResources : Nodeinfo Resources.
type NodeinfoResources struct {
	NodeinfoLinks NodeinfoLinks
	Nodeinfo      Nodeinfo
}

// NodeinfoLinks : Nodeinfo Link Resource.
type NodeinfoLinks struct {
	Links []NodeinfoLink `json:"links"`
}

// NodeinfoLink : Nodeinfo Link Resource.
type NodeinfoLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

// Nodeinfo : Nodeinfo Resource.
type Nodeinfo struct {
	Version           string           `json:"version"`
	Software          NodeinfoSoftware `json:"software"`
	Protocols         []string         `json:"protocols"`
	Services          NodeinfoServices `json:"services"`
	OpenRegistrations bool             `json:"openRegistrations"`
	Usage             NodeinfoUsage    `json:"usage"`
	Metadata          NodeinfoMetadata `json:"metadata"`
}

// NodeinfoSoftware : NodeinfoSoftware Resource.
type NodeinfoSoftware struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Repository string `json:"repository,omitempty"`
}

// NodeinfoServices : NodeinfoSoftware Resource.
type NodeinfoServices struct {
	Inbound  []string `json:"inbound"`
	Outbound []string `json:"outbound"`
}

// NodeinfoUsage : NodeinfoUsage Resource.
type NodeinfoUsage struct {
	Users NodeinfoUsageUsers `json:"users"`
}

// NodeinfoUsageUsers : NodeinfoUsageUsers Resource.
type NodeinfoUsageUsers struct {
	Total          int `json:"total"`
	ActiveMonth    int `json:"activeMonth"`
	ActiveHalfyear int `json:"activeHalfyear"`
}

// NodeinfoMetadata : NodeinfoMetadata Resource.
type NodeinfoMetadata struct {
}

// GenerateNodeinfoResources : Generate Nodeinfo resources.
func GenerateNodeinfoResources(hostname *url.URL, serverVersion string) NodeinfoResources {
	resources := new(NodeinfoResources)

	resources.NodeinfoLinks.Links = []NodeinfoLink{
		{
			"http://nodeinfo.diaspora.software/ns/schema/2.1",
			"https://" + hostname.Host + "/nodeinfo/2.1",
		},
	}
	resources.Nodeinfo = Nodeinfo{
		"2.1",
		NodeinfoSoftware{"activity-relay", serverVersion, "https://github.com/yukimochi/Activity-Relay"},
		[]string{"activitypub"},
		NodeinfoServices{[]string{}, []string{}},
		true,
		NodeinfoUsage{NodeinfoUsageUsers{0, 0, 0}},
		NodeinfoMetadata{},
	}

	return *resources
}
