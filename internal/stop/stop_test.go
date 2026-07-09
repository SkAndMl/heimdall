package stop

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/SkAndMl/heimdall/internal/config"
	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
	"github.com/SkAndMl/heimdall/internal/util"
)

func TestHandleStopCommandStopsProcessGroupAndPreservesSessionFiles(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	cmd := exec.Command("sh", "-c", "trap '' TERM; exec tail -f /dev/null")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting test process: %v", err)
	}
	pgid := cmd.Process.Pid
	defer cleanupProcessGroup(cmd, pgid)

	waitForProcessGroup(t, pgid)

	session, err := sessionPkg.NewSession("api", "", []string{"tail", "-f", "/dev/null"})
	if err != nil {
		t.Fatalf("NewSession returned error: %v", err)
	}
	session.PID = cmd.Process.Pid
	session.PGID = pgid
	session.Status = sessionPkg.StatusRunning
	if err := session.Save(); err != nil {
		t.Fatalf("saving session: %v", err)
	}

	stdoutPath, stderrPath := createSessionLogs(t, session)

	var output bytes.Buffer
	previousOutput := stopOutput
	stopOutput = &output
	t.Cleanup(func() {
		stopOutput = previousOutput
	})

	if err := HandleStopCommand(&StopArgs{SessionRef: session.ID, GraceTime: 0}); err != nil {
		t.Fatalf("HandleStopCommand returned error: %v", err)
	}

	waitForCommandExit(t, cmd)
	assertProcessGroupGone(t, pgid)

	saved := readSavedSession(t, homeDir, session.ID)
	if saved.Status != sessionPkg.StatusKilled {
		t.Fatalf("saved Status = %q, want %q", saved.Status, sessionPkg.StatusKilled)
	}

	assertFileContains(t, stdoutPath, "stdout stays available")
	assertFileContains(t, stderrPath, "stderr stays available")

	wantOutput := "Stopping " + session.ID + " (api)...\n" +
		"Sent SIGTERM to process group " + strconv.Itoa(pgid) + ".\n" +
		"All processes exited.\n" +
		"Session marked stopped.\n"
	if output.String() != wantOutput {
		t.Fatalf("stop output = %q, want %q", output.String(), wantOutput)
	}
}

func createSessionLogs(t *testing.T, session *sessionPkg.Session) (string, string) {
	t.Helper()

	stdoutFile, stderrFile, err := session.OpenLogFiles()
	if err != nil {
		t.Fatalf("OpenLogFiles returned error: %v", err)
	}
	defer stdoutFile.Close()
	defer stderrFile.Close()

	if _, err := stdoutFile.WriteString("stdout stays available"); err != nil {
		t.Fatalf("writing stdout log: %v", err)
	}
	if _, err := stderrFile.WriteString("stderr stays available"); err != nil {
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
	return stdoutPath, stderrPath
}

func cleanupProcessGroup(cmd *exec.Cmd, pgid int) {
	if cmd.Process != nil {
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		_ = cmd.Wait()
	}
}

func waitForProcessGroup(t *testing.T, pgid int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		processes, err := util.FindProcessesInGroup(pgid, false)
		if err == nil && len(processes) > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("process group %d did not appear", pgid)
}

func waitForCommandExit(t *testing.T, cmd *exec.Cmd) {
	t.Helper()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		var exitErr *exec.ExitError
		if err != nil && !errors.As(err, &exitErr) {
			t.Fatalf("waiting for command: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("process group leader did not exit after stop")
	}
}

func assertProcessGroupGone(t *testing.T, pgid int) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		processes, err := util.FindProcessesInGroup(pgid, false)
		if err == nil && len(processes) == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	t.Fatalf("process group %d still has running processes", pgid)
}

func readSavedSession(t *testing.T, homeDir, sessionID string) sessionPkg.Session {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(homeDir, config.BASE_DIR, "sessions", sessionID, "session.json"))
	if err != nil {
		t.Fatalf("reading session.json: %v", err)
	}

	var saved sessionPkg.Session
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshalling session.json: %v", err)
	}
	return saved
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("%s = %q, want to contain %q", path, string(data), want)
	}
}
