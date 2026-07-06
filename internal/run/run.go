package run

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/SkAndMl/heimdall/internal/session"
)

type RunArgs struct {
	Command []string
	Cwd     string
	Name    string
	Detach  bool
}

func HandleRunCommand(args *RunArgs) error {

	runtimeSession, err := session.NewSession(args.Name, args.Cwd, args.Command)
	if err != nil {
		return err
	}
	defer runtimeSession.Close()

	cmd := exec.Command(args.Command[0], args.Command[1:]...)
	if len(args.Cwd) > 0 {
		cmd.Dir = args.Cwd
	}

	if args.Detach {
		cmd.Stdout = runtimeSession.StdoutFile
		cmd.Stderr = runtimeSession.StderrFile
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, runtimeSession.StdoutFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, runtimeSession.StderrFile)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	runtimeSession.SetPID(cmd.Process.Pid)
	runtimeSession.SetPGID(cmd.Process.Pid)
	runtimeSession.Session.StartedAt = time.Now()

	pgid := runtimeSession.GetPGID()

	if err := runtimeSession.SetStatus("running"); err != nil {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
		_ = cmd.Wait()
		return err
	}

	if args.Detach {
		return nil
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
			runtimeSession.SetStatus(session.StatusFailed)
			return err
		}
		return runtimeSession.SetStatus(session.StatusFinished)
	case sig := <-signals:
		runtimeSession.SetStatus(session.StatusStopping)
		if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
			runtimeSession.SetStatus(session.StatusKillFailed)
		}
		err := <-waitCh
		runtimeSession.SetStatus(session.StatusKilled)
		if err != nil {
			return fmt.Errorf("received %s, terminated process group %d: %w", sig, pgid, err)
		}

		return fmt.Errorf("received %s, terminated process group %d", sig, pgid)
	}
}
