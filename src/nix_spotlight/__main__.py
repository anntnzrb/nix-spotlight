"""CLI entry point for nix-spotlight."""

import argparse
import sys
from collections.abc import Sequence
from dataclasses import dataclass
from pathlib import Path

from . import __version__
from .dock import sync_dock
from .trampoline import sync_trampolines


@dataclass(frozen=True)
class SyncArgs:
    """Parsed arguments for the sync command."""

    from_dir: Path
    to_dir: Path
    no_dock: bool


class _SyncNamespace(argparse.Namespace):
    from_dir: Path
    to_dir: Path
    no_dock: bool
    command: str

    def __init__(self) -> None:
        super().__init__()
        self.from_dir = Path()
        self.to_dir = Path()
        self.no_dock = False
        self.command = "sync"


def parse_args(argv: Sequence[str] | None = None) -> SyncArgs:
    """Parse CLI arguments."""
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

    namespace = _SyncNamespace()
    _ = parser.parse_args(argv, namespace=namespace)
    return SyncArgs(
        from_dir=namespace.from_dir,
        to_dir=namespace.to_dir,
        no_dock=namespace.no_dock,
    )


def main() -> int:
    """Run the nix-spotlight CLI."""
    args = parse_args()

    try:
        trampolines = sync_trampolines(args.from_dir, args.to_dir)
    except ValueError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 1

    if not args.no_dock:
        dock_result = sync_dock(trampolines)
        if dock_result.errors:
            for error in dock_result.errors:
                print(f"warning: {error}", file=sys.stderr)

    print(f"Synced {len(trampolines)} apps to {args.to_dir}")

    return 0


if __name__ == "__main__":
    sys.exit(main())
