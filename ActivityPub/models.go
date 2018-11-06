package activitypub

// PublicKey : Activity Certificate.
type PublicKey struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	PublicKeyPem string `json:"publicKeyPem"`
}

type endpoints struct {
	SharedInbox string `json:"sharedInbox"`
}

// Actor : ActivityPub Actor.
type Actor struct {
	Context           interface{} `json:"@context"`
	ID                string      `json:"id"`
	Type              string      `json:"type"`
	PreferredUsername string      `json:"preferredUsername"`
	Inbox             string      `json:"inbox"`
	Endpoints         *endpoints  `json:"endpoints"`
	PublicKey         PublicKey   `json:"publicKey"`
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

// Signature : ActivityPub Header Signature.
type Signature struct {
	Type           string `json:"type"`
	Creator        string `json:"creator"`
	Created        string `json:"created"`
	SignatureValue string `json:"signatureValue"`
}

// WebfingerResource : Webfinger Resource.
type WebfingerResource struct {
	Subject string          `json:"subject"`
	Links   []WebfingerLink `json:"links"`
}

// WebfingerLink : Webfinger Link Resource.
type WebfingerLink struct {
	Rel  string `json:"rel"`
	Type string `json:"type"`
	Href string `json:"href"`
}
