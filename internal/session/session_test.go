package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/SkAndMl/heimdall/internal/config"
)

func TestNewSessionCreatesFilesAndMetadata(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	session, err := NewSession("api", "/tmp/project", []string{"python", "app.py"})
	if err != nil {
		t.Fatalf("NewSession returned error: %v", err)
	}

	if session.ID == "" {
		t.Fatal("session ID is empty")
	}
	if session.Name != "api" {
		t.Fatalf("Name = %q, want api", session.Name)
	}
	if session.Status != StatusNotStarted {
		t.Fatalf("Status = %q, want %q", session.Status, StatusNotStarted)
	}

	sessionPath := filepath.Join(sessionDir(homeDir, session.ID), "session.json")
	if _, err := os.Stat(sessionPath); err != nil {
		t.Fatalf("session.json not created: %v", err)
	}

	var saved Session
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("reading session.json: %v", err)
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshalling session.json: %v", err)
	}
	if saved.ID != session.ID {
		t.Fatalf("saved ID = %q, want %q", saved.ID, session.ID)
	}
}

func TestSetStatusPersistsStatus(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	session, err := NewSession("worker", "", []string{"sleep", "1"})
	if err != nil {
		t.Fatalf("NewSession returned error: %v", err)
	}

	if err := session.SetStatus(StatusRunning); err != nil {
		t.Fatalf("SetStatus returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(sessionDir(homeDir, session.ID), "session.json"))
	if err != nil {
		t.Fatalf("reading session.json: %v", err)
	}

	var saved Session
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshalling session.json: %v", err)
	}
	if saved.Status != StatusRunning {
		t.Fatalf("saved Status = %q, want %q", saved.Status, StatusRunning)
	}
}

func TestNewSessionUsesConfiguredBaseDir(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	session, err := NewSession("api", "", []string{"true"})
	if err != nil {
		t.Fatalf("NewSession returned error: %v", err)
	}

	wantPrefix := filepath.Join(homeDir, config.BASE_DIR, "sessions")
	gotSessionDir := sessionDir(homeDir, session.ID)
	if filepath.Dir(gotSessionDir) != wantPrefix {
		t.Fatalf("SessionDir = %q, want parent %q", gotSessionDir, wantPrefix)
	}
}

func TestOpenLogFilesCreatesWritableLogs(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	session, err := NewSession("api", "", []string{"true"})
	if err != nil {
		t.Fatalf("NewSession returned error: %v", err)
	}

	stdoutFile, stderrFile, err := session.OpenLogFiles()
	if err != nil {
		t.Fatalf("OpenLogFiles returned error: %v", err)
	}
	defer stdoutFile.Close()
	defer stderrFile.Close()

	if _, err := stdoutFile.WriteString("stdout-message"); err != nil {
		t.Fatalf("writing stdout log: %v", err)
	}
	if _, err := stderrFile.WriteString("stderr-message"); err != nil {
		t.Fatalf("writing stderr log: %v", err)
	}

	stdoutPath, err := session.StdoutPath()
	if err != nil {
		t.Fatalf("StdoutPath returned error: %v", err)
	}
	stderrPath, err := session.StdErrPath()
	if err != nil {
		t.Fatalf("StdErrPath returned error: %v", err)
	}

	stdout, err := os.ReadFile(stdoutPath)
	if err != nil {
		t.Fatalf("reading stdout log: %v", err)
	}
	if string(stdout) != "stdout-message" {
		t.Fatalf("stdout.log = %q, want stdout-message", string(stdout))
	}

	stderr, err := os.ReadFile(stderrPath)
	if err != nil {
		t.Fatalf("reading stderr log: %v", err)
	}
	if string(stderr) != "stderr-message" {
		t.Fatalf("stderr.log = %q, want stderr-message", string(stderr))
	}
}

func sessionDir(homeDir, sessionID string) string {
	return filepath.Join(homeDir, config.BASE_DIR, "sessions", sessionID)
}
