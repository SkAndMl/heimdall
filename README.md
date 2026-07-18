# Heimdall

Heimdall is a small Unix process runner that keeps a durable record of every
command it starts. It can run commands in the foreground or under a detached
supervisor, list and inspect sessions, read their logs, and stop an entire
process group.

Session metadata and logs are stored under `~/.heimdall/sessions`, so they
remain available after a command exits.

## Requirements

- Go 1.26.1 or newer
- A Unix-like system with process groups, Unix signals, and the `ps` command

Heimdall currently relies on Unix-specific process APIs and is not supported on
Windows.

## Installation

Build the binary from the repository root:

```sh
go build -o heimdall ./cmd/heimdall
```

Move `heimdall` somewhere on your `PATH`, or invoke the CLI during development
with `go run`:

```sh
go run ./cmd/heimdall ps --all
```

## Quick Start

Start a named command in the background:

```sh
heimdall run --name web --detach -- python3 -m http.server 8000
```

The detached command prints its session ID:

```text
Session ID: heim_478baf61-0212-424d-a32e-8a5db5c939fa
```

You can then inspect, follow, and stop it using its name, full ID, or a unique
ID prefix:

```sh
heimdall inspect web
heimdall logs web --follow
heimdall stop web
```

## Commands

### `run`

```text
heimdall run [flags] -- command [args...]
```

The `--` separator is required. Everything after it is passed to the command
unchanged.

| Flag | Short | Description |
| --- | --- | --- |
| `--cwd <path>` | `-C <path>` | Run from the given working directory |
| `--name <name>` | `-n <name>` | Give the session a human-readable name |
| `--detach` | `-d` | Start under a background supervisor and return immediately |

Examples:

```sh
# Run in the foreground. Output is displayed and saved to the session logs.
heimdall run --name tests -- go test ./...

# Run from another directory.
heimdall run -C ./examples -- python3 app.py

# Run in the background. Output is written only to the session logs.
heimdall run -n api -d -- uvicorn app:app --port 8000
```

Foreground mode waits for the command to exit. `Ctrl-C` asks Heimdall to stop
the whole managed process group. Detached mode launches a supervisor which
tracks the command and updates its session status after the invoking terminal
returns.

### `ps`

```text
heimdall ps [flags]
```

Without flags, `ps` lists sessions whose recorded status is `running`.

| Flag | Description |
| --- | --- |
| `--all` | List sessions in every status |
| `--status <status>` | List sessions with exactly this status |
| `--json` | Emit the selected sessions as a JSON array |

Examples:

```sh
heimdall ps
heimdall ps --all
heimdall ps --status failed
heimdall ps --all --json
```

Text output has the following shape:

```text
ID                                          NAME    STATUS     PID     AGE    COMMAND
heim_478baf61-0212-424d-a32e-8a5db5c939fa api     running    4843    4m     python3 app.py
```

### `logs`

```text
heimdall logs <session-ref> [flags]
```

By default, `logs` prints the complete standard-output log and exits.

| Flag | Description |
| --- | --- |
| `--stderr` | Read the standard-error log instead of standard output |
| `--tail <lines>` | Start with the last positive number of lines |
| `--follow` | Continue waiting for appended log output |

The flags can be combined:

```sh
heimdall logs api
heimdall logs api --stderr
heimdall logs heim_478baf61 --tail 50
heimdall logs api --tail 20 --follow
```

Stop follow mode with `Ctrl-C`.

### `inspect`

```text
heimdall inspect <session-ref>
```

`inspect` displays the stored session metadata, available log paths, and other
processes currently found in the session's process group.

```text
Session:          heim_478baf61-0212-424d-a32e-8a5db5c939fa
Name:             api
Status:           running
Started:          2026-07-03 14:20:11
PID:              4843
Process group:    4843
Working dir:      /Users/example/project
Command:          uvicorn app:app --port 8000
Logs:
  stdout: ~/.heimdall/sessions/heim_478baf61-0212-424d-a32e-8a5db5c939fa/stdout.log
  stderr: ~/.heimdall/sessions/heim_478baf61-0212-424d-a32e-8a5db5c939fa/stderr.log

Other processes in group:
PID     COMMAND
4844    python worker.py
```

### `stop`

```text
heimdall stop <session-ref>
```

`stop` signals the session's runner. The runner marks the session as stopping,
sends `SIGTERM` to the managed process group, waits up to two seconds, and then
sends `SIGKILL` if the group has not exited. Metadata and logs are retained.

Only sessions recorded as `running` or `stopping` can be stopped.

## Session References

Commands that accept `<session-ref>` resolve it as one of:

- an exact session ID;
- a unique session ID prefix; or
- a session name that belongs to exactly one session.

Reusing a name is allowed, but that name becomes ambiguous if more than one
stored session has it. Use an ID or a sufficiently long unique ID prefix in
that case.

## Session Data

Each run creates a directory like:

```text
~/.heimdall/sessions/<session-id>/
├── session.json
├── stdout.log
└── stderr.log
```

`session.json` records the ID, optional name, working directory, command,
process IDs, start time, and current status. The log files are created when the
command is started.

Heimdall records these statuses:

| Status | Meaning |
| --- | --- |
| `not_started` | The session exists but its command has not started |
| `running` | The command is running |
| `stopping` | Termination has been requested |
| `finished` | The command exited successfully |
| `failed` | The command returned an error or non-zero exit status |
| `killed` | Heimdall terminated the command |
| `kill_failed` | Heimdall could not signal the process group |

The statuses are stored metadata rather than a live process scan. If Heimdall's
runner itself is forcibly terminated, a stale status may remain.

## Development

Run the test suite and static checks from the repository root:

```sh
go test ./...
go vet ./...
```

The tests replace `HOME` with temporary directories, so they do not write
session data to your real `~/.heimdall` directory.

The code is organized by CLI feature:

```text
cmd/heimdall/       command entry point
internal/cli/       argument parsing
internal/run/       foreground and detached process execution
internal/session/   session persistence and reference resolution
internal/ps/        session listing
internal/logs/      log reading and following
internal/inspect/   session and process-group inspection
internal/stop/      process-group shutdown
```
