"""CLI entry point for nix-spotlight."""

import argparse
import sys
from pathlib import Path
from typing import cast

from . import __version__
from .dock import sync_dock
from .trampoline import sync_trampolines


def main() -> int:
    """Run the nix-spotlight CLI."""
    parser = argparse.ArgumentParser(
        prog="nix-spotlight",
        description="macOS Spotlight integration for Nix apps",
    )
    _ = parser.add_argument(
        "--version",
        action="version",
        version=f"%(prog)s {__version__}",
    )

    subparsers = parser.add_subparsers(dest="command", required=True)

    # sync command
    sync_parser = subparsers.add_parser(
        "sync",
        help="Sync trampolines from source to target directory",
    )
    _ = sync_parser.add_argument(
        "from_dir",
        type=Path,
        help="Source directory containing .app bundles",
    )
    _ = sync_parser.add_argument(
        "to_dir",
        type=Path,
        help="Target directory for trampolines",
    )
    _ = sync_parser.add_argument(
        "--no-dock",
        action="store_true",
        help="Skip dock syncing",
    )

    args = parser.parse_args()

    # Only sync command exists, subparsers required=True ensures this
    from_dir = cast("Path", args.from_dir)
    to_dir = cast("Path", args.to_dir)
    no_dock = cast("bool", args.no_dock)

    if not from_dir.exists():
        print(f"error: source directory does not exist: {from_dir}", file=sys.stderr)
        return 1

    trampolines = sync_trampolines(from_dir, to_dir)

    if not no_dock:
        _ = sync_dock(trampolines)

    print(f"Synced {len(trampolines)} apps to {to_dir}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
