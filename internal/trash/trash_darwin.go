//go:build darwin

package trash

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func MoveToTrash(paths ...string) error {
	if len(paths) == 0 {
		return nil
	}

	absPaths := make([]string, 0, len(paths))

	for _, path := range paths {
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("resolve %q: %w", path, err)
		}

		// Catch stale scan results / already-deleted files early.
		if _, err := os.Lstat(abs); err != nil {
			return fmt.Errorf("cannot trash %q: %w", abs, err)
		}

		absPaths = append(absPaths, abs)
	}

	const script = `
on run argv
	repeat with p in argv
		set sourcePath to contents of p
		set sourceItem to POSIX file sourcePath

		tell application "Finder"
			delete sourceItem
		end tell
	end repeat
end run
`

	args := append([]string{"-e", script}, absPaths...)

	output, err := exec.Command("osascript", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("move to Trash: %w: %s", err, output)
	}

	return nil
}
