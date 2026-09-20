# spotumn

A lightning-fast, ultra-lightweight Spotify TUI client built in Go with pure native terminal styling, highlighted borders, synchronized lyrics, listening history, and zero UI deadspots.

---

## Features

- **Pure Terminal Aesthetic**: Runs directly on your terminal's native background colors with zero forced opaque overlays or color deadspots.
- **Dynamic Dark/Light Mode**: Automatically detects and adapts between dark and light terminal themes via Bubble Tea v2 background queries.
- **Colored Block Selection**: Selected tracks, playlists, history rows, and queue items render in solid colored highlight blocks with high-contrast text.
- **Focused Panel Highlighting**: Sidebars and panels (Playlists, Center, Queue, Player) dynamically highlight with a thick Purple border when focused.
- **Topmost Playlist Default Home**: Launches straight into the Tracks tab displaying the contents of your topmost playlist.
- **Listening History Tab (`2`)**: Dedicated History page displaying recently played Spotify tracks with natural autoplay queue.
- **Real-Time Synchronized Lyrics (`3`)**:
  - Auto-follows and drags along with the active singing line in real time as the song plays.
  - Interactive line navigation with `Up`/`Down`, and `Enter` seeking directly to line timestamps.
  - Active singing line highlighted in a glowing Mint block with `❯` pointer.
- **Dedicated Search Bar (`/`)**: Prominently boxed at the very top of the center pane, allowing instant Spotify track searching without flooding your queue.
- **Connected Devices Modal (`d`)**:
  - Live real-time discovery of Spotify Connect devices (auto-polls every 1.5s in the background).
  - Press `r` to trigger an instant rescan with visual feedback without having to close the menu.
  - Transfer playback on `Enter` with live `[Active]` badge update.
- **Playlist Organization**:
  - Category filtering (`f` or `t`): Toggle between `All`, `By You`, `Spotify`, and `Saved`.
  - Pinning (`*`): Pin favorite playlists to the top with a bold gold `★` prefix (persisted to `~/.config/spotumn/pinned.json`).
- **Live Queue**: View all live upcoming tracks with no artificial limits, and append songs with `q`.
- **Framed Window Border**:
  - Top border features `● spotumn` with a green status dot and `👤 username`.
  - Bottom border embeds quick keybindings directly into the frame.
- **Zen Mode (`z`)**: Fullscreen distraction-free music experience with enlarged album artwork, prominent typography, active lyrics, and centered playback controls.
- **Sticky Bottom Player**: Full-width seekbar, volume slider, playback controls, and device status.
- **Ultra-Lightweight (<35MB RAM)**: Strict Go memory limits, aggressive GC, and minimal CPU footprint.

---

## Getting Started

### 1. Building and Installing

Compile and install directly from source:

```bash
cd /media/SSD/code/spotumn
./install.sh
```

Or build manually:

```bash
go build -trimpath -ldflags="-s -w" -o spotumn ./cmd/spotumn
./spotumn
```

### 2. Configuration (Optional)

`spotumn` works out of the box with the standard public Spotify Connect Client ID. If you prefer to use your own Spotify Developer Client ID:

Edit `~/.config/spotumn/config.yml`:

```yaml
# ~/.config/spotumn/config.yml
client_id: "[your_client_id]"
port: 8080
```

---

## Keybindings

| Key | Action |
| --- | --- |
| `Space` | Play / Pause playback |
| `p` / `n` | Previous / Next track |
| `←` / `→` | Seek backward / forward 5s (or `<` / `>`) |
| `+` / `-` (or `=`) | Volume up / down |
| `q` | Add highlighted song to Spotify queue |
| `d` | Open Connect devices popup (`Enter` to select, `r` to rescan, `d`/`Esc` to close) |
| `[` / `]` | Cycle pane focus: `Nav` ↔ `Center` ↔ `Right (Queue)` ↔ `Player` |
| `j` / `k` (or `Down` / `Up`) | Move cursor / navigate items |
| `Enter` | Select playlist / play track / seek to lyric timestamp |
| `1` / `2` / `3` | Switch tabs: 1: Tracks (Default), 2: History, 3: Lyrics |
| `/` | Focus search bar (`Esc` to unfocus) |
| `s` | Toggle shuffle mode |
| `r` | Cycle repeat mode (off / context / track) |
| `f` / `t` | Cycle playlist category (`All`, `By You`, `Spotify`, `Saved`) |
| `*` | Pin / unpin highlighted playlist (`★ ` prefix) |
| `Shift+H` | Hide focused sidebar (restores both if hidden) |
| `z` | Toggle Zen mode (fullscreen enlarged art + lyrics) |
| `?` | Open / close Keybindings help modal |
| `Esc` | Close open modal / reset lyrics scroll / unfocus search |
| `Ctrl+C` | Quit `spotumn` |
