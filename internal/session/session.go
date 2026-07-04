package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SkAndMl/heimdall/internal/config"
	"github.com/google/uuid"
)

type Status string

const (
	StatusNotStarted Status = "not_started"
	StatusRunning    Status = "running"
	StatusStopping   Status = "stopping"
	StatusFinished   Status = "finished"
	StatusFailed     Status = "failed"
	StatusKilled     Status = "killed"
	StatusKillFailed Status = "kill_failed"
)

type Session struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Cwd       string    `json:"cwd"`
	PID       int       `json:"pid"`
	PGID      int       `json:"pgid"`
	Command   []string  `json:"command"`
	StartedAt time.Time `json:"started_at"`
	Status    Status    `json:"status"`
}

type RuntimeSession struct {
	Session    *Session
	SessionDir string
	StdoutFile *os.File
	StderrFile *os.File
}

func NewSession(name string, cwd string, command []string) (*RuntimeSession, error) {

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
		ID:      sessionId,
		Name:    name,
		Cwd:     cwd,
		Command: command,
		Status:  StatusNotStarted,
	}

	runtimeSession := &RuntimeSession{
		Session:    session,
		SessionDir: sessionDir,
		StdoutFile: stdoutFile,
		StderrFile: stderrFile,
	}

	if err := runtimeSession.SaveSession(); err != nil {
		stdoutFile.Close()
		stderrFile.Close()
		return nil, err
	}

	return runtimeSession, nil
}

func (r *RuntimeSession) SaveSession() error {
	data, err := json.MarshalIndent(r.Session, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(r.SessionDir, "session.json"), data, 0644); err != nil {
		return err
	}
	return nil
}

func (r *RuntimeSession) SetStatus(status Status) error {
	r.Session.Status = status
	return r.SaveSession()
}

func (r *RuntimeSession) Close() {
	r.StdoutFile.Close()
	r.StderrFile.Close()
}

func (r *RuntimeSession) GetPGID() int {
	return r.Session.PGID
}

func (r *RuntimeSession) SetPID(pid int) {
	r.Session.PID = pid
}

func (r *RuntimeSession) SetPGID(pgid int) {
	r.Session.PGID = pgid
}
