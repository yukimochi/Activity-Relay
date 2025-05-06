package models

import (
	"crypto/rsa"
	"github.com/go-fed/httpsig"
	"net/http"
	"regexp"
)

func compatibilityForHTTPSignature11(request *http.Request, algorithm httpsig.Algorithm) {
	signature := request.Header.Get("Signature")
	targetString := regexp.MustCompile("algorithm=\"hs2019\"")
	signature = targetString.ReplaceAllString(signature, string("algorithm=\""+algorithm+"\""))
	request.Header.Set("Signature", signature)
}

func AppendSignature(request *http.Request, body *[]byte, KeyID string, privateKey *rsa.PrivateKey) error {
	request.Header.Set("Host", request.Host)

	signer, _, err := httpsig.NewSigner([]httpsig.Algorithm{httpsig.RSA_SHA256}, httpsig.DigestSha256, []string{httpsig.RequestTarget, "Host", "Date", "Digest", "Content-Type"}, httpsig.Signature, 60*60)
	if err != nil {
		return err
	}
	err = signer.SignRequest(privateKey, KeyID, request, *body)
	if err != nil {
		return err
	}
	compatibilityForHTTPSignature11(request, httpsig.RSA_SHA256) // Compatibility for Misskey <12.111.0
	return nil
}
