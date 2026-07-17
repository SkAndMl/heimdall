package run

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
)

type runLogMode string

const (
	ForegroundMode  runLogMode = "foreground"
	SupervisorMode  runLogMode = "supervisor"
	stopGracePeriod            = 2 * time.Second
)

type RunArgs struct {
	Command []string
	Cwd     string
	Name    string
	Detach  bool
}

func stopRunningProcess(session *sessionPkg.Session, waitCh <-chan error, gracePeriod time.Duration) error {
	statusErr := session.SetStatus(sessionPkg.StatusStopping)

	if err := syscall.Kill(-session.PGID, syscall.SIGTERM); err != nil {
		statusErr = errors.Join(statusErr, session.SetStatus(sessionPkg.StatusKillFailed))
		<-waitCh
		return errors.Join(err, statusErr)
	}

	timer := time.NewTimer(gracePeriod)
	defer timer.Stop()

	select {
	case <-waitCh:
	case <-timer.C:
		if err := syscall.Kill(-session.PGID, syscall.SIGKILL); err != nil {
			statusErr = errors.Join(statusErr, session.SetStatus(sessionPkg.StatusKillFailed))
			<-waitCh
			return errors.Join(err, statusErr)
		}
		<-waitCh
	}

	return errors.Join(statusErr, session.SetStatus(sessionPkg.StatusKilled))
}

func manageRunningProcess(cmd *exec.Cmd, session *sessionPkg.Session) error {

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	// set StatusRunning after registering signals because of possible race
	// with stop subcommand
	if err := session.SetStatus(sessionPkg.StatusRunning); err != nil {
		_ = syscall.Kill(-session.PGID, syscall.SIGTERM)
		<-waitCh
		return err
	}

	select {
	case err := <-waitCh:
		if err != nil {
			return errors.Join(err, session.SetStatus(sessionPkg.StatusFailed))
		}
		return session.SetStatus(sessionPkg.StatusFinished)

	case <-signals:
		return stopRunningProcess(session, waitCh, stopGracePeriod)
	}
}

func run(args *RunArgs, logMode runLogMode) error {
	session, err := sessionPkg.NewSession(args.Name, args.Cwd, args.Command)
	if err != nil {
		return err
	}

	stdout, stderr, err := session.OpenLogFiles()
	if err != nil {
		return err
	}

	defer stdout.Close()
	defer stderr.Close()

	cmd := exec.Command(args.Command[0], args.Command[1:]...)
	if len(args.Cwd) > 0 {
		cmd.Dir = args.Cwd
	}

	switch logMode {
	case ForegroundMode:
		cmd.Stdout = io.MultiWriter(os.Stdout, stdout)
		cmd.Stderr = io.MultiWriter(os.Stderr, stderr)
	case SupervisorMode:
		cmd.Stdout = stdout
		cmd.Stderr = stderr
	default:
		return fmt.Errorf("Unrecognized logMode: %s\n", logMode)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	session.PID = cmd.Process.Pid
	session.PGID = cmd.Process.Pid
	session.RunnerPID = os.Getpid()
	session.StartedAt = time.Now()

	return manageRunningProcess(cmd, session)
}

func startSupervisor(args *RunArgs) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	supervisorArgs := []string{
		"_run-supervisor",
		"--cwd", args.Cwd,
		"--name", args.Name,
		"--",
	}
	supervisorArgs = append(supervisorArgs, args.Command...)

	cmd := exec.Command(executable, supervisorArgs...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if err := cmd.Process.Release(); err != nil {
		return err
	}

	return nil
}

func HandleSupervisorCommand(args *RunArgs) error {
	return run(args, SupervisorMode)
}

func HandleRunCommand(args *RunArgs) error {
	if args.Detach {
		return startSupervisor(args)
	}
	return run(args, ForegroundMode)
}
