package cli

import (
	"reflect"
	"testing"

	"github.com/SkAndMl/heimdall/internal/session"
)

func TestParseRunArgs(t *testing.T) {
	args, err := ParseRunArgs([]string{
		"heimdall", "run",
		"--cwd", "/tmp",
		"--name", "api",
		"--detach",
		"--", "python", "app.py",
	})
	if err != nil {
		t.Fatalf("ParseRunArgs returned error: %v", err)
	}

	if args.Cwd != "/tmp" {
		t.Fatalf("Cwd = %q, want /tmp", args.Cwd)
	}
	if args.Name != "api" {
		t.Fatalf("Name = %q, want api", args.Name)
	}
	if !args.Detach {
		t.Fatal("Detach = false, want true")
	}
	if !reflect.DeepEqual(args.Command, []string{"python", "app.py"}) {
		t.Fatalf("Command = %#v, want python app.py", args.Command)
	}
}

func TestParseRunArgsRequiresCommand(t *testing.T) {
	if _, err := ParseRunArgs([]string{"heimdall", "run", "--name", "api"}); err == nil {
		t.Fatal("ParseRunArgs returned nil error for missing command")
	}
}

func TestParsePsArgs(t *testing.T) {
	args, err := ParsePsArgs([]string{"heimdall", "ps", "--all", "--status", "failed", "--json"})
	if err != nil {
		t.Fatalf("ParsePsArgs returned error: %v", err)
	}

	if !args.All {
		t.Fatal("All = false, want true")
	}
	if args.Status != session.StatusFailed {
		t.Fatalf("Status = %q, want %q", args.Status, session.StatusFailed)
	}
	if !args.JSONOutput {
		t.Fatal("JSONOutput = false, want true")
	}
}

func TestParseInspectArgs(t *testing.T) {
	args, err := ParseInspectArgs([]string{"heimdall", "inspect", "heim_abc"})
	if err != nil {
		t.Fatalf("ParseInspectArgs returned error: %v", err)
	}
	if args.SessionRef != "heim_abc" {
		t.Fatalf("SessionRef = %q, want heim_abc", args.SessionRef)
	}
}

func TestParseStopArgs(t *testing.T) {
	args, err := ParseStopArgs([]string{"heimdall", "stop", "heim_abc"})
	if err != nil {
		t.Fatalf("ParseStopArgs returned error: %v", err)
	}
	if args.SessionRef != "heim_abc" {
		t.Fatalf("SessionRef = %q, want heim_abc", args.SessionRef)
	}
}

func TestParseStopArgsRejectsGraceFlag(t *testing.T) {
	if _, err := ParseStopArgs([]string{"heimdall", "stop", "heim_abc", "--grace", "5"}); err == nil {
		t.Fatal("ParseStopArgs returned nil error for removed --grace flag")
	}
}
