package inspect

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/SkAndMl/heimdall/internal/config"
	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
	"github.com/SkAndMl/heimdall/internal/util"
)

type InspectArgs struct {
	SessionRef string
}

type logFile struct {
	label string
	path  string
}

func formatInspectOutput(session *sessionPkg.Session, childProcesses []util.Process) string {
	if session == nil {
		return ""
	}

	var output bytes.Buffer
	writer := tabwriter.NewWriter(&output, 0, 0, 4, ' ', 0)

	fmt.Fprintf(writer, "Session:\t%s\n", session.ID)
	fmt.Fprintf(writer, "Name:\t%s\n", session.Name)
	fmt.Fprintf(writer, "Status:\t%s\n", session.Status)
	fmt.Fprintf(writer, "Started:\t%s\n", formatStartedAt(session.StartedAt))
	fmt.Fprintf(writer, "PID:\t%d\n", session.PID)
	fmt.Fprintf(writer, "Process group:\t%d\n", session.PGID)
	fmt.Fprintf(writer, "Working dir:\t%s\n", session.Cwd)
	fmt.Fprintf(writer, "Command:\t%s\n", strings.Join(session.Command, " "))
	fmt.Fprintln(writer, "Logs:")

	writer.Flush()

	for _, logFile := range findLogFiles(session.ID) {
		fmt.Fprintf(&output, "  %s: %s\n", logFile.label, logFile.path)
	}

	fmt.Fprintln(&output)
	fmt.Fprintln(&output, "Other processes in group:")

	childWriter := tabwriter.NewWriter(&output, 0, 0, 4, ' ', 0)
	fmt.Fprintln(childWriter, "PID\tCOMMAND")
	for _, process := range childProcesses {
		fmt.Fprintf(childWriter, "%d\t%s\n", process.PID, process.Command)
	}
	childWriter.Flush()

	return strings.TrimRight(output.String(), "\n")
}

func findLogFiles(sessionID string) []logFile {
	logNames := []string{"stdout", "stderr"}
	logFiles := make([]logFile, 0, len(logNames))

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return logFiles
	}

	for _, logName := range logNames {
		fileName := logName + ".log"
		absolutePath := filepath.Join(homeDir, config.BASE_DIR, "sessions", sessionID, fileName)
		if _, err := os.Stat(absolutePath); err != nil {
			continue
		}

		logFiles = append(logFiles, logFile{
			label: logName,
			path:  filepath.Join("~", config.BASE_DIR, "sessions", sessionID, fileName),
		})
	}

	return logFiles
}

func formatStartedAt(startedAt time.Time) string {
	if startedAt.IsZero() {
		return "-"
	}

	return startedAt.Local().Format("2006-01-02 15:04:05")
}

func HandleInspectCommand(args *InspectArgs) error {
	session, err := sessionPkg.FindSessionByRef(args.SessionRef)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session %q not found", args.SessionRef)
	}

	if session.Status == sessionPkg.StatusRunning && session.PGID > 0 {
		if alive, err := util.IsProcessGroupAlive(session.PGID); err == nil && !alive {
			_ = session.SetStatus(sessionPkg.StatusFinished)
		}
	}

	childProcesses, err := util.FindProcessesInGroup(session.PGID, true)
	if err != nil {
		return err
	}

	fmt.Println(formatInspectOutput(session, childProcesses))
	return nil
}
