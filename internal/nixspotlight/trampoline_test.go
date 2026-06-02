package nixspotlight

import (
	"os"
	"path/filepath"
	"testing"
)

func makeTrampolineTestApp(t *testing.T, path string) App {
	t.Helper()

	contents := filepath.Join(path, "Contents")
	if err := os.MkdirAll(contents, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", contents, err)
	}
	infoPlist := filepath.Join(contents, "Info.plist")
	if err := os.WriteFile(infoPlist, []byte{}, 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", infoPlist, err)
	}

	return NewApp(path)
}

func makeTrampolineTestDir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", path, err)
	}
}

func makeTrampolineTestFile(t *testing.T, path string) {
	t.Helper()

	makeTrampolineTestDir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}

func assertTrampolineContentsLink(t *testing.T, trampoline string, source App) {
	t.Helper()

	contents := filepath.Join(trampoline, "Contents")
	info, err := os.Lstat(contents)
	if err != nil {
		t.Fatalf("Lstat(%q): %v", contents, err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%q is not a symlink", contents)
	}
	linkTarget, err := os.Readlink(contents)
	if err != nil {
		t.Fatalf("Readlink(%q): %v", contents, err)
	}
	if linkTarget != source.Contents() {
		t.Fatalf("Readlink(%q) = %q, want %q", contents, linkTarget, source.Contents())
	}
}

func assertPathExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Stat(%q): %v", path, err)
	}
}

func assertPathDoesNotExist(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%q exists", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("Stat(%q): %v", path, err)
	}
}

func TestCreateTrampoline(t *testing.T) {
	tmp := t.TempDir()
	sourceDir := filepath.Join(tmp, "source")
	makeTrampolineTestDir(t, sourceDir)
	source := makeTrampolineTestApp(t, filepath.Join(sourceDir, "MyApp.app"))
	targetDir := filepath.Join(tmp, "target")
	makeTrampolineTestDir(t, targetDir)

	trampoline, err := CreateTrampoline(source, targetDir)
	if err != nil {
		t.Fatalf("CreateTrampoline() error = %v", err)
	}

	want := filepath.Join(targetDir, "MyApp.app")
	if trampoline != want {
		t.Fatalf("CreateTrampoline() = %q, want %q", trampoline, want)
	}
	assertPathExists(t, trampoline)
	assertTrampolineContentsLink(t, trampoline, source)
}

func TestCreateTrampolineReplacesExisting(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, contents string)
	}{
		{
			name: "symlink",
			setup: func(t *testing.T, contents string) {
				t.Helper()
				makeTrampolineTestDir(t, filepath.Dir(contents))
				if err := os.Symlink("/nonexistent", contents); err != nil {
					t.Fatalf("Symlink(%q): %v", contents, err)
				}
			},
		},
		{
			name: "directory",
			setup: func(t *testing.T, contents string) {
				t.Helper()
				makeTrampolineTestDir(t, contents)
				makeTrampolineTestFile(t, filepath.Join(contents, "old"))
			},
		},
		{
			name: "file",
			setup: func(t *testing.T, contents string) {
				t.Helper()
				makeTrampolineTestFile(t, contents)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			sourceDir := filepath.Join(tmp, "source")
			makeTrampolineTestDir(t, sourceDir)
			source := makeTrampolineTestApp(t, filepath.Join(sourceDir, "MyApp.app"))
			targetDir := filepath.Join(tmp, "target")
			trampolineDir := filepath.Join(targetDir, "MyApp.app")
			contents := filepath.Join(trampolineDir, "Contents")
			tc.setup(t, contents)

			trampoline, err := CreateTrampoline(source, targetDir)
			if err != nil {
				t.Fatalf("CreateTrampoline() error = %v", err)
			}

			assertTrampolineContentsLink(t, trampoline, source)
		})
	}
}

func TestCreateTrampolineCreatesParentDirs(t *testing.T) {
	tmp := t.TempDir()
	sourceDir := filepath.Join(tmp, "source")
	makeTrampolineTestDir(t, sourceDir)
	source := makeTrampolineTestApp(t, filepath.Join(sourceDir, "MyApp.app"))
	targetDir := filepath.Join(tmp, "nested", "target")

	trampoline, err := CreateTrampoline(source, targetDir)
	if err != nil {
		t.Fatalf("CreateTrampoline() error = %v", err)
	}

	assertPathExists(t, trampoline)
	assertTrampolineContentsLink(t, trampoline, source)
}

