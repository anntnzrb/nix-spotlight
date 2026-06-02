package nixspotlight

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncDock(t *testing.T) {
	tests := []struct {
		name            string
		apps            []string
		fakeDockutil    bool
		explicitPath    bool
		listStdout      string
		listStderr      string
		listExit        string
		addStderr       string
		addExit         string
		wantUpdated     int
		wantSkipped     int
		wantErrors      []string
		wantErrContains []string
		checkCalls      func(t *testing.T, calls [][]string, appPaths []string)
	}{
		{
			name:        "no_dockutil",
			apps:        []string{"App1.app", "App2.app"},
			wantUpdated: 0,
			wantSkipped: 0,
		},
		{
			name:         "with_explicit_path",
			apps:         []string{"App1.app"},
			fakeDockutil: true,
			explicitPath: true,
			wantUpdated:  0,
			wantSkipped:  0,
			checkCalls: func(t *testing.T, calls [][]string, _ []string) {
				t.Helper()
				if len(calls) != 1 {
					t.Fatalf("calls = %d; want 1", len(calls))
				}
				wantArgs(t, calls[0], []string{"-L"})
			},
		},
		{
			name:            "dockutil_fails",
			apps:            []string{"App1.app"},
			fakeDockutil:    true,
			listStderr:      "dockutil error",
			listExit:        "1",
			wantUpdated:     0,
			wantSkipped:     0,
			wantErrors:      []string{"dockutil -L failed: dockutil error"},
			wantErrContains: []string{"dockutil -L failed"},
		},
		{
			name:         "no_nix_items",
			apps:         []string{"App1.app"},
			fakeDockutil: true,
			listStdout:   "Safari\t/Applications/Safari.app\nChrome\t/Applications/Chrome.app",
			wantUpdated:  0,
			wantSkipped:  0,
		},
		{
			name:         "updates_matching_items",
			apps:         []string{"MyApp.app"},
			fakeDockutil: true,
			listStdout:   "MyApp\t/nix/store/abc123-myapp/Applications/MyApp.app",
			wantUpdated:  1,
			wantSkipped:  0,
			checkCalls: func(t *testing.T, calls [][]string, appPaths []string) {
				t.Helper()
				if len(calls) != 2 {
					t.Fatalf("calls = %d; want 2", len(calls))
				}
				resolved := mustEvalSymlinks(t, appPaths[0])
				wantArgs(t, calls[1], []string{"--add", resolved, "--replacing", "MyApp"})
			},
		},
		{
			name:         "updates_matching_items_by_path",
			apps:         []string{"MyApp.app"},
			fakeDockutil: true,
			listStdout:   "Different Name\t/nix/store/abc123/Applications/MyApp.app",
			wantUpdated:  1,
			wantSkipped:  0,
			checkCalls: func(t *testing.T, calls [][]string, appPaths []string) {
				t.Helper()
				if len(calls) != 2 {
					t.Fatalf("calls = %d; want 2", len(calls))
				}
				resolved := mustEvalSymlinks(t, appPaths[0])
				wantArgs(t, calls[1], []string{"--add", resolved, "--replacing", "Different Name"})
			},
		},
		{
			name:         "skips_unmatched_nix_items",
			apps:         []string{"OtherApp.app"},
			fakeDockutil: true,
			listStdout:   "MyApp\t/nix/store/abc123-myapp/Applications/MyApp.app",
			wantUpdated:  0,
			wantSkipped:  1,
		},
		{
			name:            "reports_add_errors",
			apps:            []string{"MyApp.app"},
			fakeDockutil:    true,
			listStdout:      "MyApp\t/nix/store/abc123-myapp/Applications/MyApp.app",
			addStderr:       "permission denied",
			addExit:         "1",
			wantUpdated:     0,
			wantSkipped:     0,
			wantErrors:      []string{"Failed to update MyApp: permission denied"},
			wantErrContains: []string{"Failed to update MyApp"},
		},
		{
			name:         "empty_line",
			apps:         []string{"App1.app"},
			fakeDockutil: true,
			listStdout:   "\n\n/nix/store/test\n",
			wantUpdated:  0,
			wantSkipped:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := t.TempDir()
			appPaths := make([]string, 0, len(tt.apps))
			for _, app := range tt.apps {
				appPath := filepath.Join(baseDir, app)
				if err := os.Mkdir(appPath, 0o755); err != nil {
					t.Fatalf("os.Mkdir(%q): %v", appPath, err)
				}
				appPaths = append(appPaths, appPath)
			}

			dockutilPath := ""
			callsPath := ""
			if tt.fakeDockutil {
				fakePath, callLog := writeFakeDockutil(t)
				callsPath = callLog
				if tt.explicitPath {
					dockutilPath = fakePath
					t.Setenv("PATH", t.TempDir())
				} else {
					t.Setenv("PATH", filepath.Dir(fakePath)+string(os.PathListSeparator)+os.Getenv("PATH"))
				}
				t.Setenv("DOCKUTIL_LIST_STDOUT", tt.listStdout)
				t.Setenv("DOCKUTIL_LIST_STDERR", tt.listStderr)
				t.Setenv("DOCKUTIL_LIST_EXIT", tt.listExit)
				t.Setenv("DOCKUTIL_ADD_STDERR", tt.addStderr)
				t.Setenv("DOCKUTIL_ADD_EXIT", tt.addExit)
			} else {
				t.Setenv("PATH", t.TempDir())
			}

			got := SyncDock(appPaths, dockutilPath)
			if got.Updated != tt.wantUpdated {
				t.Errorf("Updated = %d; want %d", got.Updated, tt.wantUpdated)
			}
			if got.Skipped != tt.wantSkipped {
				t.Errorf("Skipped = %d; want %d", got.Skipped, tt.wantSkipped)
			}
			wantStrings(t, got.Errors, tt.wantErrors)
			for _, want := range tt.wantErrContains {
				if len(got.Errors) == 0 || !strings.Contains(got.Errors[0], want) {
					t.Errorf("Errors = %q; want first error containing %q", got.Errors, want)
				}
			}
			if tt.checkCalls != nil {
				tt.checkCalls(t, readCalls(t, callsPath), appPaths)
			}
		})
	}
}

