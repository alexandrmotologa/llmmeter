package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCLIDemoAndReport(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "llmmeter-cli-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "cli_test.db")

	// Test cmdDemo
	cmdDemo([]string{"--db", dbPath})

	// Test cmdReport
	cmdReport([]string{"--db", dbPath})

	// Test cmdExportNegativeMargins
	outCSV := filepath.Join(tmpDir, "neg.csv")
	cmdExportNegativeMargins([]string{"--db", dbPath, "--output", outCSV})

	if _, err := os.Stat(outCSV); os.IsNotExist(err) {
		t.Errorf("expected CSV file to be generated at %s", outCSV)
	}

	// Test cmdSyncPricing
	cmdSyncPricing(nil)
}
