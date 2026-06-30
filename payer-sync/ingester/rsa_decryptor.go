package ingester

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
)

// RSADecryptor decrypts files using RSA-OAEP-SHA256 (key wrap) + AES-256-GCM.
// It expects the PEM-encoded PKCS1 private key to be provided via INGESTER_PRIVATE_KEY_PATH
// or INGESTER_PRIVATE_KEY (raw PEM bytes).
type RSADecryptor struct {
	privateKey *rsa.PrivateKey
}

// NewRSADecryptor loads the private key from environment.
func NewRSADecryptor() (*RSADecryptor, error) {
	source, err := resolvePrivateKeySourceFromEnv()
	if err != nil {
		return nil, err
	}

	var (
		pemBytes []byte
		label    string
	)
	if source.pem != "" {
		pemBytes = []byte(source.pem)
		label = "INGESTER_PRIVATE_KEY"
	} else {
		path := source.path
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read private key: %w", err)
		}
		pemBytes = b
		label = path
	}

	key, err := parsePrivateKeyPEM(pemBytes, label)
	if err != nil {
		return nil, err
	}
	return &RSADecryptor{privateKey: key}, nil
}

// NewRSADecryptorFromKey creates a decryptor from an already-loaded private key.
func NewRSADecryptorFromKey(key *rsa.PrivateKey) *RSADecryptor {
	return &RSADecryptor{privateKey: key}
}

func (d *RSADecryptor) Decrypt(encryptedBytes []byte, encryptedDataKey, nonce string) ([]byte, error) {
	wrappedKey, err := base64.StdEncoding.DecodeString(encryptedDataKey)
	if err != nil {
		return nil, fmt.Errorf("decode encrypted data key: %w", err)
	}
	dataKey, err := rsa.DecryptOAEP(sha256.New(), nil, d.privateKey, wrappedKey, nil)
	if err != nil {
		return nil, fmt.Errorf("unwrap data key: %w", err)
	}

	nonceBytes, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		return nil, fmt.Errorf("decode nonce: %w", err)
	}

	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	plaintext, err := gcm.Open(nil, nonceBytes, encryptedBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM decrypt: %w", err)
	}
	return plaintext, nil
}
