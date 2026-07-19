package util

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Process struct {
	PID     int
	Command string
}

// pgidReuseTolerance covers lstart's 1-second granularity plus the small gap
// between the process starting and session.StartedAt being recorded in run.go.
const pgidReuseTolerance = 2 * time.Second

// ProcessStartTime returns the OS-level start time of the process with the
// given PID by parsing ps lstart output.
func ProcessStartTime(pid int) (time.Time, error) {
	data, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "lstart=").Output()
	if err != nil {
		return time.Time{}, fmt.Errorf("process %d not found: %w", pid, err)
	}
	s := strings.TrimSpace(string(data))
	if s == "" {
		return time.Time{}, fmt.Errorf("process %d not found", pid)
	}
	t, err := time.ParseInLocation("Mon Jan _2 15:04:05 2006", s, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing lstart %q: %w", s, err)
	}
	return t, nil
}

// IsProcessGroupAlive reports whether the process group identified by pgid
// still belongs to the session that started at sessionStartedAt.
// It returns false when the group is empty or when the group leader started
// after sessionStartedAt+pgidReuseTolerance, which indicates the PGID was
// reused by an unrelated process.
func IsProcessGroupAlive(pgid int, sessionStartedAt time.Time) (bool, error) {
	processes, err := FindProcessesInGroup(pgid, false)
	if err != nil {
		return false, err
	}
	if len(processes) == 0 {
		return false, nil
	}

	// The group leader's PID equals its PGID. Compare its OS start time against
	// sessionStartedAt to detect PGID reuse.
	leaderStart, err := ProcessStartTime(pgid)
	if err != nil {
		// Cannot obtain start time; trust group membership.
		return true, nil
	}
	return !leaderStart.After(sessionStartedAt.Add(pgidReuseTolerance)), nil
}

func FindProcessesInGroup(pgid int, excludeGroupHead bool) ([]Process, error) {
	processes := make([]Process, 0)

	data, err := exec.Command(
		"ps",
		"-axww",
		"-o",
		"pid=,pgid=,command=",
	).Output()

	if err != nil {
		return processes, err
	}

	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		if pid == pgid && excludeGroupHead {
			continue // skip group head
		}

		if cpgid, err := strconv.Atoi(fields[1]); err != nil {
			continue
		} else if cpgid != pgid {
			continue
		}

		processes = append(processes, Process{
			PID:     pid,
			Command: strings.Join(fields[2:], " "),
		})
	}
	return processes, nil
}
