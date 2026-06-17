package util

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var essentialSystemSkips = map[string][]string{
	"darwin": {
		"/dev",
		"/Volumes",
		"/.Spotlight-V100",
		"/.fseventsd",
		"/.Trashes",
		"/private/var/vm",
	},
	"linux": {
		"/proc",
		"/sys",
		"/dev",
		"/run",
	},

	"windows": {
		`C:\System Volume Information`,
		`C:\$Recycle.Bin`,
		`C:\Recovery`,
	},
}

func CurrentOS() string {
	return runtime.GOOS
}

func ShouldSkip(path string) bool {
	clean := filepath.Clean(path)
	skipPaths := essentialSystemSkips[CurrentOS()]

	for _, skipPath := range skipPaths {
		if clean == skipPath || strings.HasPrefix(clean, skipPath+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}
