# LazySound

LazySound is a terminal UI for SoundCloud: browse, search, manage playlists, and play tracks through an mpv‑backed daemon so playback survives TUI restarts.

## Features
- Browse **My Tracks**, **Liked Tracks**, **My Playlists**, **Liked Playlists**
- Search tracks and playlists
- Playlist drill‑down (view playlist tracks)
- Quick actions:
  - Playlists: create, rename, delete (own), like/unlike
  - Tracks: like/unlike, add to playlist, remove from playlist
- Playback via **mpv** through a background daemon (survives TUI restarts)

## Quick Start (Developers)

### 1) Install dependencies
```
brew install mpv
```
On Windows, install `mpv` and ensure it is in your `PATH`.

### 2) Configure
Config is a single JSON file at:
```
~/.config/lazysound/config.json
```

Example:
```json
{
  "soundcloud": {
    "client_id": "YOUR_ID",
    "client_secret": "YOUR_SECRET",
    "redirect_uri": "http://127.0.0.1:7878/callback",
    "scope": "*"
  },
  "keymap": {
    "play": ["p"],
    "stop": ["s"]
  },
  "theme": {
    "text": "252",
    "muted": "240",
    "accent": "#2b8cff",
    "accent_dim": "#1b4d7a",
    "border": "238",
    "danger": "#ff4d4d",
    "selection_text": "231",
    "progress_fill": "#000000"
  },
  "daemon": {
    "addr": "127.0.0.1:7777"
  }
}
```

### 3) Authenticate (dev flow)
Tokens are stored in the default token store (`~/.soundcli/token.json`).  
Auth flow lives in `lib/soundcloud/auth.go` (`AuthCodePKCE`).

### 4) Run
TUI:
```
go run cmd/lazysoundtui/main.go
```

Daemon (optional, keep playback after closing TUI):
```
go run cmd/lazysoundd/main.go
```

CLI control:
```
go run cmd/lazysoundctl/main.go status
go run cmd/lazysoundctl/main.go pause
go run cmd/lazysoundctl/main.go next
```

### 5) Build local binaries
```
go install ./cmd/lazysoundtui
go install ./cmd/lazysoundd
go install ./cmd/lazysoundctl
```
Ensure `~/go/bin` is in your `PATH`.

## Workspace
- root app module
- `lib/soundcloud`
- `lib/player`
- `lib/uikit`

## Architecture
```
┌───────────────────────────────────────────────────────────────────────┐
│                               lazysoundtui                            │
│  cmd/lazysoundtui/main.go                                             │
│  - load config                                                        │
│  - init client + token                                                │
│  - fetch initial data                                                 │
│  - start TUI                                                          │
└───────────────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌───────────────────────────────────────────────────────────────────────┐
│                                bootstrap                              │
│  internal/app/bootstrap.go                                            │
│  - init client                                                        │
│  - fetch initial data in parallel                                     │
└───────────────────────────────────────────────────────────────────────┘
                         │
                         ▼
┌───────────────────────────────────────────────────────────────────────┐
│                                 UI                                    │
│  internal/ui/model.go                                                 │
│  - navigation + list selection                                        │
│  - search state                                                       │
│  - quick actions state                                                │
│  - playback state                                                     │
│  - renders panes + info                                               │
└───────────────────────────────────────────────────────────────────────┘
        │                         │                          │
        ▼                         ▼                          ▼
┌─────────────────────┐   ┌─────────────────────┐  ┌─────────────────────┐
│      API Client     │   │     UI Components   │  │        Daemon       │
│ lib/soundcloud/*    │   │ lib/uikit/*         │  │ internal/daemon/*   │
│ - SoundCloud HTTP   │   │ - tables/panes      │  │ - queue + mpv IPC   │
└─────────────────────┘   └─────────────────────┘  └─────────────────────┘
                                                     │
                                                     ▼
                                              ┌───────────────────┐
                                              │      Player       │
                                              │ lib/player/*      │
                                              │ - mpv process     │
                                              └───────────────────┘
```

## Project Layout
- `cmd/lazysoundtui/`: TUI entrypoint
- `cmd/lazysoundd/`: playback daemon entrypoint
- `cmd/lazysoundctl/`: CLI controller for playback
- `internal/app/`: config + bootstrap
- `internal/ui/`: TUI state, rendering, and components
- `lib/soundcloud/`: SoundCloud API client + models
- `lib/player/`: mpv control/IPC
- `lib/uikit/`: shared render helpers (`component`, `style`, `layout`, `math`)
- `internal/daemon/`: playback daemon + protocol

## Roadmap
- Auth command (PKCE) wired to CLI
- Pagination for search/large lists
- Queue management and “play next”
- Metadata: artwork + waveform rendering
