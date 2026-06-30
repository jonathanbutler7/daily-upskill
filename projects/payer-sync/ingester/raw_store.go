package ingester

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// LocalRawFileStore keeps the original encrypted payload on disk.
// For now this is a repo-local development implementation, not object storage.
type LocalRawFileStore struct {
	baseDir string
}

func NewLocalRawFileStore(baseDir string) *LocalRawFileStore {
	return &LocalRawFileStore{baseDir: baseDir}
}

func (s *LocalRawFileStore) Save(_ context.Context, preferredKey, sourceKey string, encryptedBytes []byte) (string, error) {
	relPath, err := localRawPath(preferredKey, sourceKey)
	if err != nil {
		return "", err
	}

	targetPath := filepath.Join(s.baseDir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return "", fmt.Errorf("mkdir raw storage path: %w", err)
	}
	if err := os.WriteFile(targetPath, encryptedBytes, 0o600); err != nil {
		return "", fmt.Errorf("write raw file: %w", err)
	}
	return filepath.ToSlash(targetPath), nil
}

func localRawPath(preferredKey, sourceKey string) (string, error) {
	for _, candidate := range []string{preferredKey, sourceKey} {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		normalized, err := normalizeRawPath(candidate)
		if err != nil {
			return "", err
		}
		if strings.HasPrefix(normalized, "raw/") {
			return strings.TrimPrefix(normalized, "raw/"), nil
		}
		return normalized, nil
	}
	return "", fmt.Errorf("raw storage key is empty")
}

func normalizeRawPath(p string) (string, error) {
	clean := path.Clean(strings.TrimPrefix(strings.TrimSpace(filepath.ToSlash(p)), "/"))
	if clean == "." || clean == "" {
		return "", fmt.Errorf("invalid raw storage path %q", p)
	}
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("raw storage path escapes base directory: %q", p)
	}
	return clean, nil
}
