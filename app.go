package nixspotlight

import (
	"os"
	"path/filepath"
)

// App represents a macOS .app bundle.
type App struct {
	path string
}

// NewApp returns an App for path.
func NewApp(path string) App {
	return App{path: path}
}

// Path returns the app bundle path.
func (a App) Path() string {
	return a.path
}

// Name returns the app bundle name.
func (a App) Name() string {
	return filepath.Base(a.path)
}

// Contents returns the app Contents directory path.
func (a App) Contents() string {
	return filepath.Join(a.path, "Contents")
}

// InfoPlist returns the app Info.plist path.
func (a App) InfoPlist() string {
	return filepath.Join(a.Contents(), "Info.plist")
}

// IsValid reports whether the app has an Info.plist.
func (a App) IsValid() bool {
	_, err := os.Stat(a.InfoPlist())
	return err == nil
}
