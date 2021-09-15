package api

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-fed/httpsig"
	"github.com/yukimochi/Activity-Relay/models"
)

func decodeActivity(request *http.Request) (*models.Activity, *models.Actor, []byte, error) {
	request.Header.Set("Host", request.Host)
	dataLen, _ := strconv.Atoi(request.Header.Get("Content-Length"))
	body := make([]byte, dataLen)
	request.Body.Read(body)

	// Verify HTTPSignature
	verifier, err := httpsig.NewVerifier(request)
	if err != nil {
		return nil, nil, nil, err
	}
	KeyID := verifier.KeyId()
	keyOwnerActor := new(models.Actor)
	err = keyOwnerActor.RetrieveRemoteActor(KeyID, fmt.Sprintf("%s (golang net/http; Activity-Relay %s; %s)", globalConfig.ServerServiceName(), version, globalConfig.ServerHostname().Host), actorCache)
	if err != nil {
		return nil, nil, nil, err
	}
	PubKey, err := models.ReadPublicKeyRSAFromString(keyOwnerActor.PublicKey.PublicKeyPem)
	if PubKey == nil {
		return nil, nil, nil, errors.New("Failed parse PublicKey from string")
	}
	if err != nil {
		return nil, nil, nil, err
	}
	err = verifier.Verify(PubKey, httpsig.RSA_SHA256)
	if err != nil {
		return nil, nil, nil, err
	}

	// Verify Digest
	givenDigest := request.Header.Get("Digest")
	hash := sha256.New()
	hash.Write(body)
	b := hash.Sum(nil)
	calculatedDigest := "SHA-256=" + base64.StdEncoding.EncodeToString(b)

	if givenDigest != calculatedDigest {
		return nil, nil, nil, errors.New("Digest header is mismatch")
	}

	// Parse Activity
	var activity models.Activity
	err = json.Unmarshal(body, &activity)
	if err != nil {
		return nil, nil, nil, err
	}

	var remoteActor models.Actor
	err = remoteActor.RetrieveRemoteActor(activity.Actor, fmt.Sprintf("%s (golang net/http; Activity-Relay %s; %s)", globalConfig.ServerServiceName(), version, globalConfig.ServerHostname().Host), actorCache)
	if err != nil {
		return nil, nil, nil, err
	}

	return &activity, &remoteActor, body, nil
}
