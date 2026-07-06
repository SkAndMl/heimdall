# Heimdall

Heimdall is a small process runner that records command sessions under
`~/.heimdall/sessions`. It can run commands in the foreground or detach them,
list recorded sessions, and inspect a session's metadata, logs, and process
group.

## Installation

Build the CLI from the repository root:

```sh
go build -o heimdall ./cmd/heimdall
```

For local development, you can also run it without building:

```sh
go run ./cmd/heimdall ps
```

The CLI currently supports `run`, `ps`, and `inspect`.

## Run A Command

```sh
heimdall run [flags] -- command [args...]
```

Flags:

- `--cwd`, `-C`: working directory for the command
- `--name`, `-n`: human-readable session name
- `--detach`, `-d`: start the command and return immediately

Examples:

```sh
heimdall run -n api -- python -m http.server 8000
heimdall run -n api -d -- uvicorn app:app --port 8000
heimdall run -n tests -- go test ./...
```

Each run creates a session directory:

```text
~/.heimdall/sessions/<session-id>/
  session.json
  stdout.log
  stderr.log
```

## List Sessions

```sh
heimdall ps [flags]
```

By default, `ps` shows running sessions.

Flags:

- `--all`: include sessions in all statuses
- `--status <status>`: include sessions with a specific status
- `--json`: print machine-readable JSON

Examples:

```sh
heimdall ps
heimdall ps --all
heimdall ps --status failed
heimdall ps --all --json
```

Text output:

```text
ID                                      NAME    STATUS     PID     AGE    COMMAND
heim_478baf61-0212-424d-a32e-8a5db5c939fa api  running    4843    4m     python app.py
```

## Inspect A Session

```sh
heimdall inspect <session-ref>
```

`<session-ref>` can be:

- a full session ID
- a unique session ID prefix
- a session name, if it matches exactly one session

Example:

```sh
heimdall inspect api
heimdall inspect heim_478baf61
```

Inspect output includes session metadata, available log files, and other
processes in the same process group:

```text
Session:          heim_478baf61-0212-424d-a32e-8a5db5c939fa
Name:             api
Status:           running
Started:          2026-07-03 14:20:11
PID:              4843
Process group:    4843
Working dir:      /Users/sathya/code/talos
Command:          uvicorn app:app --port 8000
Logs:
  stdout: ~/.heimdall/sessions/heim_478baf61-0212-424d-a32e-8a5db5c939fa/stdout.log
  stderr: ~/.heimdall/sessions/heim_478baf61-0212-424d-a32e-8a5db5c939fa/stderr.log

Other processes in group:
PID      COMMAND
4844     python worker.py
4845     python watcher.py
```

## Session Statuses

Heimdall currently records these statuses:

- `not_started`
- `running`
- `stopping`
- `finished`
- `failed`
- `killed`
- `kill_failed`

Detached sessions are recorded as `running` after the process starts. Heimdall
does not yet reconcile stored status after an externally terminated process.

## Development

Run the test suite:

```sh
go test ./...
```

Run static checks:

```sh
go vet ./...
```

The tests use temporary `HOME` directories for session storage, so they do not
write to your real `~/.heimdall` directory.
