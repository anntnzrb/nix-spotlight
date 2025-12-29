"""Dock syncing via dockutil."""

import shutil
import subprocess
from pathlib import Path

from .types import DockSyncResult


def sync_dock(apps: list[Path], dockutil_path: str | None = None) -> DockSyncResult:
    """Update dock persistent items pointing to /nix/store.

    Finds pinned dock items with /nix/store paths and updates them
    to point to the new trampoline locations.

    Args:
        apps: List of trampoline app paths
        dockutil_path: Path to dockutil binary (auto-detected if None)

    Returns:
        DockSyncResult with counts of updated, skipped items and any errors

    """
    dockutil = dockutil_path or shutil.which("dockutil")
    if not dockutil:
        return DockSyncResult()  # Silently skip if dockutil unavailable

    # Get current dock items
    result = subprocess.run(
        [dockutil, "-L"],
        capture_output=True,
        text=True,
        check=False,
    )

    if result.returncode != 0:
        return DockSyncResult(errors=(f"dockutil -L failed: {result.stderr}",))

    app_stems = {app.stem: app for app in apps}
    updated = 0
    skipped = 0
    errors: list[str] = []

    for line in result.stdout.splitlines():
        # Skip empty lines
        if not line.strip():
            continue

        # Skip non-nix items
        if "/nix/store" not in line:
            continue

        # dockutil format: "AppName\t/path/to/app"
        name = line.split("\t")[0]

        # Find matching trampoline
        if name not in app_stems:
            skipped += 1
            continue

        trampoline = app_stems[name]
        add_result = subprocess.run(
            [dockutil, "--add", str(trampoline.resolve()), "--replacing", name],
            capture_output=True,
            text=True,
            check=False,
        )

        if add_result.returncode != 0:
            errors.append(f"Failed to update {name}: {add_result.stderr}")
        else:
            updated += 1

    return DockSyncResult(updated=updated, skipped=skipped, errors=tuple(errors))
