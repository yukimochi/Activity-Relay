package api

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-fed/httpsig"
	"github.com/yukimochi/Activity-Relay/models"
)

func decodeActivity(request *http.Request) (*models.Activity, *models.Actor, []byte, error) {
	request.Header.Set("Host", request.Host)
	body, err := io.ReadAll(request.Body)

	// Verify HTTPSignature
	verifier, err := httpsig.NewVerifier(request)
	if err != nil {
		return nil, nil, nil, err
	}
	KeyID := verifier.KeyId()
	keyOwnerActor, err := models.NewActivityPubActorFromRemoteActor(KeyID, RelayActor.ID, fmt.Sprintf("%s (golang net/http; Activity-Relay %s; %s)", GlobalConfig.ServerServiceName(), version, GlobalConfig.ServerHostname().Host), GlobalConfig.ActorKey(), ActorCache)
	if err != nil {
		return nil, nil, nil, err
	}
	PubKey, err := models.ReadPublicKeyRSAFromString(keyOwnerActor.PublicKey.PublicKeyPem)
	if PubKey == nil {
		return nil, nil, nil, errors.New("failed parse PublicKey from string")
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
		return nil, nil, nil, errors.New("digest header is mismatch")
	}

	// Parse Activity
	var activity models.Activity
	err = json.Unmarshal(body, &activity)
	if err != nil {
		return nil, nil, nil, err
	}
	remoteActor, err := models.NewActivityPubActorFromRemoteActor(activity.Actor, RelayActor.ID, fmt.Sprintf("%s (golang net/http; Activity-Relay %s; %s)", GlobalConfig.ServerServiceName(), version, GlobalConfig.ServerHostname().Host), GlobalConfig.ActorKey(), ActorCache)
	if err != nil {
		return nil, nil, nil, err
	}

	return &activity, &remoteActor, body, nil
}

func fetchOriginalActivityFromURL(url string) (*models.Activity, *models.Actor, error) {
	remoteActivity, err := models.NewActivityPubActivityFromRemoteActivity(url, fmt.Sprintf("%s (golang net/http; Activity-Relay %s; %s)", GlobalConfig.ServerServiceName(), version, GlobalConfig.ServerHostname().Host))
	if err != nil {
		return nil, nil, err
	}
	remoteActor, err := models.NewActivityPubActorFromRemoteActor(remoteActivity.Actor, RelayActor.ID, fmt.Sprintf("%s (golang net/http; Activity-Relay %s; %s)", GlobalConfig.ServerServiceName(), version, GlobalConfig.ServerHostname().Host), GlobalConfig.ActorKey(), ActorCache)
	if err != nil {
		return &remoteActivity, nil, err
	}
	return &remoteActivity, &remoteActor, err
}
