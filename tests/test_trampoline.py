"""Tests for trampoline module."""

from collections.abc import Callable
from pathlib import Path

from nix_spotlight.trampoline import create_trampoline, gather_apps, sync_trampolines
from nix_spotlight.types import App


def test_app_properties(make_app: Callable[[str], Path]) -> None:
    """Test App dataclass properties."""
    app_path = make_app("Test.app")
    app = App(app_path)

    assert app.name == "Test.app"
    assert app.contents == app_path / "Contents"
    assert app.info_plist == app_path / "Contents" / "Info.plist"
    assert app.is_valid is True


def test_app_invalid_no_contents(tmp_path: Path) -> None:
    """Test invalid app detection - no Contents directory."""
    app_path = tmp_path / "Invalid.app"
    app_path.mkdir()

    app = App(app_path)
    assert app.is_valid is False


def test_app_invalid_no_plist(tmp_path: Path) -> None:
    """Test invalid app detection - Contents but no Info.plist."""
    app_path = tmp_path / "Invalid.app"
    app_path.mkdir()
    (app_path / "Contents").mkdir()

    app = App(app_path)
    assert app.is_valid is False


def test_create_trampoline(tmp_path: Path) -> None:
    """Test trampoline creation."""
    source_dir = tmp_path / "source"
    source_dir.mkdir()

    app_path = source_dir / "MyApp.app"
    app_path.mkdir()
    (app_path / "Contents").mkdir()
    (app_path / "Contents" / "Info.plist").touch()

    target_dir = tmp_path / "target"
    target_dir.mkdir()

    app = App(app_path)
    trampoline = create_trampoline(app, target_dir)

    assert trampoline == target_dir / "MyApp.app"
    assert trampoline.exists()
    assert (trampoline / "Contents").is_symlink()
    assert (trampoline / "Contents").resolve() == app_path / "Contents"


def test_create_trampoline_replaces_existing(tmp_path: Path) -> None:
    """Test trampoline creation replaces existing symlink."""
    source_dir = tmp_path / "source"
    source_dir.mkdir()
    app_path = source_dir / "MyApp.app"
    app_path.mkdir()
    (app_path / "Contents").mkdir()
    (app_path / "Contents" / "Info.plist").touch()

    target_dir = tmp_path / "target"
    target_dir.mkdir()
    trampoline_dir = target_dir / "MyApp.app"
    trampoline_dir.mkdir()
    # Create an old symlink
    old_link = trampoline_dir / "Contents"
    old_link.symlink_to("/nonexistent")

    app = App(app_path)
    trampoline = create_trampoline(app, target_dir)

    assert (trampoline / "Contents").is_symlink()
    assert (trampoline / "Contents").resolve() == app_path / "Contents"


def test_create_trampoline_creates_parent_dirs(tmp_path: Path) -> None:
    """Test trampoline creation creates parent directories."""
    source_dir = tmp_path / "source"
    source_dir.mkdir()
    app_path = source_dir / "MyApp.app"
    app_path.mkdir()
    (app_path / "Contents").mkdir()
    (app_path / "Contents" / "Info.plist").touch()

    target_dir = tmp_path / "nested" / "target"
    # Don't create target_dir - let create_trampoline do it

    app = App(app_path)
    trampoline = create_trampoline(app, target_dir)

    assert trampoline.exists()
    assert (trampoline / "Contents").is_symlink()


def test_gather_apps(tmp_path: Path) -> None:
    """Test gathering apps from directory."""
    valid_app_names = ["App1.app", "App2.app"]

    # Create valid app at root
    app1 = tmp_path / valid_app_names[0]
    app1.mkdir()
    (app1 / "Contents").mkdir()
    (app1 / "Contents" / "Info.plist").touch()

    # Create nested app
    nested = tmp_path / "Nested"
    nested.mkdir()
    app2 = nested / valid_app_names[1]
    app2.mkdir()
    (app2 / "Contents").mkdir()
    (app2 / "Contents" / "Info.plist").touch()

    # Create invalid app (no Info.plist)
    invalid = tmp_path / "Invalid.app"
    invalid.mkdir()

    apps = gather_apps(tmp_path)

    assert len(apps) == len(valid_app_names)
    names = {app.name for app in apps}
    assert names == set(valid_app_names)


def test_gather_apps_empty_dir(tmp_path: Path) -> None:
    """Test gathering apps from empty directory."""
    apps = gather_apps(tmp_path)
    assert apps == []


def test_gather_apps_no_valid_apps(tmp_path: Path) -> None:
    """Test gathering apps when none are valid."""
    # Create invalid apps only
    invalid1 = tmp_path / "Invalid1.app"
    invalid1.mkdir()

    invalid2 = tmp_path / "Invalid2.app"
    invalid2.mkdir()
    (invalid2 / "Contents").mkdir()  # No Info.plist

    apps = gather_apps(tmp_path)
    assert apps == []


def test_gather_apps_nested_invalid(tmp_path: Path) -> None:
    """Test gathering apps skips invalid nested apps."""
    # Create valid nested app
    nested = tmp_path / "Nested"
    nested.mkdir()
    valid_app = nested / "Valid.app"
    valid_app.mkdir()
    (valid_app / "Contents").mkdir()
    (valid_app / "Contents" / "Info.plist").touch()

    # Create invalid nested app
    invalid_app = nested / "Invalid.app"
    invalid_app.mkdir()
    # No Contents/Info.plist

    apps = gather_apps(tmp_path)
    assert len(apps) == 1
    assert apps[0].name == "Valid.app"


def test_sync_trampolines(tmp_path: Path) -> None:
    """Test full sync operation."""
    source = tmp_path / "source"
    source.mkdir()
    target = tmp_path / "target"

    app_names = ["App1.app", "App2.app"]
    for name in app_names:
        app = source / name
        app.mkdir()
        (app / "Contents").mkdir()
        (app / "Contents" / "Info.plist").touch()

    trampolines = sync_trampolines(source, target)

    assert len(trampolines) == len(app_names)
    assert target.exists()
    for name in app_names:
        assert (target / name / "Contents").is_symlink()


def test_sync_trampolines_cleans_existing(tmp_path: Path) -> None:
    """Test sync removes existing target directory."""
    source = tmp_path / "source"
    source.mkdir()
    target = tmp_path / "target"
    target.mkdir()

    # Create old content that should be removed
    old_app = target / "OldApp.app"
    old_app.mkdir()
    (old_app / "garbage").touch()

    # Create new app
    app = source / "NewApp.app"
    app.mkdir()
    (app / "Contents").mkdir()
    (app / "Contents" / "Info.plist").touch()

    trampolines = sync_trampolines(source, target)

    assert len(trampolines) == 1
    assert not (target / "OldApp.app").exists()
    assert (target / "NewApp.app").exists()


def test_sync_trampolines_empty_source(tmp_path: Path) -> None:
    """Test sync with empty source directory."""
    source = tmp_path / "source"
    source.mkdir()
    target = tmp_path / "target"

    trampolines = sync_trampolines(source, target)

    assert trampolines == []
    assert target.exists()
