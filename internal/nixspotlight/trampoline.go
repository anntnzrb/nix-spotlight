package nixspotlight

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GatherApps discovers valid .app bundles in fromDir.
// It collects direct *.app entries first, then one-level nested */*.app entries
// (e.g. KDE-style layout). Invalid apps (no Info.plist) are skipped.
func GatherApps(fromDir string) ([]App, error) {
	entries, err := os.ReadDir(fromDir)
	if err != nil {
		return nil, err
	}

	apps := make([]App, 0)

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".app") {
			app := NewApp(filepath.Join(fromDir, name))
			if app.IsValid() {
				apps = append(apps, app)
			}
		}
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() && !strings.HasSuffix(name, ".app") {
			nestedEntries, err := os.ReadDir(filepath.Join(fromDir, name))
			if err != nil {
				// Intentionally skip unreadable nested directories; callers only need
				// the top-level ReadDir failure to decide whether sync is safe.
				continue
			}
			for _, nested := range nestedEntries {
				if strings.HasSuffix(nested.Name(), ".app") {
					app := NewApp(filepath.Join(fromDir, name, nested.Name()))
					if app.IsValid() {
						apps = append(apps, app)
					}
				}
			}
		}
	}

	return apps, nil
}

// CreateTrampoline creates a trampoline app in targetDir for source.
// If the trampoline path already exists as a symlink, it is removed first
// to avoid operating through it. The trampoline is a directory containing a
// single Contents symlink pointing to source.Contents().
// Returns the trampoline path on success.
func CreateTrampoline(source App, targetDir string) (string, error) {
	trampoline := filepath.Join(targetDir, source.Name())

	info, err := os.Lstat(trampoline)
	if err == nil && info.Mode()&os.ModeSymlink != 0 {
		if err := os.Remove(trampoline); err != nil {
			return "", err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if err := os.MkdirAll(trampoline, 0o755); err != nil {
		return "", err
	}

	entries, err := os.ReadDir(trampoline)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(trampoline, entry.Name())); err != nil {
			return "", err
		}
	}

	contentsLink := filepath.Join(trampoline, "Contents")

	if err := os.Symlink(source.Contents(), contentsLink); err != nil {
		return "", err
	}

	return trampoline, nil
}

// SyncTrampolines syncs all .app bundles from source to a trampolines directory.
func SyncTrampolines(fromDir, toDir string) ([]string, error) {
	fromInfo, err := os.Stat(fromDir)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory does not exist: %s", fromDir)
	}
	if err != nil {
		return nil, err
	}
	if !fromInfo.IsDir() {
		return nil, fmt.Errorf("source path is not a directory: %s", fromDir)
	}

	toInfo, err := os.Lstat(toDir)
	if err == nil {
		if toInfo.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("target path must not be a symlink: %s", toDir)
		}
		if !toInfo.IsDir() {
			return nil, fmt.Errorf("target path is not a directory: %s", toDir)
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	resolvedFromDir, err := resolvedPath(fromDir)
	if err != nil {
		return nil, err
	}
	resolvedToDir, err := resolvedPath(toDir)
	if err != nil {
		return nil, err
	}
	fromInfo, err = os.Stat(resolvedFromDir)
	if err != nil {
		return nil, err
	}
	toExists := false
	toInfo, err = os.Stat(resolvedToDir)
	if err == nil {
		toExists = true
		if os.SameFile(fromInfo, toInfo) {
			return nil, fmt.Errorf("source and target directories must differ: %s", fromDir)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	if toExists {
		rel, err := filepath.Rel(resolvedToDir, resolvedFromDir)
		if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("target path must not contain source: %s", toDir)
		}
	}

	apps, err := GatherApps(fromDir)
	if err != nil {
		return nil, err
	}

	if err := os.RemoveAll(toDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(toDir, 0o755); err != nil {
		return nil, err
	}

	trampolines := make([]string, 0, len(apps))
	for _, app := range apps {
		trampoline, err := CreateTrampoline(app, toDir)
		if err != nil {
			return nil, err
		}
		trampolines = append(trampolines, trampoline)
	}

	now := time.Now()
	for _, trampoline := range trampolines {
		if err := os.Chtimes(trampoline, now, now); err != nil {
			return nil, err
		}
	}

	return trampolines, nil
}

// resolvedPath returns the canonical absolute path for path.
// It resolves symlinks and handles paths where intermediate components
// do not yet exist by walking up from the leaf.
func resolvedPath(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		return filepath.Abs(resolved)
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	missing := make([]string, 0)
	current := filepath.Clean(path)
	for {
		resolved, err = filepath.EvalSymlinks(current)
		if err == nil {
			abs, err := filepath.Abs(resolved)
			if err != nil {
				return "", err
			}
			for i := len(missing) - 1; i >= 0; i-- {
				abs = filepath.Join(abs, missing[i])
			}
			return abs, nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}

		parent := filepath.Dir(current)
		base := filepath.Base(current)
		if parent == current {
			return filepath.Abs(path)
		}
		missing = append(missing, base)
		current = parent
	}
}
