"""nix-spotlight - macOS Spotlight integration for Nix apps."""

from importlib.metadata import version

from .dock import sync_dock
from .trampoline import create_trampoline, sync_trampolines
from .types import App, DockSyncResult

__version__ = version("nix-spotlight")

__all__ = [
    "App",
    "DockSyncResult",
    "__version__",
    "create_trampoline",
    "sync_dock",
    "sync_trampolines",
]
