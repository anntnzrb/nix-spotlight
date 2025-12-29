"""Tests for dock module."""

from pathlib import Path
from unittest.mock import MagicMock, patch

from nix_spotlight.dock import sync_dock


def test_sync_dock_no_dockutil(tmp_path: Path) -> None:
    """Test sync_dock when dockutil is not available."""
    apps = [tmp_path / "App1.app", tmp_path / "App2.app"]

    with patch("shutil.which", return_value=None):
        result = sync_dock(apps)

    assert result.updated == 0
    assert result.skipped == 0
    assert result.errors == ()


def test_sync_dock_with_explicit_path(tmp_path: Path) -> None:
    """Test sync_dock with explicit dockutil path."""
    apps = [tmp_path / "App1.app"]

    mock_result = MagicMock()
    mock_result.returncode = 0
    mock_result.stdout = ""

    with patch("subprocess.run", return_value=mock_result) as mock_run:
        result = sync_dock(apps, dockutil_path="/usr/local/bin/dockutil")

    assert result.updated == 0
    mock_run.assert_called_once()
    assert mock_run.call_args[0][0] == ["/usr/local/bin/dockutil", "-L"]


def test_sync_dock_dockutil_fails(tmp_path: Path) -> None:
    """Test sync_dock when dockutil command fails."""
    apps = [tmp_path / "App1.app"]

    mock_result = MagicMock()
    mock_result.returncode = 1
    mock_result.stderr = "dockutil error"

    with (
        patch("shutil.which", return_value="/usr/bin/dockutil"),
        patch("subprocess.run", return_value=mock_result),
    ):
        result = sync_dock(apps)

    assert result.updated == 0
    assert len(result.errors) == 1
    assert "dockutil -L failed" in result.errors[0]


def test_sync_dock_no_nix_items(tmp_path: Path) -> None:
    """Test sync_dock when no nix store items in dock."""
    apps = [tmp_path / "App1.app"]

    mock_result = MagicMock()
    mock_result.returncode = 0
    mock_result.stdout = "Safari\t/Applications/Safari.app\nChrome\t/Applications/Chrome.app"

    with (
        patch("shutil.which", return_value="/usr/bin/dockutil"),
        patch("subprocess.run", return_value=mock_result),
    ):
        result = sync_dock(apps)

    assert result.updated == 0
    assert result.skipped == 0


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
    mock_add_result.stderr = ""

    calls: list[list[str]] = []

    def mock_run(cmd: list[str], **_kwargs: object) -> MagicMock:
        calls.append(cmd)
        if "-L" in cmd:
            return mock_list_result
        return mock_add_result

    with (
        patch("shutil.which", return_value="/usr/bin/dockutil"),
        patch("subprocess.run", side_effect=mock_run),
    ):
        result = sync_dock(apps)

    assert result.updated == 1
    assert result.skipped == 0
    assert result.errors == ()
    expected_calls = 1 + 1  # list + add
    assert len(calls) == expected_calls


def test_sync_dock_skips_unmatched_nix_items(tmp_path: Path) -> None:
    """Test sync_dock skips nix items without matching trampoline."""
    app1 = tmp_path / "OtherApp.app"
    app1.mkdir()
    apps = [app1]

    mock_result = MagicMock()
    mock_result.returncode = 0
    mock_result.stdout = "MyApp\t/nix/store/abc123-myapp/Applications/MyApp.app"

    with (
        patch("shutil.which", return_value="/usr/bin/dockutil"),
        patch("subprocess.run", return_value=mock_result),
    ):
        result = sync_dock(apps)

    assert result.updated == 0
    assert result.skipped == 1


def test_sync_dock_reports_add_errors(tmp_path: Path) -> None:
    """Test sync_dock reports errors when add fails."""
    app1 = tmp_path / "MyApp.app"
    app1.mkdir()
    apps = [app1]

    mock_list_result = MagicMock()
    mock_list_result.returncode = 0
    mock_list_result.stdout = "MyApp\t/nix/store/abc123-myapp/Applications/MyApp.app"

    mock_add_result = MagicMock()
    mock_add_result.returncode = 1
    mock_add_result.stderr = "permission denied"

    def mock_run(cmd: list[str], **_kwargs: object) -> MagicMock:
        if "-L" in cmd:
            return mock_list_result
        return mock_add_result

    with (
        patch("shutil.which", return_value="/usr/bin/dockutil"),
        patch("subprocess.run", side_effect=mock_run),
    ):
        result = sync_dock(apps)

    assert result.updated == 0
    assert len(result.errors) == 1
    assert "Failed to update MyApp" in result.errors[0]


def test_sync_dock_empty_line(tmp_path: Path) -> None:
    """Test sync_dock handles empty lines in dockutil output."""
    apps = [tmp_path / "App1.app"]

    mock_result = MagicMock()
    mock_result.returncode = 0
    mock_result.stdout = "\n\n/nix/store/test\n"

    with (
        patch("shutil.which", return_value="/usr/bin/dockutil"),
        patch("subprocess.run", return_value=mock_result),
    ):
        result = sync_dock(apps)

    assert result.updated == 0
