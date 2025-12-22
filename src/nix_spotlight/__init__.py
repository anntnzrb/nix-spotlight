"""nix-spotlight - macOS Spotlight integration for Nix apps."""

__version__ = "0.1.0"

from .dock import sync_dock
from .trampoline import create_trampoline, sync_trampolines
from .types import App

__all__ = ["__version__", "App", "create_trampoline", "sync_trampolines", "sync_dock"]
