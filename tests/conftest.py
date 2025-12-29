"""Shared test fixtures."""

from collections.abc import Callable
from pathlib import Path

import pytest


@pytest.fixture
def make_app(tmp_path: Path) -> Callable[[str], Path]:
    """Create a valid .app bundle for testing.

    Returns:
        A factory function that creates a valid app with the given name.

    """

    def _make_app(name: str) -> Path:
        app = tmp_path / name
        app.mkdir()
        (app / "Contents").mkdir()
        (app / "Contents" / "Info.plist").touch()
        return app

    return _make_app
