"""Tests for CLI module."""

import sys
from pathlib import Path
from unittest.mock import patch

import pytest

from nix_spotlight.__main__ import main


def test_main_no_args() -> None:
    """Test main with no arguments exits with error."""
    with patch.object(sys, "argv", ["nix-spotlight"]):
        with pytest.raises(SystemExit) as exc_info:
            main()
    assert exc_info.value.code == 2  # argparse error


def test_main_help() -> None:
    """Test main --help exits cleanly."""
    with patch.object(sys, "argv", ["nix-spotlight", "--help"]):
        with pytest.raises(SystemExit) as exc_info:
            main()
    assert exc_info.value.code == 0


def test_main_version(capsys: pytest.CaptureFixture[str]) -> None:
    """Test main --version shows version."""
    with patch.object(sys, "argv", ["nix-spotlight", "--version"]):
        with pytest.raises(SystemExit) as exc_info:
            main()
    assert exc_info.value.code == 0
    captured = capsys.readouterr()
    assert "0.1.0" in captured.out


def test_main_sync_help() -> None:
    """Test sync subcommand help."""
    with patch.object(sys, "argv", ["nix-spotlight", "sync", "--help"]):
        with pytest.raises(SystemExit) as exc_info:
            main()
    assert exc_info.value.code == 0


def test_main_sync_missing_source(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    """Test sync with non-existent source directory."""
    nonexistent = tmp_path / "nonexistent"
    target = tmp_path / "target"

    with patch.object(sys, "argv", ["nix-spotlight", "sync", str(nonexistent), str(target)]):
        result = main()

    assert result == 1
    captured = capsys.readouterr()
    assert "does not exist" in captured.err


def test_main_sync_success(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    """Test successful sync operation."""
    source = tmp_path / "source"
    source.mkdir()
    target = tmp_path / "target"

    # Create a valid app
    app = source / "Test.app"
    app.mkdir()
    (app / "Contents").mkdir()
    (app / "Contents" / "Info.plist").touch()

    with patch.object(
        sys, "argv", ["nix-spotlight", "sync", "--no-dock", str(source), str(target)]
    ):
        result = main()

    assert result == 0
    captured = capsys.readouterr()
    assert "Synced 1 apps" in captured.out
    assert (target / "Test.app" / "Contents").is_symlink()


def test_main_sync_empty_source(tmp_path: Path, capsys: pytest.CaptureFixture[str]) -> None:
    """Test sync with empty source directory."""
    source = tmp_path / "source"
    source.mkdir()
    target = tmp_path / "target"

    with patch.object(
        sys, "argv", ["nix-spotlight", "sync", "--no-dock", str(source), str(target)]
    ):
        result = main()

    assert result == 0
    captured = capsys.readouterr()
    assert "Synced 0 apps" in captured.out


def test_main_sync_with_dock(tmp_path: Path) -> None:
    """Test sync with dock syncing enabled (mocked)."""
    source = tmp_path / "source"
    source.mkdir()
    target = tmp_path / "target"

    # Create a valid app
    app = source / "Test.app"
    app.mkdir()
    (app / "Contents").mkdir()
    (app / "Contents" / "Info.plist").touch()

    with (
        patch.object(sys, "argv", ["nix-spotlight", "sync", str(source), str(target)]),
        patch("nix_spotlight.__main__.sync_dock") as mock_dock,
    ):
        result = main()

    assert result == 0
    mock_dock.assert_called_once()
