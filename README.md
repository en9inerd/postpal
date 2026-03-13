# PostPal

[![Docker Hub](https://img.shields.io/docker/v/enginerd/postpal?label=Docker%20Hub&logo=docker&sort=semver)](https://hub.docker.com/r/enginerd/postpal)

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
- **Health Check**: Built-in HTTP health endpoint on port 8080

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
docker compose up -d --build

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

## Zola Theme Requirements

For a theme to be compatible with PostPal, it must handle the following front matter and content format:

### Front Matter (TOML)

```toml
+++
title = "Post Title"
date = 2024-08-21T13:16:34.000Z

[extra]
images = ["image_0.jpg", "image_1.png"]
+++
```

### Theme Requirements

| Requirement | Description |
|-------------|-------------|
| `page.title` | Display the post title |
| `page.date` | Display the publication date (RFC3339 format) |
| `page.extra.images` | Array of co-located image filenames |
| `page.content \| safe` | Content may contain HTML tags |

### Image Handling Example

```jinja2
{% if page.extra.images %}
  {% for image in page.extra.images %}
    <img src="{{ page.colocated_path ~ image }}" alt="{{ page.title }}" />
  {% endfor %}
{% endif %}
```

### Supported Content Elements

- HTML formatting: `<strong>`, `<em>`, `<s>`, `<u>`, `<code>`, `<pre>`, `<blockquote>` (including expandable)
- Links: `<a href="...">` (text URLs, plain URLs, `@mentions`, and mention-by-ID)
- Spoilers: `<span class="spoiler">`
- Fenced code blocks with language hints
- Custom emoji entities are silently skipped

## Commands

| Command | Description |
|---------|-------------|
| `/start` | Show available commands |
| `/delete_post ids=123,456 [revoke=true]` | Delete posts from blog (optionally from Telegram) |
| `/sync_channel_info [logo=true]` | Sync channel logo to blog |

## Project Structure

```
postpal/
├── cmd/postpal/       # Entry point
├── internal/
│   ├── config/        # Configuration
│   ├── git/           # Git operations
│   ├── handlers/      # Bot event handlers
│   ├── log/           # Logging
│   └── zola/          # Zola post management
```

The bot framework lives in the separate [telekit](https://github.com/en9inerd/telekit) module.

## Development

```bash
make build           # Build binary
make build-prod      # Cross-compile for all platforms
make clean           # Remove build artifacts
make format          # Format code
make test            # Run tests
make run             # Run with .env
make run-verbose     # Run with .env and verbose logging
make docker-build    # Build Docker image locally
make docker-up       # Start with pre-built image
make docker-up-build # Build and start locally
make docker-down     # Stop containers
make docker-logs     # View logs
make docker-clean    # Remove Docker containers and images
```

## License

MIT
