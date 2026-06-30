package ingester

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// KeyRegistrar is satisfied by the seeder SDK client.
// The seeder must expose GET /encryption/public-key and PUT /encryption/public-key.
type KeyRegistrar interface {
	GetCurrentPublicKey(ctx context.Context) (string, error)
	RegisterPublicKey(ctx context.Context, publicKeyPEM string) error
}

const defaultPrivateKeyFilename = "ingester_private_key.pem"

// KeyManager generates (or loads) the ingester's RSA key pair and ensures the public key
// is registered with the seeder before any ingest run begins.
type KeyManager struct {
	keyPath            string
	privateKeyPEM      string
	requireExistingKey bool
	privateKey         *rsa.PrivateKey
}

func NewKeyManager(keyPath string) *KeyManager {
	return &KeyManager{keyPath: keyPath}
}

func NewKeyManagerFromEnv() (*KeyManager, error) {
	source, err := resolvePrivateKeySourceFromEnv()
	if err != nil {
		return nil, err
	}
	return &KeyManager{
		keyPath:            source.path,
		privateKeyPEM:      source.pem,
		requireExistingKey: true,
	}, nil
}

type privateKeySource struct {
	path string
	pem  string
}

func resolvePrivateKeySourceFromEnv() (privateKeySource, error) {
	if raw := os.Getenv("INGESTER_PRIVATE_KEY"); raw != "" {
		return privateKeySource{pem: raw}, nil
	}
	if path := os.Getenv("INGESTER_PRIVATE_KEY_PATH"); path != "" {
		return privateKeySource{path: path}, nil
	}
	path, err := DefaultPrivateKeyPath()
	if err != nil {
		return privateKeySource{}, err
	}
	return privateKeySource{path: path}, nil
}

func DefaultPrivateKeyPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}
	return filepath.Join(configDir, "payer-sync", defaultPrivateKeyFilename), nil
}

// EnsureKeyPair loads the key pair from disk if the file exists, otherwise generates
// a new RSA-2048 key pair and persists it. The private key never changes after first run.
func (km *KeyManager) EnsureKeyPair() error {
	if km.privateKeyPEM != "" {
		return km.loadPEM([]byte(km.privateKeyPEM), "INGESTER_PRIVATE_KEY")
	}
	if km.keyPath == "" {
		return fmt.Errorf("private key path is required when INGESTER_PRIVATE_KEY is unset")
	}
	if _, err := os.Stat(km.keyPath); err == nil {
		return km.load()
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat key %s: %w", km.keyPath, err)
	}
	if km.requireExistingKey {
		return fmt.Errorf("private key not found at %s; automatic generation is disabled in environment-configured mode", km.keyPath)
	}
	return km.generate()
}

func (km *KeyManager) load() error {
	pemBytes, err := os.ReadFile(km.keyPath)
	if err != nil {
		return fmt.Errorf("load key: %w", err)
	}
	return km.loadPEM(pemBytes, km.keyPath)
}

func (km *KeyManager) loadPEM(pemBytes []byte, source string) error {
	key, err := parsePrivateKeyPEM(pemBytes, source)
	if err != nil {
		return err
	}
	km.privateKey = key
	return nil
}

func (km *KeyManager) generate() error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err := os.MkdirAll(filepath.Dir(km.keyPath), 0o700); err != nil {
		return fmt.Errorf("create key directory for %s: %w", km.keyPath, err)
	}
	if err := os.WriteFile(km.keyPath, pemBytes, 0o600); err != nil {
		return fmt.Errorf("persist key to %s: %w", km.keyPath, err)
	}
	km.privateKey = key
	return nil
}

// PrivateKey returns the loaded private key. Nil if EnsureKeyPair has not been called.
func (km *KeyManager) PrivateKey() *rsa.PrivateKey {
	return km.privateKey
}

// PublicKeyPEM returns the PKIX PEM-encoded public key.
func (km *KeyManager) PublicKeyPEM() (string, error) {
	if km.privateKey == nil {
		return "", fmt.Errorf("key pair not loaded; call EnsureKeyPair first")
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&km.privateKey.PublicKey)
	if err != nil {
		return "", fmt.Errorf("marshal public key: %w", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER})), nil
}

// EnsureRegistered checks whether the seeder holds our public key and registers it if not.
// This is safe to call on every run — it is a no-op when keys already match.
func (km *KeyManager) EnsureRegistered(ctx context.Context, r KeyRegistrar) error {
	ourPEM, err := km.PublicKeyPEM()
	if err != nil {
		return err
	}

	currentPEM, err := r.GetCurrentPublicKey(ctx)
	if err != nil {
		return fmt.Errorf("get current public key from seeder: %w", err)
	}

	if publicKeysMatch(currentPEM, ourPEM) {
		return nil
	}

	if err := r.RegisterPublicKey(ctx, ourPEM); err != nil {
		return fmt.Errorf("register public key with seeder: %w", err)
	}
	return nil
}

// publicKeysMatch compares two PEM-encoded RSA public keys by modulus.
// Returns false if either is empty or unparseable — triggers re-registration.
func publicKeysMatch(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	keyA, err := parsePublicKeyPEM(a)
	if err != nil {
		return false
	}
	keyB, err := parsePublicKeyPEM(b)
	if err != nil {
		return false
	}
	return keyA.N.Cmp(keyB.N) == 0
}

func parsePublicKeyPEM(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("invalid PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaPub, nil
}

func parsePrivateKeyPEM(pemBytes []byte, source string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("invalid PEM at %s", source)
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse key from %s: %w", source, err)
	}
	return key, nil
}
