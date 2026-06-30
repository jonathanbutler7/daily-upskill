package ingester

import "testing"

func TestNormalizePayerName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already clean", "DELTA DENTAL OF CALIFORNIA", "DELTA DENTAL OF CALIFORNIA"},
		{"mixed case", "Delta Dental of California", "DELTA DENTAL OF CALIFORNIA"},
		{"leading trailing spaces", "  Delta Dental of California  ", "DELTA DENTAL OF CALIFORNIA"},
		{"extra internal spaces", "Delta  Dental  of  California", "DELTA DENTAL OF CALIFORNIA"},
		{"trailing period", "Delta Dental of California.", "DELTA DENTAL OF CALIFORNIA"},
		{"punctuation inside", "Delta, Dental. of California!", "DELTA DENTAL OF CALIFORNIA"},
		{"seeder format with dots", "  Delta  Dental of California.  ", "DELTA DENTAL OF CALIFORNIA"},
		{"all normalize together", "  delta  dental, of california.  ", "DELTA DENTAL OF CALIFORNIA"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePayerName(tt.input)
			if got != tt.want {
				t.Errorf("NormalizePayerName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCardFingerprint(t *testing.T) {
	fp1 := CardFingerprint("4111111111111111", "test-key")
	fp2 := CardFingerprint("4111111111111111", "test-key")
	fp3 := CardFingerprint("4222222222222222", "test-key")

	if fp1 != fp2 {
		t.Error("same PAN+key should produce same fingerprint")
	}
	if fp1 == fp3 {
		t.Error("different PANs should produce different fingerprints")
	}
	if fp1 == "4111111111111111" {
		t.Error("fingerprint must not be the raw PAN")
	}
	if len(fp1) < 10 {
		t.Error("fingerprint looks too short")
	}
}

func TestLast4(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"4111111111111111", "1111"},
		{"4222222222222222", "2222"},
		{"1234", "1234"},
		{"123", "123"},
		{"4111-1111-1111-1234", "1234"},
	}
	for _, tt := range tests {
		got := Last4(tt.input)
		if got != tt.want {
			t.Errorf("Last4(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
