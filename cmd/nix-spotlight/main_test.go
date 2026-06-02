package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name       string
		args       func(t *testing.T) []string
		wantCode   int
		wantStdout []string
		wantStderr []string
		verify     func(t *testing.T, args []string)
	}{
		{
			name:       "no args",
			args:       func(_ *testing.T) []string { return nil },
			wantCode:   2,
			wantStderr: []string{"usage: nix-spotlight sync <from> <to> [--no-dock]"},
		},
		{
			name:       "help",
			args:       func(_ *testing.T) []string { return []string{"--help"} },
			wantCode:   0,
			wantStdout: []string{"usage: nix-spotlight sync <from> <to> [--no-dock]"},
		},
		{
			name:       "version",
			args:       func(_ *testing.T) []string { return []string{"--version"} },
			wantCode:   0,
			wantStdout: []string{"nix-spotlight"},
		},
		{
			name:       "sync help",
			args:       func(_ *testing.T) []string { return []string{"sync", "-h"} },
			wantCode:   0,
			wantStderr: []string{"usage: nix-spotlight sync <from> <to> [--no-dock]"},
		},
		{
			name:       "sync long help",
			args:       func(_ *testing.T) []string { return []string{"sync", "--help"} },
			wantCode:   0,
			wantStderr: []string{"usage: nix-spotlight sync <from> <to> [--no-dock]"},
		},
		{
			name: "sync version is not global",
			args: func(t *testing.T) []string {
				dir := t.TempDir()
				return []string{"sync", "--version", filepath.Join(dir, "source"), filepath.Join(dir, "target")}
			},
			wantCode:   2,
			wantStderr: []string{"flag provided but not defined: -version"},
		},
		{
			name: "sync missing source",
			args: func(t *testing.T) []string {
				dir := t.TempDir()
				return []string{"sync", filepath.Join(dir, "missing"), filepath.Join(dir, "target")}
			},
			wantCode:   1,
			wantStderr: []string{"error:", "does not exist"},
		},
		{
			name: "sync success",
			args: func(t *testing.T) []string {
				dir := t.TempDir()
				source := filepath.Join(dir, "source")
				target := filepath.Join(dir, "target")
				mkdir(t, source)
				makeApp(t, source, "Test.app")
				return []string{"sync", "--no-dock", source, target}
			},
			wantCode:   0,
			wantStdout: []string{"Synced 1 apps to"},
			verify: func(t *testing.T, args []string) {
				assertTrampoline(t, args[3], "Test.app")
			},
		},
		{
			name: "sync empty source",
			args: func(t *testing.T) []string {
				dir := t.TempDir()
				source := filepath.Join(dir, "source")
				target := filepath.Join(dir, "target")
				mkdir(t, source)
				return []string{"sync", "--no-dock", source, target}
			},
			wantCode:   0,
			wantStdout: []string{"Synced 0 apps"},
		},
		{
			name: "trailing no dock flag",
			args: func(t *testing.T) []string {
				dir := t.TempDir()
				source := filepath.Join(dir, "source")
				target := filepath.Join(dir, "target")
				mkdir(t, source)
				makeApp(t, source, "Tail.app")
				return []string{"sync", source, target, "--no-dock"}
			},
			wantCode:   0,
			wantStdout: []string{"Synced 1 apps"},
			verify: func(t *testing.T, args []string) {
				assertTrampoline(t, args[2], "Tail.app")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.args(t)
			code, stdout, stderr := captureRun(t, args)
			if code != tt.wantCode {
				t.Errorf("run() code = %d; want %d", code, tt.wantCode)
			}
			for _, want := range tt.wantStdout {
				if !strings.Contains(stdout, want) {
					t.Errorf("stdout = %q; want to contain %q", stdout, want)
				}
			}
			for _, want := range tt.wantStderr {
				if !strings.Contains(stderr, want) {
					t.Errorf("stderr = %q; want to contain %q", stderr, want)
				}
			}
			if tt.verify != nil {
				tt.verify(t, args)
			}
		})
	}
}

func TestRunNoDockSkipsDockutil(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	target := filepath.Join(dir, "target")
	marker := filepath.Join(dir, "dockutil-called")
	mkdir(t, source)
	makeApp(t, source, "NoDock.app")
	installDockutil(t, dir, "#!/bin/sh\n: > \"$DOCKUTIL_MARKER\"\nexit 0\n")
	t.Setenv("DOCKUTIL_MARKER", marker)

	code, stdout, stderr := captureRun(t, []string{"sync", source, target, "--no-dock"})
	if code != 0 {
		t.Fatalf("run() code = %d; want 0; stdout = %q stderr = %q", code, stdout, stderr)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("dockutil was called with --no-dock set")
	} else if !os.IsNotExist(err) {
		t.Fatalf("os.Stat(marker): %v", err)
	}
}

