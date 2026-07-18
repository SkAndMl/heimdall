package run

import (
	"bufio"
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

func manageRunningProcess(cmd *exec.Cmd, session *sessionPkg.Session, logMode runLogMode) error {

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

	if logMode == SupervisorMode {
		readyPipe := os.NewFile(3, "ready-pipe")
		if readyPipe == nil {
			_ = syscall.Kill(-session.PGID, syscall.SIGTERM)
			<-waitCh
			return fmt.Errorf("ready-pipe not available\n")
		}
		if _, err := fmt.Fprintln(readyPipe, "READY"); err != nil {
			_ = readyPipe.Close()
			_ = syscall.Kill(-session.PGID, syscall.SIGTERM)
			<-waitCh
			return fmt.Errorf("writing supervisor readiness: %w", err)
		}
		if err := readyPipe.Close(); err != nil {
			_ = syscall.Kill(-session.PGID, syscall.SIGTERM)
			<-waitCh
			return fmt.Errorf("closing supervisor readiness pipe: %w", err)
		}
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

func run(session *sessionPkg.Session, logMode runLogMode) error {
	if len(session.Command) == 0 {
		return fmt.Errorf("session %q has no command", session.ID)
	}

	stdout, stderr, err := session.OpenLogFiles()
	if err != nil {
		return err
	}

	defer stdout.Close()
	defer stderr.Close()

	cmd := exec.Command(session.Command[0], session.Command[1:]...)
	if len(session.Cwd) > 0 {
		cmd.Dir = session.Cwd
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

	return manageRunningProcess(cmd, session, logMode)
}

func startSupervisor(session *sessionPkg.Session) error {

	readyReader, readyWriter, err := os.Pipe()
	if err != nil {
		return err
	}

	defer readyReader.Close()
	defer readyWriter.Close()

	executable, err := os.Executable()
	if err != nil {
		return err
	}

	supervisorArgs := []string{
		"_run-supervisor",
		session.ID,
	}

	cmd := exec.Command(executable, supervisorArgs...)
	cmd.ExtraFiles = []*os.File{readyWriter}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	cleanupSupervisor := func() {
		_ = cmd.Process.Signal(syscall.SIGTERM)
		_ = cmd.Wait()
	}

	if err := readyWriter.Close(); err != nil {
		cleanupSupervisor()
		return fmt.Errorf("closing launcher readiness pipe: %w", err)
	}

	reader := bufio.NewReader(readyReader)
	message, err := reader.ReadString('\n')
	if err != nil {
		cleanupSupervisor()
		return err
	}

	if message != "READY\n" {
		cleanupSupervisor()
		return fmt.Errorf("Not able to start workload\n")
	}

	if err := cmd.Process.Release(); err != nil {
		cleanupSupervisor()
		return err
	}

	return nil
}

func HandleSupervisorCommand(sessionRef string) error {
	session, err := sessionPkg.FindSessionByRef(sessionRef)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session %q not found\n", sessionRef)
	}

	if session.Status != sessionPkg.StatusNotStarted {
		return fmt.Errorf("session %q has been corrupted\n", session.ID)
	}

	return run(session, SupervisorMode)
}

func HandleRunCommand(args *RunArgs) (string, error) {

	session, err := sessionPkg.NewSession(
		args.Name,
		args.Cwd,
		args.Command,
	)
	if err != nil {
		return "", err
	}

	if args.Detach {
		if err := startSupervisor(session); err != nil {
			return "", err
		}
		return session.ID, nil
	}
	return "", run(session, ForegroundMode)
}
