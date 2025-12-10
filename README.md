# Meow

A feature-rich Discord music bot built with Go and discordgo.

## Features

-   Play music from YouTube, Spotify (playlists, albums, tracks), and many other sources
-   Queue management with shuffle, reordering, and removal
-   Playback controls: play, pause, resume, skip, previous, stop
-   Volume control (0-100%)
-   Interactive now playing embeds with button controls
-   High quality audio (128kbps Opus)

## Requirements

-   Go 1.23+
-   FFmpeg
-   yt-dlp
-   opus development libraries (for building)
-   PostgreSQL (optional, for guild settings persistence)
-   Redis (optional, for stream URL caching)

## Environment Variables

Create a `.env` file in the project root:

```
DISCORD_TOKEN=your_bot_token
SPOTIFY_CLIENT_ID=your_spotify_client_id
SPOTIFY_CLIENT_SECRET=your_spotify_client_secret
POSTGRES_URL=postgres://user:pass@localhost:5432/meow
REDIS_URL=redis://localhost:6379
DEFAULT_VOLUME=50
```

Only `DISCORD_TOKEN` is required. Spotify credentials are needed for Spotify URL support.

## Installation

### Local Development

1. Install dependencies:

    - FFmpeg: `brew install ffmpeg` (macOS) or `apt install ffmpeg` (Ubuntu)
    - yt-dlp: `pip install yt-dlp`
    - opus: `brew install opus` (macOS) or `apt install libopus-dev` (Ubuntu)

2. Clone and build:

```bash
git clone https://github.com/dickeyy/meow.git
cd meow
go mod tidy
go build -o meow ./cmd/meow
```

3. Run:

```bash
./meow
```

### Docker

The easiest way to run Meow is with Docker Compose:

```bash
docker compose up -d
```

This will start the bot along with PostgreSQL and Redis containers.

## Commands

| Command                    | Description                                |
| -------------------------- | ------------------------------------------ |
| `/play <query>`            | Play a song or playlist from URL or search |
| `/pause`                   | Pause playback                             |
| `/resume`                  | Resume playback                            |
| `/skip`                    | Skip to next track                         |
| `/previous`                | Go back to previous track                  |
| `/stop`                    | Stop playback and clear queue              |
| `/shuffle`                 | Shuffle the queue                          |
| `/volume <0-100>`          | Set playback volume                        |
| `/queue view`              | View the current queue                     |
| `/queue move <from> <to>`  | Move a track in the queue                  |
| `/queue remove <position>` | Remove a track from the queue              |
| `/queue clear`             | Clear the queue                            |
| `/nowplaying`              | Show the currently playing track           |

## Supported Sources

Thanks to yt-dlp, Meow supports over 1000 sites including:

-   YouTube (videos and playlists)
-   Spotify (tracks, albums, playlists - resolved via YouTube)
-   SoundCloud
-   Bandcamp
-   Vimeo
-   And many more

## Project Structure

```
meow/
├── cmd/meow/           # Entry point
├── internal/
│   ├── audio/          # Audio player and queue
│   ├── bot/            # Discord bot setup
│   ├── commands/       # Slash command handlers
│   ├── config/         # Configuration
│   ├── embeds/         # Discord embed builders
│   ├── services/       # YouTube/Spotify integrations
│   └── storage/        # Database layer
├── Dockerfile
├── docker-compose.yml
└── .env.example
```

## License

MIT - [LICENSE](./LICENSE)