func TestRunPrintsDockWarnings(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	target := filepath.Join(dir, "target")
	mkdir(t, source)
	makeApp(t, source, "Test.app")
	installDockutil(t, dir, "#!/bin/sh\nif [ \"$1\" = \"-L\" ]; then\n\tprintf 'Test\\t/nix/store/old/Test.app\\n'\n\texit 0\nfi\nif [ \"$1\" = \"--add\" ]; then\n\tprintf 'replace failed' >&2\n\texit 1\nfi\nexit 0\n")

	code, stdout, stderr := captureRun(t, []string{"sync", source, target})
	if code != 0 {
		t.Fatalf("run() code = %d; want 0; stdout = %q stderr = %q", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "warning: Failed to update Test: replace failed") {
		t.Errorf("stderr = %q; want dock warning", stderr)
	}
}

func TestRunSyncWithDockNoWarnings(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	target := filepath.Join(dir, "target")
	mkdir(t, source)
	makeApp(t, source, "Clean.app")
	installDockutil(t, dir, "#!/bin/sh\nif [ \"$1\" = \"-L\" ]; then\n\texit 0\nfi\nexit 0\n")

	code, stdout, stderr := captureRun(t, []string{"sync", source, target})
	if code != 0 {
		t.Fatalf("run() code = %d; want 0; stdout = %q stderr = %q", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "Synced 1 apps to") {
		t.Errorf("stdout = %q; want sync success", stdout)
	}
	if strings.Contains(stderr, "warning:") {
		t.Errorf("stderr = %q; want no warnings", stderr)
	}
}

func TestRunPrintsMultipleDockWarnings(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	target := filepath.Join(dir, "target")
	mkdir(t, source)
	makeApp(t, source, "One.app")
	makeApp(t, source, "Two.app")
	installDockutil(t, dir, "#!/bin/sh\nif [ \"$1\" = \"-L\" ]; then\n\tprintf 'One\\t/nix/store/old/One.app\\nTwo\\t/nix/store/old/Two.app\\n'\n\texit 0\nfi\nif [ \"$1\" = \"--add\" ]; then\n\tcase \"$4\" in\n\t\tOne) printf 'error1' >&2 ;;\n\t\tTwo) printf 'error2' >&2 ;;\n\tesac\n\texit 1\nfi\nexit 0\n")

	code, stdout, stderr := captureRun(t, []string{"sync", source, target})
	if code != 0 {
		t.Fatalf("run() code = %d; want 0; stdout = %q stderr = %q", code, stdout, stderr)
	}
	if !strings.Contains(stderr, "warning: Failed to update One: error1") {
		t.Errorf("stderr = %q; want first dock warning", stderr)
	}
	if !strings.Contains(stderr, "warning: Failed to update Two: error2") {
		t.Errorf("stderr = %q; want second dock warning", stderr)
	}
}

// captureRun runs the CLI with args and returns the exit code, captured stdout, and captured stderr.
func captureRun(t *testing.T, args []string) (int, string, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	oldStdout, oldStderr := os.Stdout, os.Stderr
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(stdout): %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(stderr): %v", err)
	}

	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter

	stdoutDone := make(chan error, 1)
	stderrDone := make(chan error, 1)
	go func() {
		_, err := io.Copy(&stdout, stdoutReader)
		stdoutDone <- err
	}()
	go func() {
		_, err := io.Copy(&stderr, stderrReader)
		stderrDone <- err
	}()

	code := run(args)

	if err := stdoutWriter.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	if err := stderrWriter.Close(); err != nil {
		t.Fatalf("close stderr writer: %v", err)
	}
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	if err := <-stdoutDone; err != nil {
		t.Fatalf("copy stdout: %v", err)
	}
	if err := <-stderrDone; err != nil {
		t.Fatalf("copy stderr: %v", err)
	}
	if err := stdoutReader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	if err := stderrReader.Close(); err != nil {
		t.Fatalf("close stderr reader: %v", err)
	}

	return code, stdout.String(), stderr.String()
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q): %v", path, err)
	}
}

// makeApp creates a minimal valid .app bundle in source with an Info.plist.
//
//nolint:unparam
func makeApp(t *testing.T, source, name string) string {
	t.Helper()
	app := filepath.Join(source, name)
	contents := filepath.Join(app, "Contents")
	mkdir(t, contents)
	if err := os.WriteFile(filepath.Join(contents, "Info.plist"), nil, 0o644); err != nil {
		t.Fatalf("os.WriteFile(Info.plist): %v", err)
	}
	return app
}

// assertTrampoline checks that target/name/Contents is a symlink.
func assertTrampoline(t *testing.T, target, name string) {
	t.Helper()

	info, err := os.Lstat(filepath.Join(target, name, "Contents"))
	if err != nil {
		t.Fatalf("os.Lstat(trampoline Contents): %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s Contents mode = %v; want symlink", name, info.Mode())
	}
}

// installDockutil writes a fake dockutil script to dir/bin/dockutil and prepends it to PATH.
func installDockutil(t *testing.T, dir, script string) {
	t.Helper()

	bin := filepath.Join(dir, "bin")
	mkdir(t, bin)
	path := filepath.Join(bin, "dockutil")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("os.WriteFile(dockutil): %v", err)
	}
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+oldPath)
}
