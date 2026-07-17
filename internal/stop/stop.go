package stop

import (
	"fmt"
	"syscall"

	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
)

type StopArgs struct {
	SessionRef string
}

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

	if session.RunnerPID <= 0 {
		return fmt.Errorf(
			"session %q has invalid runner PID %d\n",
			session.ID,
			session.RunnerPID,
		)
	}

	return syscall.Kill(session.RunnerPID, syscall.SIGTERM)
}
