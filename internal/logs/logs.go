package logs

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	sessionPkg "github.com/SkAndMl/heimdall/internal/session"
)

type LogArgs struct {
	SessionRef string
	StderrFlag bool
	LastNLines int
	FollowFlag bool
}

func getLastNLines(file *os.File, n int) ([]string, error) {
	lines := make([]string, 0)
	reader := bufio.NewReader(file)

	var lineString strings.Builder

	for {
		fragment, err := reader.ReadString('\n')
		if len(fragment) > 0 {
			lineString.WriteString(fragment)
			if strings.HasSuffix(fragment, "\n") {
				lines = append(lines, lineString.String())
				lineString.Reset()
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return []string{}, err
		}
	}
	if lineString.Len() > 0 {
		lines = append(lines, lineString.String())
	}

	if n == 0 {
		return lines, nil
	}

	return lines[max(0, len(lines)-n):], nil
}

func followLogFile(file *os.File, n int) error {
	lastNLines, err := getLastNLines(file, n)
	if err != nil {
		return err
	}
	for _, line := range lastNLines {
		fmt.Print(line)
	}

	reader := bufio.NewReader(file)
	var pendingLine strings.Builder

	for {
		fragment, err := reader.ReadString('\n')
		if len(fragment) > 0 {
			pendingLine.WriteString(fragment)
		}
		if err == nil && pendingLine.Len() > 0 {
			fmt.Print(pendingLine.String())
			pendingLine.Reset()
		}

		if errors.Is(err, io.EOF) {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if err != nil {
			return err
		}
	}

}

func HandleLogsCommand(args *LogArgs) error {
	session, err := sessionPkg.FindSessionByRef(args.SessionRef)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("Cannot find session %q\n", args.SessionRef)
	}

	fileToOpen := ""
	if args.StderrFlag {
		fileToOpen, err = session.StdErrPath()
		if err != nil {
			return err
		}
	} else {
		fileToOpen, err = session.StdoutPath()
		if err != nil {
			return err
		}
	}

	file, err := os.Open(fileToOpen)
	if err != nil {
		return err
	}
	defer file.Close()

	if !args.FollowFlag {
		lines, err := getLastNLines(file, args.LastNLines)
		if err != nil {
			return err
		}
		for _, line := range lines {
			fmt.Print(line)
		}
	} else {
		return followLogFile(file, args.LastNLines)
	}

	return nil
}
