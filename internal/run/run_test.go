package run

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SkAndMl/heimdall/internal/config"
	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
)

func TestHandleRunCommandCapturesLogsAndFinishesSession(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	err := HandleRunCommand(&RunArgs{
		Name:    "smoke",
		Command: []string{"sh", "-c", "printf stdout-message; printf stderr-message >&2"},
	})
	if err != nil {
		t.Fatalf("HandleRunCommand returned error: %v", err)
	}

	sessionDir := onlySessionDir(t, homeDir)
	data, err := os.ReadFile(filepath.Join(sessionDir, "session.json"))
	if err != nil {
		t.Fatalf("reading session.json: %v", err)
	}

	var saved sessionPkg.Session
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshalling session.json: %v", err)
	}
	if saved.Name != "smoke" {
		t.Fatalf("Name = %q, want smoke", saved.Name)
	}
	if saved.Status != sessionPkg.StatusFinished {
		t.Fatalf("Status = %q, want %q", saved.Status, sessionPkg.StatusFinished)
	}
	if saved.PID == 0 || saved.PGID == 0 {
		t.Fatalf("PID/PGID not set: pid=%d pgid=%d", saved.PID, saved.PGID)
	}

	stdout, err := os.ReadFile(filepath.Join(sessionDir, "stdout.log"))
	if err != nil {
		t.Fatalf("reading stdout.log: %v", err)
	}
	if string(stdout) != "stdout-message" {
		t.Fatalf("stdout.log = %q, want stdout-message", string(stdout))
	}

	stderr, err := os.ReadFile(filepath.Join(sessionDir, "stderr.log"))
	if err != nil {
		t.Fatalf("reading stderr.log: %v", err)
	}
	if string(stderr) != "stderr-message" {
		t.Fatalf("stderr.log = %q, want stderr-message", string(stderr))
	}
}

func onlySessionDir(t *testing.T, homeDir string) string {
	t.Helper()

	sessionsDir := filepath.Join(homeDir, config.BASE_DIR, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		t.Fatalf("reading sessions dir: %v", err)
	}

	sessionDirs := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "heim_") {
			sessionDirs = append(sessionDirs, filepath.Join(sessionsDir, entry.Name()))
		}
	}

	if len(sessionDirs) != 1 {
		t.Fatalf("found %d session dirs, want 1", len(sessionDirs))
	}
	return sessionDirs[0]
}
