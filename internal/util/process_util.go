package util

import (
	"os/exec"
	"strconv"
	"strings"
)

type Process struct {
	PID     int
	Command string
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
