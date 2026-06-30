package ingester

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---- test double ----

type mockKeyRegistrar struct {
	currentPEM  string
	registerErr error
	registered  []string
}

func (m *mockKeyRegistrar) GetCurrentPublicKey(_ context.Context) (string, error) {
	return m.currentPEM, nil
}

func (m *mockKeyRegistrar) RegisterPublicKey(_ context.Context, publicKeyPEM string) error {
	m.registered = append(m.registered, publicKeyPEM)
	return m.registerErr
}

// generateForeignPublicKeyPEM returns a PEM-encoded RSA public key unrelated to the test KeyManager.
func generateForeignPublicKeyPEM(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))
}

// ---- tests ----

// KM-001: A new RSA-2048 key pair is generated and persisted to disk when no key file exists.
func TestKeyManager_KM001_GeneratesKeyPairOnFirstRun(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "private_key.pem")

	km := NewKeyManager(keyPath)
	if err := km.EnsureKeyPair(); err != nil {
		t.Fatalf("EnsureKeyPair() unexpected error: %v", err)
	}

	if km.PrivateKey() == nil {
		t.Fatal("expected private key to be non-nil after EnsureKeyPair()")
	}
	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("expected key file to be persisted at %s: %v", keyPath, err)
	}
}

// KM-002: When a key file already exists, the same key is loaded — not regenerated.
// This ensures the public key registered with the seeder never silently changes between runs.
func TestKeyManager_KM002_LoadsExistingKeyWithoutRegenerating(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "private_key.pem")

	km1 := NewKeyManager(keyPath)
	if err := km1.EnsureKeyPair(); err != nil {
		t.Fatal(err)
	}
	originalModulus := km1.PrivateKey().N.Bytes()

	km2 := NewKeyManager(keyPath) // fresh manager, same file
	if err := km2.EnsureKeyPair(); err != nil {
		t.Fatal(err)
	}

	if string(km2.PrivateKey().N.Bytes()) != string(originalModulus) {
		t.Fatal("key was regenerated instead of loaded — public key mismatch would break seeder registration")
	}
}

func TestKeyManager_GeneratesKeyPairWhenParentDirectoryDoesNotExist(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "nested", "keys", "private_key.pem")

	km := NewKeyManager(keyPath)
	if err := km.EnsureKeyPair(); err != nil {
		t.Fatalf("EnsureKeyPair() unexpected error: %v", err)
	}

	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("expected nested key file to be persisted at %s: %v", keyPath, err)
	}
}

func TestNewKeyManagerFromEnv_PrefersRawPEM(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "private_key.pem")

	sourceKM := NewKeyManager(keyPath)
	if err := sourceKM.EnsureKeyPair(); err != nil {
		t.Fatal(err)
	}

	pemBytes, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("INGESTER_PRIVATE_KEY", string(pemBytes))
	t.Setenv("INGESTER_PRIVATE_KEY_PATH", filepath.Join(t.TempDir(), "should-not-be-used.pem"))

	km, err := NewKeyManagerFromEnv()
	if err != nil {
		t.Fatalf("NewKeyManagerFromEnv() unexpected error: %v", err)
	}
	if err := km.EnsureKeyPair(); err != nil {
		t.Fatalf("EnsureKeyPair() unexpected error: %v", err)
	}

	if km.keyPath != "" {
		t.Fatalf("expected raw PEM source to avoid a file path, got %q", km.keyPath)
	}
	if km.PrivateKey() == nil {
		t.Fatal("expected private key to be loaded from INGESTER_PRIVATE_KEY")
	}
}

func TestKeyManager_RequireExistingKey_DoesNotGenerateWhenMissing(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing.pem")

	t.Setenv("INGESTER_PRIVATE_KEY", "")
	t.Setenv("INGESTER_PRIVATE_KEY_PATH", missingPath)

	km, err := NewKeyManagerFromEnv()
	if err != nil {
		t.Fatalf("NewKeyManagerFromEnv() unexpected error: %v", err)
	}

	err = km.EnsureKeyPair()
	if err == nil {
		t.Fatal("expected EnsureKeyPair() to fail when key is missing in strict mode")
	}
	if !strings.Contains(err.Error(), "automatic generation is disabled") {
		t.Fatalf("expected strict-mode error to mention generation disabled, got: %v", err)
	}
}

