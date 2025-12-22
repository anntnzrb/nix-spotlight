"""Tests for dock module."""

from pathlib import Path
from unittest.mock import MagicMock, patch

from nix_spotlight.dock import sync_dock


def test_sync_dock_no_dockutil() -> None:
    """Test sync_dock when dockutil is not available."""
    apps = [Path("/tmp/App1.app"), Path("/tmp/App2.app")]

    with patch("shutil.which", return_value=None):
        result = sync_dock(apps)

    assert result == 0


def test_sync_dock_with_explicit_path() -> None:
    """Test sync_dock with explicit dockutil path."""
    apps = [Path("/tmp/App1.app")]

    mock_result = MagicMock()
    mock_result.returncode = 0
    mock_result.stdout = ""

    with patch("subprocess.run", return_value=mock_result) as mock_run:
        result = sync_dock(apps, dockutil_path="/usr/local/bin/dockutil")

    assert result == 0
    mock_run.assert_called_once()
    assert mock_run.call_args[0][0] == ["/usr/local/bin/dockutil", "-L"]


def test_sync_dock_dockutil_fails() -> None:
    """Test sync_dock when dockutil command fails."""
    apps = [Path("/tmp/App1.app")]

    mock_result = MagicMock()
    mock_result.returncode = 1

    with (
        patch("shutil.which", return_value="/usr/bin/dockutil"),
        patch("subprocess.run", return_value=mock_result),
    ):
        result = sync_dock(apps)

    assert result == 0


def test_sync_dock_no_nix_items() -> None:
    """Test sync_dock when no nix store items in dock."""
    apps = [Path("/tmp/App1.app")]

    mock_result = MagicMock()
    mock_result.returncode = 0
    mock_result.stdout = "Safari\t/Applications/Safari.app\nChrome\t/Applications/Chrome.app"

    with (
        patch("shutil.which", return_value="/usr/bin/dockutil"),
        patch("subprocess.run", return_value=mock_result),
    ):
        result = sync_dock(apps)

    assert result == 0


def test_sync_dock_updates_matching_items(tmp_path: Path) -> None:
    """Test sync_dock updates matching nix store items."""
    app1 = tmp_path / "MyApp.app"
    app1.mkdir()
    apps = [app1]

    mock_list_result = MagicMock()
    mock_list_result.returncode = 0
    mock_list_result.stdout = "MyApp\t/nix/store/abc123-myapp/Applications/MyApp.app"

    mock_add_result = MagicMock()
    mock_add_result.returncode = 0

    call_count = 0

    def mock_run(cmd: list[str], **kwargs: object) -> MagicMock:
        nonlocal call_count
        call_count += 1
        if "-L" in cmd:
            return mock_list_result
        return mock_add_result

    with (
        patch("shutil.which", return_value="/usr/bin/dockutil"),
        patch("subprocess.run", side_effect=mock_run) as mock_subprocess,
    ):
        result = sync_dock(apps)

    assert result == 1
    assert call_count == 2  # List + Add


def test_sync_dock_empty_line() -> None:
    """Test sync_dock handles empty lines in dockutil output."""
    apps = [Path("/tmp/App1.app")]

    mock_result = MagicMock()
    mock_result.returncode = 0
    mock_result.stdout = "\n\n/nix/store/test\n"

    with (
        patch("shutil.which", return_value="/usr/bin/dockutil"),
        patch("subprocess.run", return_value=mock_result),
    ):
        result = sync_dock(apps)

    assert result == 0
