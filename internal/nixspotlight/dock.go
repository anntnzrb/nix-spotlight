package nixspotlight

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// DockSyncResult is the result of a dock sync operation.
type DockSyncResult struct {
	Updated int
	Skipped int
	Errors  []string
}

// SyncDock updates dock items pointing to /nix/store paths to use trampoline locations.
func SyncDock(apps []string, dockutilPath string) DockSyncResult {
	dockutil := dockutilPath
	if dockutil == "" {
		var err error
		dockutil, err = exec.LookPath("dockutil")
		if err != nil || dockutil == "" {
			return DockSyncResult{}
		}
	}

	stdout, stderr, ok := runDockutil(dockutil, "-L")
	if !ok {
		return DockSyncResult{Errors: []string{"dockutil -L failed: " + stderr}}
	}

	appStems := make(map[string]string, len(apps))
	for _, app := range apps {
		appStems[strings.TrimSuffix(filepath.Base(app), ".app")] = app
	}

	result := DockSyncResult{}
	for _, line := range strings.Split(stdout, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.Contains(line, "/nix/store") {
			continue
		}

		parts := strings.SplitN(line, "\t", 2)
		name := strings.TrimSpace(parts[0])
		path := ""
		if len(parts) > 1 {
			path = strings.TrimSpace(parts[1])
		}

		trampoline := appStems[name]
		if trampoline == "" && path != "" {
			trampoline = appStems[strings.TrimSuffix(filepath.Base(path), filepath.Ext(filepath.Base(path)))]
		}
		if trampoline == "" {
			result.Skipped++
			continue
		}

		resolved := trampoline
		if resolvedPath, err := filepath.EvalSymlinks(trampoline); err == nil {
			resolved = resolvedPath
		}

		_, addStderr, addOK := runDockutil(dockutil, "--add", resolved, "--replacing", name)
		if !addOK {
			result.Errors = append(result.Errors, "Failed to update "+name+": "+addStderr)
			continue
		}
		result.Updated++
	}

	return result
}

func runDockutil(dockutil string, args ...string) (stdout string, stderr string, ok bool) {
	output, err := exec.Command(dockutil, args...).Output() //nolint:noctx // short-lived CLI, no context needed
	if err != nil {
		if exitErr, isExitErr := err.(*exec.ExitError); isExitErr {
			return string(output), string(exitErr.Stderr), false
		}
		return string(output), "", false
	}
	return string(output), "", true
}
