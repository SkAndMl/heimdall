package stop

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
	"github.com/SkAndMl/heimdall/internal/util"
)

var stopOutput io.Writer = os.Stdout

type StopArgs struct {
	SessionRef string
	GraceTime  int
}

// TODO: remove print functions inside HandleStopCommand

func HandleStopCommand(args *StopArgs) error {

	session, err := sessionPkg.FindSessionByRef(args.SessionRef)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session %q not found\n", args.SessionRef)
	}

	if session.Status != sessionPkg.StatusRunning && session.Status != sessionPkg.StatusStopping {
		return fmt.Errorf("session %q is currently not running\n", session.ID)
	}

	if session.PGID <= 0 {
		return fmt.Errorf("session %q invalid process group id %d\n", session.ID, session.PGID)
	}

	printStopping(stopOutput, session)
	if err := syscall.Kill(-session.PGID, syscall.SIGTERM); err != nil {
		return err
	}
	printSentSIGTERM(stopOutput, session.PGID)

	time.Sleep(time.Second * time.Duration(args.GraceTime))

	processes, err := util.FindProcessesInGroup(session.PGID, false)
	if err != nil {
		return err
	}
	if len(processes) > 0 {
		if err := syscall.Kill(-session.PGID, syscall.SIGKILL); err != nil {
			_ = session.SetStatus(sessionPkg.StatusKillFailed)
			return err
		}
	}
	printAllProcessesExited(stopOutput)

	if err := session.SetStatus(sessionPkg.StatusKilled); err != nil {
		return err
	}
	printSessionMarkedStopped(stopOutput)

	return nil
}

func printStopping(writer io.Writer, session *sessionPkg.Session) {
	fmt.Fprintf(writer, "Stopping %s (%s)...\n", session.ID, session.Name)
}

func printSentSIGTERM(writer io.Writer, pgid int) {
	fmt.Fprintf(writer, "Sent SIGTERM to process group %d.\n", pgid)
}

func printAllProcessesExited(writer io.Writer) {
	fmt.Fprintln(writer, "All processes exited.")
}

func printSessionMarkedStopped(writer io.Writer) {
	fmt.Fprintln(writer, "Session marked stopped.")
}
