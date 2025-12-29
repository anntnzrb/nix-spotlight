"""Dock syncing via dockutil."""

import shutil
import subprocess
from pathlib import Path


def sync_dock(apps: list[Path], dockutil_path: str | None = None) -> int:
    """Update dock persistent items pointing to /nix/store.

    Finds pinned dock items with /nix/store paths and updates them
    to point to the new trampoline locations.

    Args:
        apps: List of trampoline app paths
        dockutil_path: Path to dockutil binary (auto-detected if None)

    Returns:
        Number of dock items updated

    """
    dockutil = dockutil_path or shutil.which("dockutil")
    if not dockutil:
        return 0  # Silently skip if dockutil unavailable

    # Get current dock items
    result = subprocess.run(
        [dockutil, "-L"],
        capture_output=True,
        text=True,
        check=False,
    )

    if result.returncode != 0:
        return 0

    updated = 0
    app_stems = {app.stem: app for app in apps}

    for line in result.stdout.splitlines():
        # Skip non-nix items
        if "/nix/store" not in line:
            continue

        # dockutil format: "AppName\t/path/to/app"
        name = line.split("\t")[0]

        # Find matching trampoline
        if name in app_stems:
            trampoline = app_stems[name]
            _ = subprocess.run(
                [dockutil, "--add", str(trampoline.resolve()), "--replacing", name],
                check=False,
            )
            updated += 1

    return updated
