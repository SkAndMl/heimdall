package ps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/SkAndMl/heimdall/internal/config"
	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
)

type PsArgs struct {
	All        bool
	Status     sessionPkg.Status
	JSONOutput bool
}

func getSessions(args *PsArgs) ([]sessionPkg.Session, error) {
	dirRe := regexp.MustCompile(`heim_[a-z0-9-]+`)
	sessions := make([]sessionPkg.Session, 0)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return sessions, err
	}

	sessionsDir := filepath.Join(homeDir, config.BASE_DIR, "sessions")
	info, err := os.Lstat(sessionsDir)
	if err != nil {
		return sessions, err
	}

	if !info.IsDir() {
		return sessions, nil
	}

	dirEntries, err := os.ReadDir(sessionsDir)
	if err != nil {
		return sessions, err
	}

	for _, entry := range dirEntries {
		entryInfo, err := os.Lstat(filepath.Join(sessionsDir, entry.Name()))
		if err != nil {
			continue
		}
		if !entryInfo.IsDir() || !dirRe.MatchString(entry.Name()) {
			continue
		}

		sessionData, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name(), "session.json"))
		if err != nil {
			continue
		}

		session := sessionPkg.Session{}
		if err := json.Unmarshal(sessionData, &session); err != nil {
			continue
		}

		_ = session.Reconcile()

		if args.Status != "" {
			if session.Status == args.Status {
				sessions = append(sessions, session)
			}
			continue
		}

		if args.All || session.Status == sessionPkg.StatusRunning {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

func formatPsOutput(sessions []sessionPkg.Session, jsonOutput bool) (string, error) {
	if jsonOutput {
		data, err := json.Marshal(sessions)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	var output bytes.Buffer
	writer := tabwriter.NewWriter(&output, 0, 0, 4, ' ', 0)

	fmt.Fprintln(writer, "ID\tNAME\tSTATUS\tPID\tAGE\tCOMMAND")
	for _, session := range sessions {
		fmt.Fprintf(
			writer,
			"%s\t%s\t%s\t%d\t%s\t%s\n",
			session.ID,
			session.Name,
			session.Status,
			session.PID,
			formatAge(session.StartedAt),
			strings.Join(session.Command, " "),
		)
	}

	writer.Flush()
	return strings.TrimRight(output.String(), "\n"), nil
}

func formatAge(startedAt time.Time) string {
	if startedAt.IsZero() {
		return "-"
	}

	age := time.Since(startedAt)
	age = max(0, age)

	switch {
	case age < time.Minute:
		return fmt.Sprintf("%ds", int(age.Seconds()))
	case age < time.Hour:
		return fmt.Sprintf("%dm", int(age.Minutes()))
	case age < 24*time.Hour:
		return fmt.Sprintf("%dh", int(age.Hours()))
	default:
		return fmt.Sprintf("%dd", int(age.Hours()/24))
	}
}

func HandlePsCommand(args *PsArgs) error {
	sessions, err := getSessions(args)
	if err != nil {
		return err
	}

	toPrint, err := formatPsOutput(sessions, args.JSONOutput)
	if err != nil {
		return err
	}
	fmt.Println(toPrint)

	return nil
}
