"""Trampoline creation using symlink-based approach."""

import shutil
from pathlib import Path

from .types import App


def create_trampoline(source: App, target_dir: Path) -> Path:
    """Create a symlink-based trampoline for a .app bundle.

    Creates target_dir/AppName.app/Contents -> source.app/Contents

    Args:
        source: The source .app bundle
        target_dir: Directory to create trampoline in

    Returns:
        Path to the created trampoline

    """
    trampoline = target_dir / source.name
    trampoline.mkdir(parents=True, exist_ok=True)

    contents_link = trampoline / "Contents"
    contents_link.unlink(missing_ok=True)
    contents_link.symlink_to(source.contents)

    return trampoline


def gather_apps(from_dir: Path) -> list[App]:
    """Gather all valid .app bundles from a directory.

    Searches one level deep (for nested apps like KDE/).

    Args:
        from_dir: Directory to search

    Returns:
        List of valid App instances

    """
    apps: list[App] = []

    # Direct .app bundles
    for path in from_dir.glob("*.app"):
        app = App(path)
        if app.is_valid:
            apps.append(app)

    # Nested one level (e.g., KDE/Dolphin.app)
    for path in from_dir.glob("*/*.app"):
        app = App(path)
        if app.is_valid:
            apps.append(app)

    return apps


def sync_trampolines(from_dir: Path, to_dir: Path) -> list[Path]:
    """Sync all .app bundles from source to trampolines directory.

    Removes existing trampolines directory and recreates it fresh.

    Args:
        from_dir: Source directory containing .app bundles
        to_dir: Target directory for trampolines

    Returns:
        List of created trampoline paths

    """
    # Clean slate
    shutil.rmtree(to_dir, ignore_errors=True)
    to_dir.mkdir(parents=True)

    apps = gather_apps(from_dir)
    trampolines: list[Path] = []

    for app in apps:
        trampoline = create_trampoline(app, to_dir)
        trampolines.append(trampoline)

    # Touch to trigger Spotlight reindex
    for trampoline in trampolines:
        trampoline.touch()

    return trampolines
