package run

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SkAndMl/heimdall/internal/config"
	"github.com/google/uuid"
)

type Session struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Cwd        string    `json:"cwd"`
	PID        int       `json:"pid"`
	PGID       int       `json:"pgid"`
	Command    []string  `json:"command"`
	StartedAt  time.Time `json:"started_at"`
	Status     string    `json:"status"`
	SessionDir string    `json:"-"`
	StdoutFile *os.File  `json:"-"`
	StderrFile *os.File  `json:"-"`
}

func NewSession(name string, cwd string, command []string) (*Session, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sessionId := fmt.Sprintf("heim_%s", uuid.NewString())
	sessionDir := filepath.Join(home, config.BASE_DIR, "sessions", sessionId)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, err
	}

	stdoutFile, err := os.Create(filepath.Join(sessionDir, "stdout.log"))
	if err != nil {
		return nil, err
	}
	stderrFile, err := os.Create(filepath.Join(sessionDir, "stderr.log"))
	if err != nil {
		stdoutFile.Close()
		return nil, err
	}

	session := &Session{
		ID:         sessionId,
		Name:       name,
		Cwd:        cwd,
		Command:    command,
		SessionDir: sessionDir,
		Status:     "not_started",
		StdoutFile: stdoutFile,
		StderrFile: stderrFile,
	}

	if err := session.SaveSession(); err != nil {
		stdoutFile.Close()
		stderrFile.Close()
		return nil, err
	}

	return session, nil
}

func (s *Session) SaveSession() error {
	data, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(s.SessionDir, "session.json"), data, 0644); err != nil {
		return err
	}
	return nil
}

func (s *Session) SetStatus(status string) error {
	s.Status = status
	return s.SaveSession()
}

func (s *Session) Close() {
	s.StdoutFile.Close()
	s.StderrFile.Close()
}
