package clean

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanReturnsScanError(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing")

	report, err := Clean(Options{Path: missingPath, DryRun: true})

	if err == nil {
		t.Fatal("Clean() error = nil, want scan error")
	}
	if report != "" {
		t.Fatalf("Clean() report = %q, want empty", report)
	}
	if !strings.Contains(err.Error(), "scan cleanup candidates") {
		t.Fatalf("Clean() error = %q, want cleanup scan context", err.Error())
	}
}

func TestCleanDryRunReturnsReport(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "__pycache__")
	if err := os.Mkdir(cacheDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cacheDir, "module.pyc"), []byte("cache"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Clean(Options{Path: dir, DryRun: true})

	if err != nil {
		t.Fatalf("Clean() error = %v, want nil", err)
	}
	if !strings.Contains(report, "◉ HEIMDALL") {
		t.Fatalf("Clean() report = %q, want cleanup plan", report)
	}
	if !strings.Contains(report, "Cleanup plan") {
		t.Fatalf("Clean() report = %q, want cleanup plan title", report)
	}
	if !strings.Contains(report, "Potentially reclaimable") {
		t.Fatalf("Clean() report = %q, want reclaimable summary", report)
	}
	if !strings.Contains(report, "Python bytecode cache") {
		t.Fatalf("Clean() report = %q, want python cache summary", report)
	}
	if !strings.Contains(report, "No files were changed.") {
		t.Fatalf("Clean() report = %q, want dry-run safety copy", report)
	}
}
