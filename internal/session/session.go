package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	RunnerPID int       `json:"runner_pid"`
	Command   []string  `json:"command"`
	StartedAt time.Time `json:"started_at"`
	Status    Status    `json:"status"`
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

	session := &Session{
		ID:      sessionId,
		Name:    name,
		Cwd:     cwd,
		Command: command,
		Status:  StatusNotStarted,
	}

	if err := session.Save(); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Session) Save() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	sessionDir := filepath.Join(home, config.BASE_DIR, "sessions", s.ID)
	data, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(sessionDir, "session.json"), data, 0644); err != nil {
		return err
	}
	return nil
}

func (s *Session) SetStatus(status Status) error {
	s.Status = status
	return s.Save()
}


func (s *Session) StdoutPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, config.BASE_DIR, "sessions", s.ID, "stdout.log"), nil
}

func (s *Session) StdErrPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, config.BASE_DIR, "sessions", s.ID, "stderr.log"), nil
}

func (s *Session) OpenLogFiles() (*os.File, *os.File, error) {
	stdoutPath, err := s.StdoutPath()
	if err != nil {
		return nil, nil, err
	}

	stderrPath, err := s.StdErrPath()
	if err != nil {
		return nil, nil, err
	}

	stdoutFile, err := os.Create(stdoutPath)
	if err != nil {
		return nil, nil, err
	}
	stderrFile, err := os.Create(stderrPath)
	if err != nil {
		stdoutFile.Close()
		return nil, nil, err
	}

	return stdoutFile, stderrFile, nil

}

func FindSessionByRef(sessionRef string) (*Session, error) {

	dirRe := regexp.MustCompile(`^heim_[a-z0-9-]+$`)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sessionsDir := filepath.Join(homeDir, config.BASE_DIR, "sessions")
	info, err := os.Lstat(sessionsDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("Sessions dir does not exist")
	}

	dirEntries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return nil, err
	}

	matchedSessionsByIDPrefix := make([]Session, 0)
	matchedSessionByName := make([]Session, 0)

	for _, entry := range dirEntries {
		var session Session

		info, err := os.Lstat(filepath.Join(sessionsDir, entry.Name()))
		if err != nil || !info.IsDir() || !dirRe.MatchString(entry.Name()) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name(), "session.json"))
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		if session.ID == sessionRef {
			return &session, nil
		} else if strings.HasPrefix(session.ID, sessionRef) {
			matchedSessionsByIDPrefix = append(matchedSessionsByIDPrefix, session)
		} else if session.Name == sessionRef {
			matchedSessionByName = append(matchedSessionByName, session)
		}
	}

	if len(matchedSessionsByIDPrefix) == 1 {
		return &matchedSessionsByIDPrefix[0], nil
	}
	if len(matchedSessionsByIDPrefix) > 1 {
		return nil, fmt.Errorf("More than one session matched\n")
	}

	if len(matchedSessionByName) == 1 {
		return &matchedSessionByName[0], nil
	}
	if len(matchedSessionByName) > 1 {
		return nil, fmt.Errorf("More than one session matched\n")
	}

	return nil, nil
}
