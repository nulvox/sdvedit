# SDV Edit

A full-featured Stardew Valley save editor that runs entirely in your browser — no installation, no server, no data ever leaves your machine.

[![CI](https://github.com/nulvox/sdvedit/actions/workflows/ci.yml/badge.svg)](https://github.com/nulvox/sdvedit/actions/workflows/ci.yml)
[![GitHub Sponsors](https://img.shields.io/github/sponsors/nulvox)](https://github.com/sponsors/nulvox)

## Features

- **Player** — name, money, energy, health, total hours played
- **Skills** — farming, mining, foraging, fishing, combat levels and XP
- **Friendships** — hearts and gift counts for every NPC
- **Animals** — friendship, mood, name, produce quality; add new animals to existing barns and coops
- **Farm buildings** — position, paint colors, hay capacity, occupant limits; add new buildings with collision detection
- **Inventory** — item stack sizes and quality
- **World state** — season, day, year, daily luck, mine progress
- **Community Center bundles** — view completion status
- **Cooking & crafting recipes** — learned/unlearned
- **Mail flags** — add or remove received-mail flags
- **Quests** — active quest list

### Building placement collision detection

When adding a building the editor warns about:
- Placement outside the farm boundary (per farm type)
- Permanent terrain obstacles — water, cliffs, forest zones (Standard, Riverland, Forest, Hill-top, Wilderness, Four Corners, Beach farms each have distinct layouts)
- Objects currently on those tiles — trees, crops, stones, weeds, machines, etc. read directly from your save
- Large boulders, logs, and stumps (resource clumps)
- Overlap with existing buildings

All warnings are informational; you can proceed anyway and clear the area in-game first. A "Find a spot" button suggests the first clear position automatically.

## Usage

1. Open the editor at **https://nulvox.github.io/sdvedit**
2. Drag your save file onto the drop zone, or click to browse
   - Save files are usually at `~/.config/StardewValley/Saves/<farmname>/<farmname>` on Linux/Mac or `%APPDATA%\StardewValley\Saves\<farmname>\<farmname>` on Windows
3. Edit values across the tabs
4. Click **Save File** to download the modified save
5. Replace your original save with the downloaded file

> **Back up your save before editing.** The game also keeps a `_old` backup copy in the same folder.

## Tech stack

- **Go → WebAssembly** — all save parsing and editing logic runs as a WASM binary compiled from pure Go (no external Go dependencies)
- **Plain JavaScript** — a minimal JS layer loads the WASM and wires up the UI; no npm, no bundler, no framework
- **GitHub Pages** — static hosting of `site/` directly from the repository

## Local development

**Prerequisites:** Go 1.21+ (tested with 1.26.3)

```sh
git clone https://github.com/nulvox/sdvedit
cd sdvedit
make          # build site/main.wasm
make serve    # serve on http://localhost:8080
make serve PORT=9090  # custom port
make test     # run the Go test suite
```

The Makefile uses the Go toolchain from `GOROOT` (defaults to `/usr/local/go`). Override with `make GOROOT=/path/to/go`.

## Contributing

Pull requests welcome. Please run `make test` before submitting. The test suite covers the XML parser, all accessor functions, animal/building factories, and the collision detection logic — keep it green.

## Support

If this tool saves you time, consider [sponsoring on GitHub](https://github.com/sponsors/nulvox).
