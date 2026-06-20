package detectors

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/SkAndMl/heimdall/internal/util"
)

func isPythonVenv(path string) bool {
	cfgPath := filepath.Join(path, "pyvenv.cfg")
	cfgInfo, err := os.Stat(cfgPath)
	if err != nil || cfgInfo.IsDir() {
		return false
	}

	pythonPath := ""
	if util.CurrentOS() == "windows" {
		pythonPath = filepath.Join(path, "Scripts", "python.exe")
	} else {
		pythonPath = filepath.Join(path, "bin", "python")
	}

	if info, err := os.Stat(pythonPath); err == nil && !info.IsDir() {
		return true
	}
	return false
}

func isHuggingFaceCache(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		name := entry.Name()
		isRepoDir := strings.HasPrefix(name, "models--") || strings.HasPrefix(name, "datasets--") || strings.HasPrefix(name, "spaces--")
		if !isRepoDir {
			continue
		}

		repoPath := filepath.Join(path, name)
		if info, err := os.Stat(filepath.Join(repoPath, "snapshots")); err == nil && info.IsDir() {
			return true
		}
	}
	return false

}

func ClassifyDir(path string) string {

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return "unknown"
	}

	if isPythonVenv(path) {
		return "python_virtual_environment"
	}

	if filepath.Base(path) == "__pycache__" {
		return "python_cache"
	}

	if filepath.Base(path) == "node_modules" {
		return "node_modules"
	}

	if isHuggingFaceCache(path) {
		return "huggingface_cache"
	}

	return "unknown"
}

func ClassifyFile(path string) string {
	name := strings.ToLower(filepath.Base(path))
	switch {
	case strings.HasSuffix(name, ".dmg"),
		strings.HasSuffix(name, ".pkg"),
		strings.HasSuffix(name, ".mpkg"),
		strings.HasSuffix(name, ".msi"),
		strings.HasSuffix(name, ".exe"),
		strings.HasSuffix(name, ".deb"),
		strings.HasSuffix(name, ".rpm"),
		strings.HasSuffix(name, ".appimage"):
		return "installer"
	case strings.HasSuffix(name, ".tar.gz"),
		strings.HasSuffix(name, ".tar.bz2"),
		strings.HasSuffix(name, ".tgz"),
		strings.HasSuffix(name, ".tar.xz"),
		strings.HasSuffix(name, ".zip"),
		strings.HasSuffix(name, ".7z"),
		strings.HasSuffix(name, ".rar"),
		strings.HasSuffix(name, ".tar"),
		strings.HasSuffix(name, ".gz"),
		strings.HasSuffix(name, ".bz2"),
		strings.HasSuffix(name, ".xz"):
		return "archive"
	}

	return "unknown"
}
