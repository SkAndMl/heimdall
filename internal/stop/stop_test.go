package stop

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"

	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
)

func TestHandleStopCommandSignalsRunner(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	defer signal.Stop(signals)

	session, err := sessionPkg.NewSession("api", "", []string{"sleep", "10"})
	if err != nil {
		t.Fatalf("NewSession returned error: %v", err)
	}
	session.RunnerPID = os.Getpid()
	session.Status = sessionPkg.StatusRunning
	if err := session.Save(); err != nil {
		t.Fatalf("saving session: %v", err)
	}

	if err := HandleStopCommand(&StopArgs{SessionRef: session.ID}); err != nil {
		t.Fatalf("HandleStopCommand returned error: %v", err)
	}

	select {
	case sig := <-signals:
		if sig != syscall.SIGTERM {
			t.Fatalf("received signal %v, want SIGTERM", sig)
		}
	case <-time.After(time.Second):
		t.Fatal("runner did not receive SIGTERM")
	}

	saved, err := sessionPkg.FindSessionByRef(session.ID)
	if err != nil {
		t.Fatalf("finding session: %v", err)
	}
	if saved.Status != sessionPkg.StatusRunning {
		t.Fatalf("Status = %q, want %q", saved.Status, sessionPkg.StatusRunning)
	}
}

func TestHandleStopCommandRejectsInvalidRunnerPID(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	session, err := sessionPkg.NewSession("legacy", "", []string{"sleep", "10"})
	if err != nil {
		t.Fatalf("NewSession returned error: %v", err)
	}
	session.Status = sessionPkg.StatusRunning
	if err := session.Save(); err != nil {
		t.Fatalf("saving session: %v", err)
	}

	err = HandleStopCommand(&StopArgs{SessionRef: session.ID})
	if err == nil || !strings.Contains(err.Error(), "invalid runner PID 0") {
		t.Fatalf("HandleStopCommand error = %v, want invalid runner PID", err)
	}
}
