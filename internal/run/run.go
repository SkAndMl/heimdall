package run

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/SkAndMl/heimdall/internal/config"
	"github.com/google/uuid"
)

type RunArgs struct {
	Command []string
	Cwd     string
	Name    string
	Detach  bool
}

type Session struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Cwd     string   `json:"cwd"`
	PID     int      `json:"pid"`
	PGID    int      `json:"pgid"`
	Command []string `json:"command"`
}

func Run(args *RunArgs) error {

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	uuid := uuid.New().ID()
	sessionId := fmt.Sprintf("heim_%d", uuid)
	sessionDir := filepath.Join(home, config.BASE_DIR, "sessions", sessionId)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}

	stdoutFile, err := os.Create(filepath.Join(sessionDir, "stdout.log"))
	if err != nil {
		return err
	}
	stderrFile, err := os.Create(filepath.Join(sessionDir, "stderr.log"))
	if err != nil {
		return err
	}

	filesClosed := false
	defer func() {
		if !filesClosed {
			stdoutFile.Close()
			stderrFile.Close()
		}
	}()

	session := Session{
		ID:      sessionId,
		Name:    args.Name,
		Cwd:     args.Cwd,
		Command: args.Command[:],
	}

	cmd := exec.Command(args.Command[0], args.Command[1:]...)
	if len(args.Cwd) > 0 {
		cmd.Dir = args.Cwd
	}

	if args.Detach {
		cmd.Stdout = stdoutFile
		cmd.Stderr = stderrFile
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, stdoutFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, stderrFile)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	pid := cmd.Process.Pid
	pgid := pid

	session.PID = pid
	session.PGID = pgid

	data, err := json.MarshalIndent(session, "", " ")
	if err != nil {
		return err
	}
	sessionSavePath := filepath.Join(sessionDir, "sessions.json")
	if err := os.WriteFile(sessionSavePath, data, 0644); err != nil {
		return err
	}

	if args.Detach {
		stdoutFile.Close()
		stderrFile.Close()
		filesClosed = true
		return nil
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