func TestCreateTrampolineSymlinkedAppDir(t *testing.T) {
	tmp := t.TempDir()
	sourceDir := filepath.Join(tmp, "source")
	makeTrampolineTestDir(t, sourceDir)
	source := makeTrampolineTestApp(t, filepath.Join(sourceDir, "MyApp.app"))

	targetDir := filepath.Join(tmp, "target")
	makeTrampolineTestDir(t, targetDir)
	linkedTarget := filepath.Join(tmp, "linked-target.app")
	protectedContents := filepath.Join(linkedTarget, "Contents")
	makeTrampolineTestDir(t, protectedContents)
	protectedFile := filepath.Join(protectedContents, "protected")
	makeTrampolineTestFile(t, protectedFile)

	trampolinePath := filepath.Join(targetDir, "MyApp.app")
	if err := os.Symlink(linkedTarget, trampolinePath); err != nil {
		t.Fatalf("Symlink(%q, %q): %v", linkedTarget, trampolinePath, err)
	}

	trampoline, err := CreateTrampoline(source, targetDir)
	if err != nil {
		t.Fatalf("CreateTrampoline() error = %v", err)
	}

	info, err := os.Lstat(trampoline)
	if err != nil {
		t.Fatalf("Lstat(%q): %v", trampoline, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("%q is still a symlink", trampoline)
	}
	assertPathExists(t, protectedFile)
	assertTrampolineContentsLink(t, trampoline, source)
}

func TestGatherApps(t *testing.T) {
	tmp := t.TempDir()
	validAppNames := []string{"App1.app", "App2.app"}
	makeTrampolineTestApp(t, filepath.Join(tmp, validAppNames[0]))
	nested := filepath.Join(tmp, "Nested")
	makeTrampolineTestDir(t, nested)
	makeTrampolineTestApp(t, filepath.Join(nested, validAppNames[1]))
	makeTrampolineTestDir(t, filepath.Join(tmp, "Invalid.app"))

	apps := GatherApps(tmp)

	if len(apps) != len(validAppNames) {
		t.Fatalf("len(GatherApps()) = %d, want %d", len(apps), len(validAppNames))
	}
	for i, name := range validAppNames {
		if apps[i].Name() != name {
			t.Fatalf("GatherApps()[%d].Name() = %q, want %q", i, apps[i].Name(), name)
		}
	}
}

func TestGatherAppsEmptyDir(t *testing.T) {
	tmp := t.TempDir()

	apps := GatherApps(tmp)
	if apps == nil {
		t.Fatal("GatherApps() returned nil, want empty slice")
	}
	if len(apps) != 0 {
		t.Fatalf("len(GatherApps()) = %d, want 0", len(apps))
	}
}

func TestGatherAppsNoValidApps(t *testing.T) {
	tmp := t.TempDir()
	makeTrampolineTestDir(t, filepath.Join(tmp, "Invalid1.app"))
	makeTrampolineTestDir(t, filepath.Join(tmp, "Invalid2.app", "Contents"))

	apps := GatherApps(tmp)
	if apps == nil {
		t.Fatal("GatherApps() returned nil, want empty slice")
	}
	if len(apps) != 0 {
		t.Fatalf("len(GatherApps()) = %d, want 0", len(apps))
	}
}

func TestGatherAppsNestedInvalid(t *testing.T) {
	tmp := t.TempDir()
	nested := filepath.Join(tmp, "Nested")
	makeTrampolineTestDir(t, nested)
	makeTrampolineTestApp(t, filepath.Join(nested, "Valid.app"))
	makeTrampolineTestDir(t, filepath.Join(nested, "Invalid.app"))

	apps := GatherApps(tmp)
	if len(apps) != 1 {
		t.Fatalf("len(GatherApps()) = %d, want 1", len(apps))
	}
	if apps[0].Name() != "Valid.app" {
		t.Fatalf("GatherApps()[0].Name() = %q, want %q", apps[0].Name(), "Valid.app")
	}
}

func TestSyncTrampolines(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source")
	makeTrampolineTestDir(t, source)
	target := filepath.Join(tmp, "target")

	appNames := []string{"App1.app", "App2.app"}
	for _, name := range appNames {
		makeTrampolineTestApp(t, filepath.Join(source, name))
	}

	trampolines, err := SyncTrampolines(source, target)
	if err != nil {
		t.Fatalf("SyncTrampolines() error = %v", err)
	}

	if len(trampolines) != len(appNames) {
		t.Fatalf("len(SyncTrampolines()) = %d, want %d", len(trampolines), len(appNames))
	}
	assertPathExists(t, target)
	for _, name := range appNames {
		assertTrampolineContentsLink(t, filepath.Join(target, name), NewApp(filepath.Join(source, name)))
	}
}

func TestSyncTrampolinesCleansExisting(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source")
	makeTrampolineTestDir(t, source)
	target := filepath.Join(tmp, "target")
	makeTrampolineTestDir(t, target)
	oldApp := filepath.Join(target, "OldApp.app")
	makeTrampolineTestDir(t, oldApp)
	makeTrampolineTestFile(t, filepath.Join(oldApp, "garbage"))
	makeTrampolineTestApp(t, filepath.Join(source, "NewApp.app"))

	trampolines, err := SyncTrampolines(source, target)
	if err != nil {
		t.Fatalf("SyncTrampolines() error = %v", err)
	}

	if len(trampolines) != 1 {
		t.Fatalf("len(SyncTrampolines()) = %d, want 1", len(trampolines))
	}
	assertPathDoesNotExist(t, filepath.Join(target, "OldApp.app"))
	assertPathExists(t, filepath.Join(target, "NewApp.app"))
}

func TestSyncTrampolinesEmptySource(t *testing.T) {
	tmp := t.TempDir()
	source := filepath.Join(tmp, "source")
	makeTrampolineTestDir(t, source)
	target := filepath.Join(tmp, "target")

	trampolines, err := SyncTrampolines(source, target)
	if err != nil {
		t.Fatalf("SyncTrampolines() error = %v", err)
	}

	if trampolines == nil {
		t.Fatal("SyncTrampolines() returned nil, want empty slice")
	}
	if len(trampolines) != 0 {
		t.Fatalf("len(SyncTrampolines()) = %d, want 0", len(trampolines))
	}
	assertPathExists(t, target)
}

func TestSyncTrampolinesErrors(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(t *testing.T, tmp string) (fromDir string, toDir string)
		wantErr func(fromDir string, toDir string) string
	}{
		{
			name: "missing source",
			setup: func(t *testing.T, tmp string) (string, string) {
				t.Helper()
				return filepath.Join(tmp, "missing"), filepath.Join(tmp, "target")
			},
			wantErr: func(fromDir string, _ string) string {
				return "source directory does not exist: " + fromDir
			},
		},
		{
			name: "source not dir",
			setup: func(t *testing.T, tmp string) (string, string) {
				t.Helper()
				source := filepath.Join(tmp, "source")
				makeTrampolineTestFile(t, source)
				return source, filepath.Join(tmp, "target")
			},
			wantErr: func(fromDir string, _ string) string {
				return "source path is not a directory: " + fromDir
			},
		},
		{
			name: "target not dir",
			setup: func(t *testing.T, tmp string) (string, string) {
				t.Helper()
				source := filepath.Join(tmp, "source")
				makeTrampolineTestDir(t, source)
				target := filepath.Join(tmp, "target")
				makeTrampolineTestFile(t, target)
				return source, target
			},
			wantErr: func(_ string, toDir string) string {
				return "target path is not a directory: " + toDir
			},
		},
		{
			name: "source target same",
			setup: func(t *testing.T, tmp string) (string, string) {
				t.Helper()
				source := filepath.Join(tmp, "source")
				makeTrampolineTestDir(t, source)
				return source, source
			},
			wantErr: func(fromDir string, _ string) string {
				return "source and target directories must differ: " + fromDir
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			fromDir, toDir := tc.setup(t, tmp)

			_, err := SyncTrampolines(fromDir, toDir)
			if err == nil {
				t.Fatal("SyncTrampolines() error = nil, want error")
			}
			if err.Error() != tc.wantErr(fromDir, toDir) {
				t.Fatalf("SyncTrampolines() error = %q, want %q", err.Error(), tc.wantErr(fromDir, toDir))
			}
		})
	}
}
