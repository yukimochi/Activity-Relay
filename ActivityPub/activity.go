package activitypub

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/yukimochi/httpsig"
)

var UA_STRING = os.Getenv("RELAY_SERVICENAME") + " (golang net/http; Activity-Relay v0.1.1; " + os.Getenv("RELAY_DOMAIN") + ")"

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
