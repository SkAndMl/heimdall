package inspect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/SkAndMl/heimdall/internal/config"
	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
)

func TestFindSessionByRefMatchesExactIDPrefixAndName(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	session := sessionPkg.Session{
		ID:     "heim_aaaaaaaa-1111-1111-1111-aaaaaaaaaaaa",
		Name:   "api",
		Status: sessionPkg.StatusRunning,
	}
	writeInspectSession(t, homeDir, session)

	for _, ref := range []string{session.ID, "heim_aaaaaaaa", "api"} {
		t.Run(ref, func(t *testing.T) {
			found, err := sessionPkg.FindSessionByRef(ref)
			if err != nil {
				t.Fatalf("findSessionByRef returned error: %v", err)
			}
			if found == nil {
				t.Fatal("findSessionByRef returned nil")
			}
			if found.ID != session.ID {
				t.Fatalf("found ID = %q, want %q", found.ID, session.ID)
			}
		})
	}
}

func TestFindSessionByRefRejectsAmbiguousPrefix(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	writeInspectSession(t, homeDir, sessionPkg.Session{
		ID:   "heim_bbbbbbbb-1111-1111-1111-bbbbbbbbbbbb",
		Name: "api",
	})
	writeInspectSession(t, homeDir, sessionPkg.Session{
		ID:   "heim_bbbbbbbb-2222-2222-2222-bbbbbbbbbbbb",
		Name: "worker",
	})

	if _, err := sessionPkg.FindSessionByRef("heim_bbbbbbbb"); err == nil {
		t.Fatal("findSessionByRef returned nil error for ambiguous prefix")
	}
}

func TestFormatInspectOutputIncludesLogsAndProcesses(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	session := sessionPkg.Session{
		ID:        "heim_cccccccc-1111-1111-1111-cccccccccccc",
		Name:      "api",
		Cwd:       "/tmp/project",
		PID:       1234,
		PGID:      1234,
		Command:   []string{"python", "app.py"},
		StartedAt: time.Date(2026, 7, 3, 14, 20, 11, 0, time.Local),
		Status:    sessionPkg.StatusRunning,
	}
	sessionDir := filepath.Join(homeDir, config.BASE_DIR, "sessions", session.ID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("creating session dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "stdout.log"), nil, 0644); err != nil {
		t.Fatalf("creating stdout.log: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "stderr.log"), nil, 0644); err != nil {
		t.Fatalf("creating stderr.log: %v", err)
	}

	output := formatInspectOutput(&session, []process{
		{pid: 1235, command: "python worker.py"},
		{pid: 1236, command: "python watcher.py"},
	})

	for _, want := range []string{
		"Session:", session.ID,
		"Name:", "api",
		"Status:", "running",
		"Started:", "2026-07-03 14:20:11",
		"Process group:", "1234",
		"Working dir:", "/tmp/project",
		"Command:", "python app.py",
		"Logs:",
		"stdout: ~/.heimdall/sessions/" + session.ID + "/stdout.log",
		"stderr: ~/.heimdall/sessions/" + session.ID + "/stderr.log",
		"Other processes in group:",
		"1235", "python worker.py",
		"1236", "python watcher.py",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func writeInspectSession(t *testing.T, homeDir string, session sessionPkg.Session) {
	t.Helper()

	sessionDir := filepath.Join(homeDir, config.BASE_DIR, "sessions", session.ID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		t.Fatalf("creating session dir: %v", err)
	}

	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("marshalling session: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "session.json"), data, 0644); err != nil {
		t.Fatalf("writing session.json: %v", err)
	}
}
