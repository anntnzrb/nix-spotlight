package nixspotlight

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppProperties(t *testing.T) {
	tests := []struct {
		name string
		app  string
	}{
		{
			name: "valid app",
			app:  "Test.app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			appPath := filepath.Join(tmpDir, tt.app)
			contentsPath := filepath.Join(appPath, "Contents")
			infoPlistPath := filepath.Join(contentsPath, "Info.plist")

			if err := os.MkdirAll(contentsPath, 0o755); err != nil {
				t.Fatal(err)
			}
			file, err := os.Create(infoPlistPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}

			app := NewApp(appPath)

			if got := app.Name(); got != tt.app {
				t.Fatalf("Name() = %q, want %q", got, tt.app)
			}
			if got := app.Contents(); got != contentsPath {
				t.Fatalf("Contents() = %q, want %q", got, contentsPath)
			}
			if got := app.InfoPlist(); got != infoPlistPath {
				t.Fatalf("InfoPlist() = %q, want %q", got, infoPlistPath)
			}
			if !app.IsValid() {
				t.Fatal("IsValid() = false, want true")
			}
		})
	}
}

func TestAppInvalid(t *testing.T) {
	tests := []struct {
		name         string
		makeContents bool
	}{
		{
			name: "no Contents directory",
		},
		{
			name:         "no Info.plist",
			makeContents: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			appPath := filepath.Join(tmpDir, "Invalid.app")
			contentsPath := filepath.Join(appPath, "Contents")

			if err := os.Mkdir(appPath, 0o755); err != nil {
				t.Fatal(err)
			}
			if tt.makeContents {
				if err := os.Mkdir(contentsPath, 0o755); err != nil {
					t.Fatal(err)
				}
			}

			app := NewApp(appPath)

			if app.IsValid() {
				t.Fatal("IsValid() = true, want false")
			}
		})
	}
}

func TestAppIsValidContentsNotDir(t *testing.T) {
	tmpDir := t.TempDir()
	appPath := filepath.Join(tmpDir, "Invalid.app")
	if err := os.Mkdir(appPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(appPath, "Contents"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	app := NewApp(appPath)
	if app.IsValid() {
		t.Fatal("IsValid() = true, want false")
	}
}
