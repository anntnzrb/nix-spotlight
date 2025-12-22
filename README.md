# nix-spotlight

[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](COPYING)
[![Nix Flake](https://img.shields.io/badge/Nix-Flake-5277C3?logo=nixos)](https://nixos.wiki/wiki/Flakes)
[![macOS](https://img.shields.io/badge/macOS-aarch64%20%7C%20x86__64-000000?logo=apple)](https://github.com/anntnzrb/nix-spotlight)

macOS Spotlight integration for Nix apps.

## What it does

Nix-installed `.app` bundles live in `/nix/store/` which Spotlight doesn't index. This tool creates "trampoline" apps in indexed locations (`~/Applications/`) that Spotlight can find.

Unlike AppleScript-based solutions, nix-spotlight uses symlinks which properly handle URL schemes and file associations.

### Features

- **Spotlight indexing** - Find Nix apps via Cmd+Space
- **URL scheme support** - Links open in the correct app (http, https, mailto, etc.)
- **Finder integration** - "Open With" works correctly
- **Dock persistence** - Pinned apps survive Nix rebuilds
- **Zero configuration** - Works out of the box with Home Manager or nix-darwin

## Requirements

- **macOS** (aarch64 or x86_64)
- **Nix** with flakes enabled
- **Python 3.13+** (provided by the flake)
- **dockutil** (optional) - for automatic Dock item syncing

## Installation

### Home Manager

```nix
{
  inputs.nix-spotlight.url = "github:anntnzrb/nix-spotlight";

  # In your home configuration:
  imports = [ inputs.nix-spotlight.homeManagerModules.default ];
  programs.nix-spotlight.enable = true;
}
```

### nix-darwin

```nix
{
  inputs.nix-spotlight.url = "github:anntnzrb/nix-spotlight";

  # In your darwin configuration:
  imports = [ inputs.nix-spotlight.darwinModules.default ];
  services.nix-spotlight.enable = true;
}
```

## CLI Usage

```bash
# Sync apps from source to trampolines directory
nix-spotlight sync /path/to/apps /path/to/trampolines

# Skip dock syncing
nix-spotlight sync --no-dock /path/to/apps /path/to/trampolines
```

## How it works

For each `.app` in the source directory, nix-spotlight creates:

```
Trampolines/MyApp.app/
└── Contents -> /nix/store/.../MyApp.app/Contents
```

This symlink-based approach (inspired by mac-app-util's `link-contents` branch):
- Gets indexed by Spotlight
- Properly handles URL schemes (http, https, mailto, etc.)
- Works with "Open With" in Finder
- Updates automatically when the Nix store path changes

## Why this exists

This project was born from some minor issues with [mac-app-util](https://github.com/hraban/mac-app-util). While it solves the Spotlight indexing problem, its AppleScript-based trampolines break URL handling - clicking links in other apps wouldn't open my browser (Zen Browser installed via Nix).

The core issue: AppleScript wrappers don't forward URL arguments to the target application. This means `open -a "My Browser" https://example.com` silently drops the URL.

**Why rewrite instead of contributing upstream?**

- mac-app-util is written in Common Lisp - a language with a smaller user base
- Python is ubiquitous, making this codebase accessible to virtually any developer
- Python stdlib is already present on macOS; Common Lisp requires pulling in SBCL
- The fix required a fundamental architecture change (symlinks vs AppleScript), not a patch

## License

[AGPL-3.0](COPYING)
