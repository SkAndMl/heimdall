package run

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
)

type RunArgs struct {
	Command []string
	Cwd     string
	Name    string
	Detach  bool
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

func runInForeground(args *RunArgs) error {
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

	cmd.Stdout = io.MultiWriter(os.Stdout, stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr, stderr)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	session.PID = cmd.Process.Pid
	session.PGID = cmd.Process.Pid
	session.StartedAt = time.Now()

	if err := session.SetStatus("running"); err != nil {
		_ = syscall.Kill(-session.PGID, syscall.SIGTERM)
		_ = cmd.Wait()
		return err
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case err := <-waitCh:
		if err != nil {
			session.SetStatus(sessionPkg.StatusFailed)
			return err
		}
		return session.SetStatus(sessionPkg.StatusFinished)
	case sig := <-signals:
		session.SetStatus(sessionPkg.StatusStopping)
		if err := syscall.Kill(-session.PGID, syscall.SIGTERM); err != nil {
			session.SetStatus(sessionPkg.StatusKillFailed)
		}
		err := <-waitCh
		session.SetStatus(sessionPkg.StatusKilled)
		if err != nil {
			return fmt.Errorf("received %s, terminated process group %d: %w", sig, session.PGID, err)
		}
		return fmt.Errorf("received %s, terminated process group %d", sig, session.PGID)
	}
}

func HandleSupervisorCommand(args *RunArgs) {
	session, err := sessionPkg.NewSession(args.Name, args.Cwd, args.Command)
	if err != nil {
		return
	}

	stdout, stderr, err := session.OpenLogFiles()
	if err != nil {
		return
	}

	defer stdout.Close()
	defer stderr.Close()

	cmd := exec.Command(args.Command[0], args.Command[1:]...)
	if len(args.Cwd) > 0 {
		cmd.Dir = args.Cwd
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return
	}

	session.PID = cmd.Process.Pid
	session.PGID = cmd.Process.Pid
	session.StartedAt = time.Now()

	if err := session.SetStatus("running"); err != nil {
		_ = syscall.Kill(-session.PGID, syscall.SIGTERM)
		_ = cmd.Wait()
		return
	}

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)

	select {
	case err := <-waitCh:
		if err != nil {
			session.SetStatus(sessionPkg.StatusFailed)
			return
		}
		session.SetStatus(sessionPkg.StatusFinished)
		return
	case <-signals:
		session.SetStatus(sessionPkg.StatusStopping)
		if err := syscall.Kill(-session.PGID, syscall.SIGTERM); err != nil {
			session.SetStatus(sessionPkg.StatusKillFailed)
		}
		<-waitCh
		session.SetStatus(sessionPkg.StatusKilled)
	}
}

func HandleRunCommand(args *RunArgs) error {

	if args.Detach {
		return startSupervisor(args)
	}
	return runInForeground(args)
}
