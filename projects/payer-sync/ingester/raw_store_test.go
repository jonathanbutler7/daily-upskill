package ingester

import (
	"context"
	"os"
	"testing"
)

func TestLocalRawFileStore_SaveWritesEncryptedBytes(t *testing.T) {
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	store := NewLocalRawFileStore("raw-storage")
	storageKey, err := store.Save(context.Background(), "raw/seeded/seed-42/file_001.835", "seeded/seed-42/file_001.835", []byte("ciphertext"))
	if err != nil {
		t.Fatalf("Save() unexpected error: %v", err)
	}
	if storageKey != "raw-storage/seeded/seed-42/file_001.835" {
		t.Fatalf("storageKey = %q, want raw-storage/seeded/seed-42/file_001.835", storageKey)
	}

	got, err := os.ReadFile(storageKey)
	if err != nil {
		t.Fatalf("ReadFile() unexpected error: %v", err)
	}
	if string(got) != "ciphertext" {
		t.Fatalf("stored bytes = %q, want ciphertext", got)
	}
}

func TestLocalRawFileStore_SaveRejectsPathTraversal(t *testing.T) {
	store := NewLocalRawFileStore("raw-storage")
	if _, err := store.Save(context.Background(), "raw/../../escape.txt", "ignored", []byte("ciphertext")); err == nil {
		t.Fatal("Save() error = nil, want traversal rejection")
	}
}
