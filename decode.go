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

func decodeActivity(r *http.Request) (*activitypub.Activity, *activitypub.Actor, []byte, error) {
	r.Header.Set("Host", r.Host)
	dataLen, _ := strconv.Atoi(r.Header.Get("Content-Length"))
	body := make([]byte, dataLen)
	r.Body.Read(body)

	// Verify HTTPSignature
	verifier, _ := httpsig.NewVerifier(r)
	KeyID := verifier.KeyId()
	remoteActor, err := activitypub.RetrieveActor(KeyID)
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
	givenDigest := r.Header.Get("Digest")
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
