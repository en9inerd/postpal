# PostPal

> **⚠️ Under Development** - This project is currently under active development and may have incomplete features or breaking changes.

PostPal is a Go service that publishes posts to configured Telegram channels using a Telegram bot. It's designed to work as part of a Zola static website publishing workflow, allowing you to automatically share your blog posts or content to Telegram channels when your site is built or updated.

## Features

- **Telegram Channel Publishing**: Automatically publish posts to configured Telegram channels using a bot
- **Zola Integration**: Designed to work seamlessly with Zola static site generator workflows
- **Robust Error Handling**: Automatic retry logic for transient failures
- **Request Validation**: Built-in validation for all Telegram API requests
- **Structured Logging**: Comprehensive logging for debugging and monitoring
- **HTTP API**: RESTful API for programmatic control
- **Security**: Request size limits, throttling, and security headers

## Project Structure

```
.
├── cmd/
│   └── app/              # Application entry point
│       └── main.go
├── internal/
│   ├── config/           # Configuration parsing
│   ├── log/              # Logging utilities
│   ├── server/           # HTTP server setup and handlers
│   ├── telegram/         # Telegram Bot API client
│   └── validator/        # Validation utilities
├── ui/
│   ├── static/           # Static assets (CSS, JS)
│   └── templates/        # HTML templates
├── scripts/              # Build and utility scripts
├── .github/
│   └── workflows/        # CI/CD workflows
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Getting Started

### Prerequisites

- Go 1.25.1 or later
- A Telegram Bot Token (get one from [@BotFather](https://t.me/botfather))
- A Telegram channel where your bot is an administrator

### Installation

1. Clone the repository:
```bash
git clone https://github.com/en9inerd/postpal.git
cd postpal
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
make build
```

### Configuration

PostPal can be configured via command-line flags or environment variables. Flags take precedence over environment variables.

**Required Configuration:**
- `--telegram-token` or `TELEGRAM_BOT_TOKEN`: Your Telegram bot token

**Optional Configuration:**
- `--port` or `APP_PORT`: Server port (default: `8000`)
- `--verbose` or `-v`: Enable verbose logging

### Running

```bash
# Using command-line flags
./dist/postpal --telegram-token YOUR_BOT_TOKEN --port 8000 --verbose

# Using environment variables
TELEGRAM_BOT_TOKEN=YOUR_BOT_TOKEN APP_PORT=8000 ./dist/postpal --verbose

# Or run directly during development
go run ./cmd/app --telegram-token YOUR_BOT_TOKEN --verbose
```

## Usage

### Telegram Bot Setup

1. Create a bot with [@BotFather](https://t.me/botfather) on Telegram
2. Get your bot token
3. Add the bot as an administrator to your Telegram channel
4. Configure PostPal with your bot token

### Publishing Posts

PostPal provides a Telegram Bot API client for publishing posts to channels. The client supports:

- **Send Message**: Send text messages to channels
- **Edit Message Text**: Edit the text of existing messages
- **Edit Message Caption**: Edit captions of media messages
- **Edit Message Media**: Edit media content of messages
- **Delete Message**: Delete messages from channels
- **Forward Message**: Forward messages between channels
- **Copy Message**: Copy messages to channels
- **Pin Chat Message**: Pin messages in channels
- **Unpin Chat Message**: Unpin specific or all messages in channels

See the [Telegram package documentation](internal/telegram/README.md) for detailed usage examples.

### Zola Integration

PostPal is designed to integrate with Zola static site generation workflows. You can:

1. Set up a webhook or API endpoint that Zola calls after building
2. Use PostPal's HTTP API to publish posts when your site is updated
3. Automate the publishing process as part of your CI/CD pipeline

Example workflow:
```bash
# Build your Zola site
zola build

# Publish to Telegram via PostPal API
curl -X POST http://localhost:8000/api/publish \
  -H "Content-Type: application/json" \
  -d '{"channel": "@your_channel", "text": "New post published!"}'
```

## API Endpoints

The HTTP API provides endpoints for programmatic control:

- `GET /health` - Health check endpoint
- `POST /api/publish` - Publish a post to a Telegram channel (coming soon)

## Features

### Configuration

The configuration system supports both command-line flags and environment variables:

```bash
# Via flag
./postpal --port 8080 --telegram-token YOUR_TOKEN

# Via environment variable
APP_PORT=8080 TELEGRAM_BOT_TOKEN=YOUR_TOKEN ./postpal
```

### Logging

Logging is controlled by the `--verbose` or `-v` flag:

```bash
# Verbose logging (debug level)
./postpal --verbose

# Silent mode (errors only)
./postpal
```

### Docker

PostPal includes Docker support for containerized deployments:

```bash
# Build
docker build -t postpal:latest .

# Run
docker run -p 8000:8000 \
  -e TELEGRAM_BOT_TOKEN=YOUR_TOKEN \
  postpal:latest

# Or use docker-compose
docker-compose up
```

### Multi-Architecture Builds

Build for multiple platforms:

```bash
make build-prod
```

This creates binaries in the `dist/` directory for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)

## Makefile Targets

- `make build` - Build binary locally
- `make build-prod` - Build for multiple platforms
- `make clean` - Remove build artifacts
- `make format` - Format Go code
- `make test` - Run tests
- `make docker-build` - Build Docker image
- `make docker-clean` - Clean up Docker resources
- `make docker-clean-all` - Clean Docker resources and build cache

## Security

PostPal includes several security features:

- Security headers middleware (CSP, X-Frame-Options, etc.)
- Request throttling (1000 concurrent requests)
- Request size limits (10MB max)
- Graceful shutdown
- Health check endpoint

## Dependencies

PostPal uses:
- `github.com/en9inerd/go-pkgs` - Router, middleware, HTTP client, and validation utilities

## Development

### Project Status

⚠️ **This project is under active development.** Features may be incomplete, and breaking changes may occur.

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is provided as-is. Feel free to use and modify for your needs.
