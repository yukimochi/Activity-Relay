package deliver

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/Songmu/go-httpdate"
	"github.com/go-fed/httpsig"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"
)

func generatePublicKeyPEMString(publicKey *rsa.PublicKey) string {
	publicKeyByte := x509.MarshalPKCS1PublicKey(publicKey)
	publicKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicKeyByte,
		},
	)
	return string(publicKeyPem)
}

func TestAppendSignature(t *testing.T) {
	file, _ := os.Open("../misc/test/create.json")
	body, _ := ioutil.ReadAll(file)
	req, _ := http.NewRequest("POST", "https://localhost", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/activity+json")
	req.Header.Set("Date", httpdate.Time2Str(time.Now()))
	appendSignature(req, &body, "https://innocent.yukimochi.io/users/YUKIMOCHI#main-key", globalConfig.ActorKey())

	// Activated compatibilityForHTTPSignature11
	sign := req.Header.Get("Signature")
	activated := regexp.MustCompile(string("algorithm=\"" + httpsig.RSA_SHA256 + "\"")).MatchString(sign)
	if !activated {
		t.Fatalf("Failed - " + "compatibilityForHTTPSignature11 is not activated")
	}

	// Verify HTTPSignature
	verifier, err := httpsig.NewVerifier(req)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}
	err = verifier.Verify(globalConfig.ActorKey().Public(), httpsig.RSA_SHA256)
	if err != nil {
		t.Fatalf("Failed - " + err.Error())
	}

	// Verify Digest
	givenDigest := req.Header.Get("Digest")
	hash := sha256.New()
	hash.Write(body)
	b := hash.Sum(nil)
	calculatedDigest := "SHA-256=" + base64.StdEncoding.EncodeToString(b)

	if givenDigest != calculatedDigest {
		t.Fatalf("Failed - " + err.Error())
	}
}
