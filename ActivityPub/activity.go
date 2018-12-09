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

// UaString : Use for User-Agent
var UaString = os.Getenv("RELAY_SERVICENAME") + " (golang net/http; Activity-Relay v0.1.1; " + os.Getenv("RELAY_DOMAIN") + ")"

func appendSignature(request *http.Request, body *[]byte, KeyID string, publicKey *rsa.PrivateKey) error {
	hash := sha256.New()
	hash.Write(*body)
	b := hash.Sum(nil)
	request.Header.Set("Digest", "SHA-256="+base64.StdEncoding.EncodeToString(b))
	request.Header.Set("Host", request.Host)

	signer, _, err := httpsig.NewSigner([]httpsig.Algorithm{httpsig.RSA_SHA256}, []string{httpsig.RequestTarget, "Host", "Date", "Digest", "Content-Type"}, httpsig.Signature)
	if err != nil {
		return err
	}
	err = signer.SignRequest(publicKey, KeyID, request)
	if err != nil {
		return err
	}
	return nil
}

// SendActivity : Send ActivityPub activity
func SendActivity(inboxURL string, KeyID string, body []byte, publicKey *rsa.PrivateKey) error {
	req, _ := http.NewRequest("POST", inboxURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/activity+json, application/ld+json")
	req.Header.Set("User-Agent", UaString)
	req.Header.Set("Date", httpdate.Time2Str(time.Now()))
	appendSignature(req, &body, KeyID, publicKey)
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
