package models

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func ReadPublicKeyRSAFromString(pemString string) (*rsa.PublicKey, error) {
	pemByte := []byte(pemString)
	decoded, _ := pem.Decode(pemByte)
	defer func() {
		recover()
	}()
	keyInterface, err := x509.ParsePKIXPublicKey(decoded.Bytes)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	pub := keyInterface.(*rsa.PublicKey)
	return pub, nil
}

// RedisScanKeys returns all keys matching pattern using SCAN instead of
// the blocking KEYS command, so that large keyspaces do not stall Redis.
func RedisScanKeys(redisClient *redis.Client, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64
	for {
		batch, nextCursor, err := redisClient.Scan(context.TODO(), cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)
		cursor = nextCursor
		if cursor == 0 {
			return keys, nil
		}
	}
}

func redisHGetOrCreateWithDefault(redisClient *redis.Client, key string, field string, defaultValue string) (string, error) {
	keyExist, err := redisClient.HExists(context.TODO(), key, field).Result()
	if err != nil {
		return "", err
	}
	if keyExist {
		value, err := redisClient.HGet(context.TODO(), key, field).Result()
		if err != nil {
			return "", err
		}
		return value, nil
	} else {
		_, err := redisClient.HSet(context.TODO(), key, field, defaultValue).Result()
		if err != nil {
			return "", err
		}
		return defaultValue, nil
	}
}

func readPrivateKeyRSA(keyPath string) (*rsa.PrivateKey, error) {
	file, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	decoded, _ := pem.Decode(file)
	if decoded == nil {
		return nil, errors.New("ACTOR_PEM IS INVALID. FAILED TO READ")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(decoded.Bytes)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

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
