package util

import (
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestProcessStartTimeReturnsReasonableTimeForLiveProcess(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting process: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })

	before := time.Now()
	got, err := ProcessStartTime(cmd.Process.Pid)
	if err != nil {
		t.Fatalf("ProcessStartTime returned error: %v", err)
	}

	// lstart has 1-second granularity, so allow 2 seconds of slack before.
	if got.Before(before.Add(-2 * time.Second)) {
		t.Fatalf("start time %v is more than 2s before process was started (%v)", got, before)
	}
	if got.After(before.Add(time.Second)) {
		t.Fatalf("start time %v is after now (%v)", got, before)
	}
}

func TestProcessStartTimeErrorsForNonExistentPID(t *testing.T) {
	if _, err := ProcessStartTime(99999999); err == nil {
		t.Fatal("ProcessStartTime returned nil error for non-existent PID")
	}
}

func TestIsProcessGroupAliveReturnsTrueForLiveGroup(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting process: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })

	sessionStartedAt := time.Now()

	alive, err := IsProcessGroupAlive(cmd.Process.Pid, sessionStartedAt)
	if err != nil {
		t.Fatalf("IsProcessGroupAlive returned error: %v", err)
	}
	if !alive {
		t.Fatal("IsProcessGroupAlive returned false for a live process group")
	}
}

func TestIsProcessGroupAliveReturnsFalseForNonExistentGroup(t *testing.T) {
	alive, err := IsProcessGroupAlive(99999999, time.Now())
	if err != nil {
		t.Fatalf("IsProcessGroupAlive returned error: %v", err)
	}
	if alive {
		t.Fatal("IsProcessGroupAlive returned true for non-existent group")
	}
}

func TestIsProcessGroupAliveDetectsPGIDReuse(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting process: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })

	// Simulate a stale session: the recorded start time is 1 hour in the past,
	// but the process in this group just started — so the PGID was reused.
	staleStartedAt := time.Now().Add(-1 * time.Hour)

	alive, err := IsProcessGroupAlive(cmd.Process.Pid, staleStartedAt)
	if err != nil {
		t.Fatalf("IsProcessGroupAlive returned error: %v", err)
	}
	if alive {
		t.Fatal("IsProcessGroupAlive returned true for a reused PGID")
	}
}
