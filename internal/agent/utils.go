package agent

import (
	"bytes"
	"compress/gzip"
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

func compressBody(jsonData []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(jsonData); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
