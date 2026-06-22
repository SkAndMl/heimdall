package scan

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SkAndMl/heimdall/internal/categories"
)

func TestNewScannerStoresFindingCategory(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "project.tar.gz")
	if err := os.WriteFile(archivePath, []byte("archive"), 0o644); err != nil {
		t.Fatal(err)
	}

	scanner, err := NewScanner(dir, -1)
	if err != nil {
		t.Fatal(err)
	}

	findings := scanner.Categories[categories.Archive]
	if len(findings) != 1 {
		t.Fatalf("archive findings = %d, want 1", len(findings))
	}
	if findings[0].Path != archivePath {
		t.Fatalf("archive path = %q, want %q", findings[0].Path, archivePath)
	}
	if findings[0].Category != categories.Archive {
		t.Fatalf("archive category = %q, want %q", findings[0].Category, categories.Archive)
	}
}
