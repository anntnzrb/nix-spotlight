"""Dock syncing via dockutil."""

import shutil
import subprocess
from pathlib import Path

from .types import DockSyncResult


def _iter_nix_dock_items(output: str) -> list[tuple[str, str]]:
    """Extract dock item name/path pairs that reference /nix/store."""
    items: list[tuple[str, str]] = []
    for line in output.splitlines():
        if not line.strip():
            continue
        if "/nix/store" not in line:
            continue

        parts = line.split("\t", maxsplit=1)
        name = parts[0].strip()
        path = parts[1].strip() if len(parts) > 1 else ""
        items.append((name, path))

    return items


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
        return DockSyncResult()

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

    for name, path in _iter_nix_dock_items(result.stdout):
        path_stem = Path(path).stem if path else ""

        trampoline = app_stems.get(name) or app_stems.get(path_stem)
        if not trampoline:
            skipped += 1
            continue

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
