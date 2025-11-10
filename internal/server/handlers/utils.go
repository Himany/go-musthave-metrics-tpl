package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
)

func bodySignature(jsonData []byte, key string) []byte {
	if key == "" {
		return nil
	}

	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write(jsonData)
	return hasher.Sum(nil)
}
