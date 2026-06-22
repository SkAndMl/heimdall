package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReturnsUsageCodeWithoutExiting(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"heimdall", "scan"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "expected command: scan <path>") {
		t.Fatalf("stderr = %q, want scan path usage error", stderr.String())
	}
	if !strings.Contains(stderr.String(), "heimdall scan <path>") {
		t.Fatalf("stderr = %q, want scan usage", stderr.String())
	}
}

func TestRunCleanMissingModeShowsCleanUsage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"heimdall", "clean", "~/Desktop1"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "expected command: clean <path> (--dry-run | --interactive)") {
		t.Fatalf("stderr = %q, want clean usage error", stderr.String())
	}
	if !strings.Contains(stderr.String(), "heimdall clean <path> (--dry-run | --interactive)") {
		t.Fatalf("stderr = %q, want clean usage", stderr.String())
	}
	if strings.Contains(stderr.String(), "heimdall scan <path>") {
		t.Fatalf("stderr = %q, did not want scan usage", stderr.String())
	}
}

func TestRunCleanUnknownOptionShowsCleanUsage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"heimdall", "clean", "~", "--force"}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("Run() code = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unknown option "--force"`) {
		t.Fatalf("stderr = %q, want unknown clean option error", stderr.String())
	}
	if !strings.Contains(stderr.String(), "heimdall clean <path> (--dry-run | --interactive)") {
		t.Fatalf("stderr = %q, want clean usage", stderr.String())
	}
}

func TestRunScanWritesReport(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sample.txt"), []byte("sample"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"heimdall", "scan", dir, "--limit", "1"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("Run() code = %d, want 0; stderr = %q", code, stderr.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
	if !strings.Contains(stdout.String(), "◉ HEIMDALL") {
		t.Fatalf("stdout = %q, want scan report", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Disk scan") {
		t.Fatalf("stdout = %q, want scan report title", stdout.String())
	}
	if !strings.Contains(stdout.String(), dir) {
		t.Fatalf("stdout = %q, want scanned path %q", stdout.String(), dir)
	}
}
