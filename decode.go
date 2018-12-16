package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/yukimochi/Activity-Relay/ActivityPub"
	"github.com/yukimochi/Activity-Relay/KeyLoader"
	"github.com/yukimochi/httpsig"
)

func decodeActivity(request *http.Request) (*activitypub.Activity, *activitypub.Actor, []byte, error) {
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
	remoteActor := new(activitypub.Actor)
	err = remoteActor.RetrieveRemoteActor(KeyID, actorCache)
	if err != nil {
		return nil, nil, nil, err
	}
	PubKey, err := keyloader.ReadPublicKeyRSAfromString(remoteActor.PublicKey.PublicKeyPem)
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
	calcurateDigest := "SHA-256=" + base64.StdEncoding.EncodeToString(b)

	if givenDigest != calcurateDigest {
		return nil, nil, nil, errors.New("Digest header is mismatch")
	}

	var activity activitypub.Activity
	err = json.Unmarshal(body, &activity)
	if err != nil {
		return nil, nil, nil, err
	}

	return &activity, remoteActor, body, nil
}
