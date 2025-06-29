package deliver

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/Songmu/go-httpdate"
	"github.com/go-fed/httpsig"
)

func TestAppendSignature(t *testing.T) {
	file, _ := os.Open("../misc/test/create.json")
	body, _ := io.ReadAll(file)
	req, _ := http.NewRequest("POST", "https://localhost", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/activity+json")
	req.Header.Set("Date", httpdate.Time2Str(time.Now()))
	appendSignature(req, &body, "https://innocent.yukimochi.io/users/YUKIMOCHI#main-key", GlobalConfig.ActorKey())

	// Activated compatibilityForHTTPSignature11
	sign := req.Header.Get("Signature")
	activated := regexp.MustCompile(string("algorithm=" + httpsig.RSA_SHA256 + "\"")).MatchString(sign)
	if !activated {
		t.Fatalf("Expected Signature header to contain algorithm=\"%s\", but got: %s", httpsig.RSA_SHA256, sign)
	}

	// Verify HTTPSignature
	verifier, err := httpsig.NewVerifier(req)
	if err != nil {
		t.Fatalf("Failed to create HTTPSignature verifier: %v", err)
	}
	err = verifier.Verify(GlobalConfig.ActorKey().Public(), httpsig.RSA_SHA256)
	if err != nil {
		t.Fatalf("HTTPSignature verification failed: %v", err)
	}

	// Verify Digest
	givenDigest := req.Header.Get("Digest")
	hash := sha256.New()
	hash.Write(body)
	b := hash.Sum(nil)
	calculatedDigest := "SHA-256=" + base64.StdEncoding.EncodeToString(b)

	if givenDigest != calculatedDigest {
		t.Fatalf("Expected Digest header to be '%s', but got '%s'", calculatedDigest, givenDigest)
	}
}
