package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpShowsShortQuickStart(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--help"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("run --help exit code = %d, want 0; stderr=%q", exitCode, stderr.String())
	}

	output := stdout.String()
	for _, want := range []string{
		"LedgerDB stress test",
		"Usage: go run ./cmd/stress [flags]",
		"Common runs:",
		"Smoke",
		"Duplicate/idempotency",
		"Most-used flags:",
		"-operations",
		"-max-open-conns",
		"-duplicate-percent",
		"Scoring:",
		"Grades: A>=90, B>=80, C>=70, D>=60, F<60.",
		"go run ./cmd/stress --help-full",
		"--help-full includes sample commands for targeting A/B/C/D/F scores.",
		"docs/stress-test.md",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("help output missing %q:\n%s", want, output)
		}
	}

	for _, notWant := range []string{
		"Core workload flags:",
		"Database flags:",
		"Output and validation flags:",
		"default=postgresql://ledger_db",
		"-format",
	} {
		if strings.Contains(output, notWant) {
			t.Fatalf("short help should not include %q:\n%s", notWant, output)
		}
	}
}

func TestFullHelpShowsGroupedFlags(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--help-full"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("run --help-full exit code = %d, want 0; stderr=%q", exitCode, stderr.String())
	}

	output := stdout.String()
	for _, want := range []string{
		"Good starting point:",
		"Heavier comparison runs:",
		"Core workload flags:",
		"Database flags:",
		"Concurrent request and idempotency flags:",
		"-duplicate-percent",
		"-help-full",
		"How to read the results:",
		"Current scoring system:",
		"Correctness failed: -40",
		"P99 latency: >=100ms -8, >=250ms -15, >=500ms -20, >=1000ms -25",
		"Pool wait per op: >=1ms -5, >=5ms -10, >=10ms -15",
		"How to move scores:",
		"Sample grade targets:",
		"Target A",
		"Target B",
		"Target C",
		"Target D",
		"Target F",
		"-operations",
		"go run ./cmd/stress -accounts=500 -workers=200 -operations=50000 -max-open-conns=10 -think-min=0s -think-max=0s -hot-accounts=2 -hot-transfer-percent=80",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("full help output missing %q:\n%s", want, output)
		}
	}

	for _, notWant := range []string{
		"-dsn",
		"-migrations-dir",
		"-reset",
		"-operation-timeout",
		"-validation-interval",
		"-format",
		"-seed-balance",
		"-max-amount",
	} {
		if strings.Contains(output, notWant) {
			t.Fatalf("full help should not include removed flag %q:\n%s", notWant, output)
		}
	}
}

func TestRemovedFlagsAreRejected(t *testing.T) {
	removedFlags := []string{
		"-dsn=postgresql://example",
		"-migrations-dir=db/migrations",
		"-reset=false",
		"-operation-timeout=1s",
		"-validation-interval=1s",
		"-format=json",
		"-seed-balance=100",
		"-max-amount=10",
	}

	for _, removedFlag := range removedFlags {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		exitCode := run([]string{removedFlag}, &stdout, &stderr)
		if exitCode != 2 {
			t.Fatalf("run %s exit code = %d, want 2", removedFlag, exitCode)
		}
		if !strings.Contains(stderr.String(), "flag provided but not defined") {
			t.Fatalf("run %s stderr missing undefined flag error:\n%s", removedFlag, stderr.String())
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