func TestKeyManager_RequireExistingKey_LoadsExistingKey(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "private_key.pem")

	sourceKM := NewKeyManager(keyPath)
	if err := sourceKM.EnsureKeyPair(); err != nil {
		t.Fatal(err)
	}

	t.Setenv("INGESTER_PRIVATE_KEY", "")
	t.Setenv("INGESTER_PRIVATE_KEY_PATH", keyPath)

	km, err := NewKeyManagerFromEnv()
	if err != nil {
		t.Fatalf("NewKeyManagerFromEnv() unexpected error: %v", err)
	}
	if err := km.EnsureKeyPair(); err != nil {
		t.Fatalf("EnsureKeyPair() unexpected error in strict mode with existing key: %v", err)
	}
	if km.PrivateKey() == nil {
		t.Fatal("expected private key to load in strict mode when file exists")
	}
}

// KM-003: Registration is skipped when the seeder already holds our public key.
// This is the steady-state path: every run after the first is a no-op.
func TestKeyManager_KM003_SkipsRegistrationWhenSeederAlreadyHasOurKey(t *testing.T) {
	km := NewKeyManager(filepath.Join(t.TempDir(), "key.pem"))
	if err := km.EnsureKeyPair(); err != nil {
		t.Fatal(err)
	}
	ourPEM, _ := km.PublicKeyPEM()

	registrar := &mockKeyRegistrar{currentPEM: ourPEM}
	if err := km.EnsureRegistered(context.Background(), registrar); err != nil {
		t.Fatalf("EnsureRegistered() unexpected error: %v", err)
	}

	if len(registrar.registered) != 0 {
		t.Fatalf("expected 0 registration calls (key already registered), got %d", len(registrar.registered))
	}
}

// KM-004: Our public key is POSTed to the seeder when it holds a different key.
// This covers first run (seeder has its own generated key) and key rotation.
func TestKeyManager_KM004_RegistersWhenSeederHasDifferentKey(t *testing.T) {
	km := NewKeyManager(filepath.Join(t.TempDir(), "key.pem"))
	if err := km.EnsureKeyPair(); err != nil {
		t.Fatal(err)
	}
	ourPEM, _ := km.PublicKeyPEM()

	registrar := &mockKeyRegistrar{currentPEM: generateForeignPublicKeyPEM(t)}
	if err := km.EnsureRegistered(context.Background(), registrar); err != nil {
		t.Fatalf("EnsureRegistered() unexpected error: %v", err)
	}

	if len(registrar.registered) != 1 {
		t.Fatalf("expected 1 registration call, got %d", len(registrar.registered))
	}
	if registrar.registered[0] != ourPEM {
		t.Fatal("registered PEM does not match our public key")
	}
}

// KM-005: A failed registration returns an error. Ingest must not proceed.
// If the seeder can't encrypt with our key, downloading and decrypting files will fail anyway.
func TestKeyManager_KM005_RegistrationFailureReturnsError(t *testing.T) {
	km := NewKeyManager(filepath.Join(t.TempDir(), "key.pem"))
	if err := km.EnsureKeyPair(); err != nil {
		t.Fatal(err)
	}

	registrar := &mockKeyRegistrar{
		currentPEM:  generateForeignPublicKeyPEM(t),
		registerErr: errors.New("seeder rejected key"),
	}

	if err := km.EnsureRegistered(context.Background(), registrar); err == nil {
		t.Fatal("expected error when registration fails, got nil")
	}
}

// KM-006: Registration is triggered when the seeder has no key yet (empty response).
// This happens on the very first run against a brand-new seeder instance.
func TestKeyManager_KM006_RegistersWhenSeederHasNoKeyYet(t *testing.T) {
	km := NewKeyManager(filepath.Join(t.TempDir(), "key.pem"))
	if err := km.EnsureKeyPair(); err != nil {
		t.Fatal(err)
	}

	registrar := &mockKeyRegistrar{currentPEM: ""} // seeder has nothing
	if err := km.EnsureRegistered(context.Background(), registrar); err != nil {
		t.Fatalf("EnsureRegistered() unexpected error: %v", err)
	}

	if len(registrar.registered) != 1 {
		t.Fatalf("expected 1 registration call (seeder was empty), got %d", len(registrar.registered))
	}
}
