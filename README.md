# Heimdall

Heimdall is a disk usage scanner and cleanup helper. It scans a path, reports the largest files and directories, explains common cleanup categories, and can move selected cleanup candidates to the system Trash.

## Install

```sh
go install github.com/SkAndMl/heimdall/cmd/heimdall@latest
```

For local development:

```sh
go run ./cmd/heimdall scan .
go test ./...
```

## Usage

```sh
heimdall scan <path> [--json] [--explain] [--max-depth <depth>] [--limit <count>]
heimdall clean <path> (--dry-run | --interactive)
```

Examples:

```sh
heimdall scan ~
heimdall scan ~ --json
heimdall scan ~ --explain
heimdall scan ~ --max-depth 2
heimdall scan ~ --limit 25
heimdall clean ~ --dry-run
heimdall clean ~/Downloads --interactive
```

## Scan

`scan` walks the target path and prints a report with:

- total scanned size
- files and directories scanned
- skipped paths and warnings
- largest directories
- largest files

Use `--json` for machine-readable output:

```sh
heimdall scan ~/Downloads --json
```

Use `--max-depth` to limit traversal depth:

```sh
heimdall scan ~/Desktop --max-depth 2
```

Use `--limit` to control how many largest files/directories are shown:

```sh
heimdall scan ~ --limit 10
```

## Explain

`--explain` summarizes detected cleanup categories with risk, reason, and suggested action.

```sh
heimdall scan ~ --explain
```

Example output:

```text
8.3 GB   Installers and archives
Risk:     Usually safe
Why:      Detected old DMG, PKG, ZIP, and TAR files.
Action:   Review installers that are no longer needed.
```

`--json --explain` returns the same category summary as JSON.

## Clean

`clean` scans for cleanup candidates. It does not delete files directly; selected files are moved to the system Trash.

Dry run:

```sh
heimdall clean ~ --dry-run
```

Interactive cleanup:

```sh
heimdall clean ~ --interactive
```

The interactive mode lets you select candidates with the keyboard:

```text
up/down or j/k: move | space: toggle | y: confirm | q: quit
```

Cleanup candidates currently include:

- Python bytecode cache (`__pycache__`)
- Python virtual environments
- Node dependencies (`node_modules`)
- Hugging Face cache directories
- installers and archives (`.dmg`, `.pkg`, `.zip`, `.tar`, `.gz`, etc.)

## Safety Notes

Heimdall is conservative by design:

- `clean --dry-run` only prints a cleanup plan.
- `clean --interactive` requires explicit selection.
- cleanup moves selected paths to Trash instead of permanently deleting them.
- permission errors and symlinks are reported as warnings.

On macOS, scanning protected locations may require Full Disk Access for your terminal.

## Development

Run tests:

```sh
go test ./...
```

Run the CLI locally:

```sh
go run ./cmd/heimdall scan .
go run ./cmd/heimdall clean . --dry-run
```

## License

MIT
