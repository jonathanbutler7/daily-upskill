package ingester

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

var multiSpace = regexp.MustCompile(`\s+`)
var punctuation = regexp.MustCompile(`[^\w\s]`)

// NormalizePayerName trims whitespace, collapses internal spaces, uppercases,
// and removes common punctuation. PRD §9.4.
func NormalizePayerName(raw string) string {
	s := strings.TrimSpace(raw)
	s = multiSpace.ReplaceAllString(s, " ")
	s = punctuation.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	return strings.ToUpper(s)
}

// CardFingerprint returns an HMAC-SHA256 hex fingerprint of the PAN.
// The key should come from config/env — never hardcoded.
// Result is prefixed with "fp:" for easy identification.
func CardFingerprint(pan, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(pan))
	return "fp:" + hex.EncodeToString(mac.Sum(nil))
}

// Last4 returns the last 4 digits of a PAN, or the full string if shorter.
func Last4(pan string) string {
	digits := strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, pan)
	if len(digits) <= 4 {
		return digits
	}
	return digits[len(digits)-4:]
}
