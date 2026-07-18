package ps

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

func TestGetSessionsDefaultsToRunningSessions(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	writeSession(t, homeDir, sessionPkg.Session{
		ID:     "heim_11111111-1111-1111-1111-111111111111",
		Name:   "api",
		Status: sessionPkg.StatusRunning,
	})
	writeSession(t, homeDir, sessionPkg.Session{
		ID:     "heim_22222222-2222-2222-2222-222222222222",
		Name:   "tests",
		Status: sessionPkg.StatusFinished,
	})

	sessions, err := getSessions(&PsArgs{})
	if err != nil {
		t.Fatalf("getSessions returned error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}
	if sessions[0].Name != "api" {
		t.Fatalf("session name = %q, want api", sessions[0].Name)
	}
}

func TestGetSessionsFiltersByExplicitStatus(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	writeSession(t, homeDir, sessionPkg.Session{
		ID:     "heim_33333333-3333-3333-3333-333333333333",
		Name:   "api",
		Status: sessionPkg.StatusRunning,
	})
	writeSession(t, homeDir, sessionPkg.Session{
		ID:     "heim_44444444-4444-4444-4444-444444444444",
		Name:   "failed-job",
		Status: sessionPkg.StatusFailed,
	})

	sessions, err := getSessions(&PsArgs{Status: sessionPkg.StatusFailed})
	if err != nil {
		t.Fatalf("getSessions returned error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}
	if sessions[0].Name != "failed-job" {
		t.Fatalf("session name = %q, want failed-job", sessions[0].Name)
	}
}

func TestFormatPsOutputTable(t *testing.T) {
	startedAt := time.Now().Add(-4 * time.Minute)
	output, err := formatPsOutput([]sessionPkg.Session{{
		ID:        "heim_abc",
		Name:      "api",
		Status:    sessionPkg.StatusRunning,
		PID:       1234,
		StartedAt: startedAt,
		Command:   []string{"uvicorn", "app:app", "--port", "8000"},
	}}, false)
	if err != nil {
		t.Fatalf("formatPsOutput returned error: %v", err)
	}

	for _, want := range []string{"ID", "NAME", "STATUS", "PID", "AGE", "COMMAND", "heim_abc", "api", "running", "1234", "uvicorn app:app --port 8000"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output missing %q:\n%s", want, output)
		}
	}
}

func TestFormatPsOutputJSON(t *testing.T) {
	output, err := formatPsOutput([]sessionPkg.Session{{
		ID:      "heim_abc",
		Name:    "api",
		Status:  sessionPkg.StatusRunning,
		Command: []string{"go", "test", "./..."},
	}}, true)
	if err != nil {
		t.Fatalf("formatPsOutput returned error: %v", err)
	}

	var sessions []sessionPkg.Session
	if err := json.Unmarshal([]byte(output), &sessions); err != nil {
		t.Fatalf("JSON output did not unmarshal: %v\n%s", err, output)
	}
	if len(sessions) != 1 {
		t.Fatalf("len(sessions) = %d, want 1", len(sessions))
	}
	if sessions[0].Command[0] != "go" {
		t.Fatalf("command = %#v, want go test ./...", sessions[0].Command)
	}
}

func TestGetSessionsReconcilesStaleRunningSession(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	writeSession(t, homeDir, sessionPkg.Session{
		ID:     "heim_55555555-5555-5555-5555-555555555555",
		Name:   "stale",
		Status: sessionPkg.StatusRunning,
		PID:    99999999,
		PGID:   99999999,
	})

	sessions, err := getSessions(&PsArgs{})
	if err != nil {
		t.Fatalf("getSessions returned error: %v", err)
	}
	if len(sessions) != 0 {
		t.Fatalf("len(sessions) = %d, want 0 (stale session should be reconciled away)", len(sessions))
	}
}

func writeSession(t *testing.T, homeDir string, session sessionPkg.Session) {
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
