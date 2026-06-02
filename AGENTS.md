# Repository Guidelines

## Project Overview

nix-spotlight is a macOS CLI tool that creates Spotlight-indexable trampoline apps for Nix-installed `.app` bundles. It creates a directory app with a `Contents` symlink pointing to the Nix store app, then optionally updates the macOS Dock to point at the trampoline instead of the volatile `/nix/store/` path.

## Architecture & Data Flow

```
cmd/nix-spotlight/main.go          (CLI: flag parsing, dispatch)
        │
        ▼
internal/nixspotlight/             (private library, package nixspotlight)
  ├── app.go          App struct — wraps .app bundle path
  ├── trampoline.go   GatherApps, CreateTrampoline, SyncTrampolines
  └── dock.go         SyncDock, DockSyncResult, runDockutil
```

**Flow**: `main.go` parses `sync <from> <to> [--no-dock]` → `SyncTrampolines(fromDir, toDir)` (validates, gathers apps, creates trampoline tree, bumps mtime) → `SyncDock(trampolines, "")` (runs `dockutil -L`, swaps `/nix/store/` paths for trampoline paths).

**Module**: `github.com/anntnzrb/nix-spotlight` (Go 1.26). Zero external deps — stdlib only. Internal package boundary enforced by `internal/` directory.

**No interfaces, no DI** — direct function calls and struct value types throughout.

## Key Directories

| Path | Purpose |
|---|---|
| `cmd/nix-spotlight/` | CLI binary entrypoint (`package main`) |
| `internal/nixspotlight/` | Private library (`package nixspotlight`) |
| `nix/` | Nix build infra: derivation, devshell, formatter, modules |
| `nix/modules/` | HM + darwin modules sharing `shared.nix` option library |
| `.github/workflows/` | CI (flake check + go build/vet/lint/test) |

## Development Commands

```bash
# Build
nix build .#

# Dev shell
nix develop .#

# Run checks (inside dev shell)
go build ./...
go vet ./...
golangci-lint run ./...

# Run tests
go test -race -count=1 -shuffle=on ./...

# Full Nix check
nix flake check --allow-import-from-derivation

# Format
gofumpt -w .
```

## Code Conventions & Common Patterns

### Go

- Load the Go/Golang skill if available.
- **stdlib only** — `go.mod` has no `require` block. Import only from stdlib + internal package.
- **Error handling**: `fmt.Errorf` with context for validation/precondition errors; `fmt.Fprintf(os.Stderr, ...)` for warnings. Dock errors collected as `[]string` in `DockSyncResult.Errors` (non-fatal).
- **No build tags** — single codepath for all platforms (effectively darwin-only via flake).
- **Version injection**: `-X main.version=<ver>` via `ldflags` in Nix derivation.
- **Flag parsing workaround**: `main.go` has `syncFlagArgs()` to reorder `--no-dock` before positional args because Go's `flag` package stops parsing at the first non-flag argument.
- **`strings.SplitSeq`** used where applicable (Go 1.26 modernization).

## Important Files

| File | Role |
|---|---|
| `cmd/nix-spotlight/main.go` | CLI entrypoint |
| `internal/nixspotlight/trampoline.go` | Core sync logic |
| `internal/nixspotlight/dock.go` | Dock integration via `dockutil` CLI |
| `internal/nixspotlight/app.go` | App bundle type |
| `flake.nix` | Flake entrypoint |
| `nix/go-package.nix` | buildGoModule derivation |
| `nix/modules/shared.nix` | Shared NixOS module option library |
| `.golangci.yaml` | Linter configuration |
| `.github/workflows/ci.yaml` | CI pipeline |

## Runtime/Tooling Preferences

- **Go**: 1.26+ via nixpkgs. `GOTOOLCHAIN=local` to prevent auto-download.
- **Formatter**: `gofumpt` (via `nixfmt-tree` wrapper).
- **Linter**: `golangci-lint` v2.

## Testing & QA

- **Framework**: Standard `testing` package. No external test deps.
- **Pattern**: Table-driven tests throughout. `t.TempDir()` for isolation.
- **External dep mocking**: Fake shell scripts written to temp dirs, controlled via `t.Setenv("PATH", ...)` and environment variables. No mocks — real filesystem + fake executables.
- **dockutil mocking**: Test helper `writeFakeDockutil()` writes a shell script that logs all invocations to a calls file; `installDockutil()` prepends it to PATH.
- **CLI testing**: `captureRun()` redirects `os.Stdout`/`os.Stderr` and captures `os.Exit` via a test-only `run()` function return code.
- **Coverage focus**: Conditional branches, edge values, error paths. OS-level failure branches (disk full, permission denied) are intentionally not covered (no DI/fakes).
- **Race detection**: All tests run with `-race`.
- **CI**: Tests run on macOS only (`if: runner.os == 'macOS'`). Build + vet + lint on both ubuntu and macOS.
