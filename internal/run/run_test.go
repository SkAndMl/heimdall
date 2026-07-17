package run

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

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

func TestStopRunningProcessEscalatesAfterGracePeriod(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	readyPath := filepath.Join(t.TempDir(), "ready")
	cmd := exec.Command("sh", "-c", `trap '' TERM; printf ready > "$1"; while :; do sleep 1; done`, "sh", readyPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting command: %v", err)
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	deadline := time.Now().Add(time.Second)
	for {
		if _, err := os.Stat(readyPath); err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
			<-waitCh
			t.Fatal("command did not install SIGTERM handler")
		}
		time.Sleep(10 * time.Millisecond)
	}

	session, err := sessionPkg.NewSession("stubborn", "", []string{"sh"})
	if err != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-waitCh
		t.Fatalf("NewSession returned error: %v", err)
	}
	session.PID = cmd.Process.Pid
	session.PGID = cmd.Process.Pid
	session.Status = sessionPkg.StatusRunning
	if err := session.Save(); err != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		<-waitCh
		t.Fatalf("saving session: %v", err)
	}

	gracePeriod := 50 * time.Millisecond
	startedAt := time.Now()
	if err := stopRunningProcess(session, waitCh, gracePeriod); err != nil {
		t.Fatalf("stopRunningProcess returned error: %v", err)
	}
	if elapsed := time.Since(startedAt); elapsed < gracePeriod {
		t.Fatalf("stopRunningProcess returned after %v, before grace period %v", elapsed, gracePeriod)
	}

	saved, err := sessionPkg.FindSessionByRef(session.ID)
	if err != nil {
		t.Fatalf("finding session: %v", err)
	}
	if saved.Status != sessionPkg.StatusKilled {
		t.Fatalf("Status = %q, want %q", saved.Status, sessionPkg.StatusKilled)
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
