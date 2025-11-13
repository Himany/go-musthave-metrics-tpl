package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// RSAEncryptor предоставляет функции для асимметричного шифрования RSA
type RSAEncryptor struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// NewRSAEncryptorFromPublicKey создает новый RSAEncryptor с публичным ключом для шифрования
func NewRSAEncryptorFromPublicKey(publicKeyPath string) (*RSAEncryptor, error) {
	if publicKeyPath == "" {
		return &RSAEncryptor{}, nil
	}

	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &RSAEncryptor{
		publicKey: publicKey,
	}, nil
}

// NewRSAEncryptorFromPrivateKey создает новый RSAEncryptor с приватным ключом для дешифрования
func NewRSAEncryptorFromPrivateKey(privateKeyPath string) (*RSAEncryptor, error) {
	if privateKeyPath == "" {
		return &RSAEncryptor{}, nil
	}

	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return &RSAEncryptor{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// Encrypt шифрует данные с помощью публичного ключа
func (r *RSAEncryptor) Encrypt(data []byte) ([]byte, error) {
	if r.publicKey == nil {
		return data, nil
	}

	encryptedData, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, r.publicKey, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	return encryptedData, nil
}

// Decrypt дешифрует данные с помощью приватного ключа
func (r *RSAEncryptor) Decrypt(encryptedData []byte) ([]byte, error) {
	if r.privateKey == nil {
		return encryptedData, nil
	}

	decryptedData, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, r.privateKey, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return decryptedData, nil
}

// IsEnabled возвращает true, если шифрование включено (есть ключи)
func (r *RSAEncryptor) IsEnabled() bool {
	return r.publicKey != nil || r.privateKey != nil
}

// loadPublicKey загружает публичный ключ из файла
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

// loadPrivateKey загружает приватный ключ из файла
func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privateKey, nil
}
