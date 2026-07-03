package run

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type RunArgs struct {
	Command []string
	Cwd     string
	Name    string
	Detach  bool
}

func Run(args *RunArgs) error {

	session, err := NewSession(args.Name, args.Cwd, args.Command)
	if err != nil {
		return err
	}
	defer session.Close()

	cmd := exec.Command(args.Command[0], args.Command[1:]...)
	if len(args.Cwd) > 0 {
		cmd.Dir = args.Cwd
	}

	if args.Detach {
		cmd.Stdout = session.StdoutFile
		cmd.Stderr = session.StderrFile
	} else {
		cmd.Stdout = io.MultiWriter(os.Stdout, session.StdoutFile)
		cmd.Stderr = io.MultiWriter(os.Stderr, session.StderrFile)
	}

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
			session.SetStatus("failed")
			return err
		}
		return session.SetStatus("finished")
	case sig := <-signals:
		session.SetStatus("stopping")
		if err := syscall.Kill(-session.PGID, syscall.SIGTERM); err != nil {
			session.SetStatus("kill_failed")
		}
		err := <-waitCh
		session.SetStatus("killed")
		if err != nil {
			return fmt.Errorf("received %s, terminated process group %d: %w", sig, session.PGID, err)
		}

		return fmt.Errorf("received %s, terminated process group %d", sig, session.PGID)
	}
}
