package deliver

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	httpdate "github.com/Songmu/go-httpdate"
	"github.com/go-fed/httpsig"
	"github.com/sirupsen/logrus"
)

func appendSignature(request *http.Request, body *[]byte, KeyID string, publicKey *rsa.PrivateKey) error {
	hash := sha256.New()
	hash.Write(*body)
	b := hash.Sum(nil)
	request.Header.Set("Digest", "SHA-256="+base64.StdEncoding.EncodeToString(b))
	request.Header.Set("Host", request.Host)

	signer, _, err := httpsig.NewSigner([]httpsig.Algorithm{httpsig.RSA_SHA256}, httpsig.DigestSha256, []string{httpsig.RequestTarget, "Host", "Date", "Digest", "Content-Type"}, httpsig.Signature, 3600)
	if err != nil {
		return err
	}
	err = signer.SignRequest(publicKey, KeyID, request, *body)
	if err != nil {
		return err
	}
	return nil
}

func sendActivity(inboxURL string, KeyID string, body []byte, publicKey *rsa.PrivateKey) error {
	req, _ := http.NewRequest("POST", inboxURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/activity+json")
	req.Header.Set("User-Agent", fmt.Sprintf("%s (golang net/http; Activity-Relay %s; %s)", globalConfig.ServerServiceName(), version, globalConfig.ServerHostname().Host))
	req.Header.Set("Date", httpdate.Time2Str(time.Now()))
	appendSignature(req, &body, KeyID, publicKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	logrus.Debug(inboxURL, resp.StatusCode)
	if resp.StatusCode/100 != 2 {
		return errors.New("Post " + inboxURL + ": " + resp.Status)
	}

	return nil
}
