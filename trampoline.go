package nixspotlight

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var appPatterns = [...]string{"*.app", "*/*.app"}

// GatherApps gathers all valid .app bundles from a directory.
func GatherApps(fromDir string) ([]App, error) {
	apps := make([]App, 0)

	for _, pattern := range appPatterns {
		matches, err := filepath.Glob(filepath.Join(fromDir, pattern))
		if err != nil {
			return apps, err
		}
		for _, match := range matches {
			app := NewApp(match)
			if app.IsValid() {
				apps = append(apps, app)
			}
		}
	}

	return apps, nil
}

// CreateTrampoline creates a symlink-based trampoline for a .app bundle.
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

	contentsLink := filepath.Join(trampoline, "Contents")
	info, err = os.Lstat(contentsLink)
	if err == nil {
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			err = os.Remove(contentsLink)
		case info.IsDir():
			err = os.RemoveAll(contentsLink)
		default:
			err = os.Remove(contentsLink)
		}
		if err != nil {
			return "", err
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}

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

	toInfo, err := os.Stat(toDir)
	if err == nil && !toInfo.IsDir() {
		return nil, fmt.Errorf("target path is not a directory: %s", toDir)
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	resolvedFrom, err := resolvedPath(fromDir)
	if err != nil {
		return nil, err
	}
	resolvedTo, err := resolvedPath(toDir)
	if err != nil {
		return nil, err
	}
	if resolvedFrom == resolvedTo {
		return nil, fmt.Errorf("source and target directories must differ: %s", fromDir)
	}

	if err := os.RemoveAll(toDir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(toDir, 0o755); err != nil {
		return nil, err
	}

	apps, err := GatherApps(fromDir)
	if err != nil {
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
