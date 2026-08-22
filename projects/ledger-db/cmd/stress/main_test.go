package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpShowsStressScenariosAndFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--help"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("run --help exit code = %d, want 0; stderr=%q", exitCode, stderr.String())
	}

	output := stdout.String()
	for _, want := range []string{
		"LedgerDB stress test",
		"Good starting point:",
		"Heavier comparison runs:",
		"Settlement contention check",
		"Hot-account skew check",
		"intentionally creates row-lock contention",
		"Core workload flags:",
		"-accounts",
		"-max-open-conns",
		"-settlement-buckets",
		"-hot-transfer-percent",
		"How to read the results:",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("help output missing %q:\n%s", want, output)
		}
	}
}

func TestShortHelpAlias(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"-h"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("run -h exit code = %d, want 0; stderr=%q", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("short help did not print usage:\n%s", stdout.String())
	}
}
