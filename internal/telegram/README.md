# Telegram Bot API Service

This package provides a Go client for interacting with the Telegram Bot API, specifically focused on channel post management operations.

## Features

The service wraps the following Telegram Bot API methods for channel post management:

- **Send Message** - Send text messages to channels
- **Edit Message Text** - Edit the text of existing messages
- **Edit Message Caption** - Edit captions of media messages
- **Edit Message Media** - Edit media content of messages
- **Delete Message** - Delete messages from channels
- **Forward Message** - Forward messages between channels
- **Copy Message** - Copy messages to channels
- **Pin Chat Message** - Pin messages in channels
- **Unpin Chat Message** - Unpin specific or all messages in channels

## Usage

### Basic Setup

```go
import (
    "log/slog"
    "github.com/en9inerd/postpal/internal/telegram"
)

// Create a new client
logger := slog.Default()
client := telegram.NewClient("YOUR_BOT_TOKEN", logger)

// Optional: Configure timeout
client = client.WithTimeout(60 * time.Second)
```

### Send a Message

```go
msg, err := client.SendMessage(telegram.SendMessageRequest{
    ChatID:    "@your_channel",  // or channel ID as string
    Text:      "Hello, world!",
    ParseMode: "HTML",            // Optional: "HTML", "Markdown", "MarkdownV2"
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Message sent with ID: %d\n", msg.MessageID)
```

### Edit Message Text

```go
editedMsg, err := client.EditMessageText(telegram.EditMessageTextRequest{
    ChatID:    "@your_channel",
    MessageID: 123,
    Text:      "Updated message text",
    ParseMode: "HTML",
})
if err != nil {
    log.Fatal(err)
}
```

### Edit Message Caption

```go
editedMsg, err := client.EditMessageCaption(telegram.EditMessageCaptionRequest{
    ChatID:    "@your_channel",
    MessageID: 123,
    Caption:   "New caption text",
    ParseMode: "HTML",
})
if err != nil {
    log.Fatal(err)
}
```

### Delete Message

```go
success, err := client.DeleteMessage(telegram.DeleteMessageRequest{
    ChatID:    "@your_channel",
    MessageID: 123,
})
if err != nil {
    log.Fatal(err)
}
if success {
    fmt.Println("Message deleted successfully")
}
```

### Forward Message

```go
forwardedMsg, err := client.ForwardMessage(telegram.ForwardMessageRequest{
    ChatID:     "@target_channel",
    FromChatID: "@source_channel",
    MessageID:  123,
})
if err != nil {
    log.Fatal(err)
}
```

### Copy Message

```go
copiedMsg, err := client.CopyMessage(telegram.CopyMessageRequest{
    ChatID:     "@target_channel",
    FromChatID: "@source_channel",
    MessageID:  123,
    Caption:    "Optional new caption",
})
if err != nil {
    log.Fatal(err)
}
```

### Pin Message

```go
success, err := client.PinChatMessage(telegram.PinChatMessageRequest{
    ChatID:    "@your_channel",
    MessageID: 123,
})
if err != nil {
    log.Fatal(err)
}
```

### Unpin Message

```go
// Unpin a specific message
success, err := client.UnpinChatMessage(telegram.UnpinChatMessageRequest{
    ChatID:    "@your_channel",
    MessageID: 123,
})

// Or unpin all messages
success, err := client.UnpinAllChatMessages(telegram.UnpinAllChatMessagesRequest{
    ChatID: "@your_channel",
})
```

## Configuration

The Telegram Bot Token can be configured via:

1. **Environment Variable**: `TELEGRAM_BOT_TOKEN`
2. **Command-line Flag**: `--telegram-token`

Example:

```bash
# Via environment variable
export TELEGRAM_BOT_TOKEN=your-token-here
./app

# Via command-line flag
./app --telegram-token=your-token-here
```

## Error Handling

All methods return errors that can be checked:

```go
msg, err := client.SendMessage(req)
if err != nil {
    // Handle error - could be network error, API error, etc.
    log.Printf("Failed to send message: %v", err)
    return
}
```

The service handles Telegram API errors and returns descriptive error messages including the error code and description from the Telegram API.

## Types

The package provides comprehensive types for all requests and responses:

- `Message` - Represents a Telegram message
- `Chat` - Represents a Telegram chat/channel
- `User` - Represents a Telegram user
- `APIResponse` - Generic API response wrapper
- Request types for each operation (e.g., `SendMessageRequest`, `EditMessageTextRequest`, etc.)

## Request Validation

All request types implement validation using `go-pkgs/validator`. Requests are automatically validated before being sent to the Telegram API. If validation fails, an error is returned with details about which fields failed validation.

Example:

```go
req := telegram.SendMessageRequest{
    ChatID: "",  // Missing required field
    Text:   "Hello",
}

msg, err := client.SendMessage(req)
if err != nil {
    // err will contain: "validation failed: {"fieldErrors":{"chat_id":["chat_id is required"]}}"
    log.Printf("Validation error: %v", err)
}
```

The validation ensures:
- Required fields are present
- Text length limits (4096 characters for messages)
- Parse mode values are valid ("HTML", "Markdown", "MarkdownV2")
- Message IDs are positive integers
- Proper field combinations (e.g., either `message_id` or `inline_message_id` must be provided)

## Notes

- Channel identifiers can be provided as usernames (e.g., `"@channel"`) or as string IDs (e.g., `"-1001234567890"`)
- The service uses a default timeout of 30 seconds, which can be customized
- All methods are thread-safe and can be called concurrently
- The client logs debug information when making API requests (if logger is configured)
- All requests are automatically validated before being sent to the Telegram API
