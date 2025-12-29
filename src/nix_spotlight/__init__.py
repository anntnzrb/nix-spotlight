"""nix-spotlight - macOS Spotlight integration for Nix apps."""

__version__ = "0.1.0"

from .dock import sync_dock
from .trampoline import create_trampoline, sync_trampolines
from .types import App

__all__ = ["App", "__version__", "create_trampoline", "sync_dock", "sync_trampolines"]