func writeFakeDockutil(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "dockutil")
	callsPath := filepath.Join(dir, "calls")
	script := `#!/bin/sh
{
	printf 'CALL\n'
	for arg in "$@"; do
		printf '%s\n' "$arg"
	done
} >> "$DOCKUTIL_CALLS"

if [ "$1" = "-L" ]; then
	printf '%s' "$DOCKUTIL_LIST_STDOUT"
	printf '%s' "$DOCKUTIL_LIST_STDERR" >&2
	exit "${DOCKUTIL_LIST_EXIT:-0}"
fi

if [ "$1" = "--add" ]; then
	printf '%s' "$DOCKUTIL_ADD_STDERR" >&2
	exit "${DOCKUTIL_ADD_EXIT:-0}"
fi

exit 0
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("os.WriteFile(%q): %v", path, err)
	}
	if err := os.Chmod(path, 0o755); err != nil {
		t.Fatalf("os.Chmod(%q): %v", path, err)
	}
	t.Setenv("DOCKUTIL_CALLS", callsPath)
	return path, callsPath
}

func readCalls(t *testing.T, path string) [][]string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", path, err)
	}
	blocks := strings.Split(string(data), "CALL\n")
	calls := make([][]string, 0, len(blocks)-1)
	for _, block := range blocks {
		block = strings.TrimSuffix(block, "\n")
		if block == "" {
			continue
		}
		calls = append(calls, strings.Split(block, "\n"))
	}
	return calls
}

func wantArgs(t *testing.T, got []string, want []string) {
	t.Helper()
	wantStrings(t, got, want)
}

func wantStrings(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("len = %d; want %d; got %q want %q", len(got), len(want), got, want)
		return
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q; want %q; got %q want %q", i, got[i], want[i], got, want)
			return
		}
	}
}

func mustEvalSymlinks(t *testing.T, path string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks(%q): %v", path, err)
	}
	return resolved
}
