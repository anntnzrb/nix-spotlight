"""Type definitions for nix-spotlight."""

from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True, slots=True)
class App:
    """Represents a macOS .app bundle."""

    path: Path

    @property
    def name(self) -> str:
        """Bundle name (e.g., 'MyApp.app')."""
        return self.path.name

    @property
    def contents(self) -> Path:
        """Path to Contents/ directory."""
        return self.path / "Contents"

    @property
    def info_plist(self) -> Path:
        """Path to Info.plist."""
        return self.contents / "Info.plist"

    @property
    def is_valid(self) -> bool:
        """Check if this is a valid .app bundle."""
        return self.info_plist.exists()
