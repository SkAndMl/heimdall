package detectors

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/SkAndMl/heimdall/internal/categories"
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

func ClassifyDir(path string) categories.ID {

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return categories.Unknown
	}

	if isPythonVenv(path) {
		return categories.PythonVirtualEnvironment
	}

	if filepath.Base(path) == "__pycache__" {
		return categories.PythonCache
	}

	if filepath.Base(path) == "node_modules" {
		return categories.NodeModules
	}

	if isHuggingFaceCache(path) {
		return categories.HuggingFaceCache
	}

	return categories.Unknown
}
