package keyloader

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"unsafe"
)

func ReadPrivateKeyRSAfromPath(path string) (*rsa.PrivateKey, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoded, _ := pem.Decode(file)
	priv, err := x509.ParsePKCS1PrivateKey(decoded.Bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

func ReadPublicKeyRSAfromString(pemString string) (*rsa.PublicKey, error) {
	pemByte := *(*[]byte)(unsafe.Pointer(&pemString))
	decoded, _ := pem.Decode(pemByte)
	keyInterface, err := x509.ParsePKIXPublicKey(decoded.Bytes)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	pub := keyInterface.(*rsa.PublicKey)
	return pub, nil
}

func GeneratePublicKeyPEMString(pub *rsa.PublicKey) string {
	pubkeyBytes := x509.MarshalPKCS1PublicKey(pub)
	pubkeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: pubkeyBytes,
		},
	)
	return string(pubkeyPem)
}
