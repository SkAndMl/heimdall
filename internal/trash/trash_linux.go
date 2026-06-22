//go:build linux

package trash

import (
	"fmt"
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

		absPaths = append(absPaths, abs)
	}

	gioPath, err := exec.LookPath("gio")
	if err != nil {
		return fmt.Errorf(
			"could not find `gio`; install GLib/GIO to use the desktop Trash: %w",
			err,
		)
	}

	args := append([]string{"trash"}, absPaths...)

	output, err := exec.Command(gioPath, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("move to Trash: %w: %s", err, output)
	}

	return nil
}
