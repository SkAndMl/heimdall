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

	runtimeSession, err := NewSession("api", "/tmp/project", []string{"python", "app.py"})
	if err != nil {
		t.Fatalf("NewSession returned error: %v", err)
	}
	defer runtimeSession.Close()

	if runtimeSession.Session.ID == "" {
		t.Fatal("session ID is empty")
	}
	if runtimeSession.Session.Name != "api" {
		t.Fatalf("Name = %q, want api", runtimeSession.Session.Name)
	}
	if runtimeSession.Session.Status != StatusNotStarted {
		t.Fatalf("Status = %q, want %q", runtimeSession.Session.Status, StatusNotStarted)
	}

	sessionPath := filepath.Join(runtimeSession.SessionDir, "session.json")
	if _, err := os.Stat(sessionPath); err != nil {
		t.Fatalf("session.json not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(runtimeSession.SessionDir, "stdout.log")); err != nil {
		t.Fatalf("stdout.log not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(runtimeSession.SessionDir, "stderr.log")); err != nil {
		t.Fatalf("stderr.log not created: %v", err)
	}

	var saved Session
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("reading session.json: %v", err)
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshalling session.json: %v", err)
	}
	if saved.ID != runtimeSession.Session.ID {
		t.Fatalf("saved ID = %q, want %q", saved.ID, runtimeSession.Session.ID)
	}
}

func TestSetStatusPersistsStatus(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	runtimeSession, err := NewSession("worker", "", []string{"sleep", "1"})
	if err != nil {
		t.Fatalf("NewSession returned error: %v", err)
	}
	defer runtimeSession.Close()

	if err := runtimeSession.SetStatus(StatusRunning); err != nil {
		t.Fatalf("SetStatus returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(runtimeSession.SessionDir, "session.json"))
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

	runtimeSession, err := NewSession("api", "", []string{"true"})
	if err != nil {
		t.Fatalf("NewSession returned error: %v", err)
	}
	defer runtimeSession.Close()

	wantPrefix := filepath.Join(homeDir, config.BASE_DIR, "sessions")
	if filepath.Dir(runtimeSession.SessionDir) != wantPrefix {
		t.Fatalf("SessionDir = %q, want parent %q", runtimeSession.SessionDir, wantPrefix)
	}
}
