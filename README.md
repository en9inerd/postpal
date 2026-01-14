# PostPal

A Go bot that syncs Telegram channel posts to a Zola static blog. Built with MTProto via [gotd/td](https://github.com/gotd/td).

Example Zola theme: [after-dark](https://github.com/en9inerd/after-dark)

## How It Works

```
┌─────────────────────┐     ┌─────────────────┐     ┌─────────────────────┐
│  Telegram Channel   │     │    PostPal      │     │  GitHub Repository  │
│  (New/Edit Post)    │────▶│  (MTProto Bot)  │     │    (Zola Blog)      │
└─────────────────────┘     └─────────────────┘     └─────────────────────┘
                                    │                         │
                                    ▼                         │
                            ┌─────────────────┐               │
                            │ Create/Update   │               │
                            │ Markdown Post   │               │
                            └─────────────────┘               │
                                    │                         │
                                    ▼                         │
                            ┌─────────────────┐               │
                            │  Git Commit &   │──────────────▶│
                            │     Push        │               │
                            └─────────────────┘               │
                                                              ▼
                                              ┌───────────────────────────┐
                                              │   GitHub Actions          │
                                              │   (Build & Deploy)        │
                                              └───────────────────────────┘
                                                              │
                                                              ▼
                                              ┌───────────────────────────┐
                                              │   Zola Blog Website       │
                                              │   (Post Published)        │
                                              └───────────────────────────┘
```

## Features

- **Channel → Blog Sync**: Automatically creates Zola posts from new channel messages
- **Edit Sync**: Updates blog posts when channel messages are edited
- **Album Support**: Handles grouped media (albums) as single posts with multiple images
- **Media Download**: Downloads photos and video thumbnails
- **Commands**: `/delete_post`, `/sync_channel_info` for management
- **Git Integration**: Auto-commits and pushes changes to your blog repo

## Quick Start

### Prerequisites

- Go 1.25+ (for building from source)
- Telegram API credentials from [my.telegram.org](https://my.telegram.org)
- A bot token from [@BotFather](https://t.me/botfather)
- A Git repository for your Zola blog

### Configuration

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

Required environment variables:

| Variable | Description |
|----------|-------------|
| `TELEGRAM_API_ID` | API ID from my.telegram.org |
| `TELEGRAM_API_HASH` | API Hash from my.telegram.org |
| `TELEGRAM_BOT_TOKEN` | Bot token from @BotFather |
| `TELEGRAM_CHANNEL` | Channel to monitor (`@username` or numeric ID) |
| `TELEGRAM_AUTHOR` | Authorized user (`@username` or numeric ID) |
| `GIT_REPO_URL` | Git repository URL |
| `GIT_AUTH_TOKEN` | Git authentication token |

### Docker (Recommended)

```bash
# Using pre-built image
docker compose up -d

# Or build locally
docker compose -f docker-compose.build.yml up -d --build

# View logs
docker compose logs -f

# Stop
docker compose down
```

### From Source

```bash
git clone https://github.com/en9inerd/postpal.git
cd postpal
make build
./dist/postpal -v
```

## Post Structure

Posts without media:
```
content/posts/123.md
```

Posts with media:
```
content/posts/123/
├── index.md
├── image_0.jpg
└── image_1.png
```

## Commands

| Command | Description |
|---------|-------------|
| `/start` | Show available commands |
| `/delete_post ids=123,456 [revoke=true]` | Delete posts from blog (optionally from Telegram) |
| `/sync_channel_info [logo=true]` | Sync channel logo to blog |

## Project Structure

```
postpal/
├── cmd/app/           # Entry point
├── internal/
│   ├── config/        # Configuration
│   ├── git/           # Git operations
│   ├── handlers/      # Bot event handlers
│   ├── log/           # Logging
│   └── zola/          # Zola post management
└── pkg/tgbot/         # Telegram bot library
```

## Development

```bash
make build           # Build binary
make test            # Run tests
make format          # Format code
make run             # Run with .env
make docker-build    # Build Docker image locally
make docker-up       # Start with pre-built image
make docker-up-build # Build and start
make docker-logs     # View logs
make docker-down     # Stop
```

## License

MIT
