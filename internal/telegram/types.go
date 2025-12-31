package telegram

import "github.com/en9inerd/go-pkgs/validator"

// Message represents a Telegram message
type Message struct {
	MessageID int64  `json:"message_id"`
	Date      int64  `json:"date"`
	Chat      *Chat  `json:"chat,omitempty"`
	Text      string `json:"text,omitempty"`
	Caption   string `json:"caption,omitempty"`
	From      *User  `json:"from,omitempty"`
}

// Chat represents a Telegram chat (channel, group, etc.)
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"` // "channel", "group", "supergroup", "private"
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// User represents a Telegram user
type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// APIResponse represents a generic Telegram API response
// Result is a generic interface{} to handle different response types
// (Message for most operations, bool for delete/pin operations, etc.)
type APIResponse struct {
	OK          bool        `json:"ok"`
	Description string      `json:"description,omitempty"`
	ErrorCode   int         `json:"error_code,omitempty"`
	Result      interface{} `json:"result,omitempty"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	ChatID                string `json:"chat_id"`                          // Channel username (e.g., "@channel") or ID
	Text                  string `json:"text"`                             // Message text
	ParseMode             string `json:"parse_mode,omitempty"`             // "HTML", "Markdown", "MarkdownV2"
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
	DisableNotification   bool   `json:"disable_notification,omitempty"`
	ReplyToMessageID      int64  `json:"reply_to_message_id,omitempty"`
}

// Validate validates the SendMessageRequest
func (r *SendMessageRequest) Validate(v *validator.Validator) {
	v.CheckField(validator.NotBlank(r.ChatID), "chat_id", "chat_id is required")
	v.CheckField(validator.NotBlank(r.Text), "text", "text is required")
	v.CheckField(validator.MaxChars(r.Text, 4096), "text", "text must be 4096 characters or less")
	if r.ParseMode != "" {
		v.CheckField(validator.PermittedValue(r.ParseMode, "HTML", "Markdown", "MarkdownV2"), "parse_mode", "parse_mode must be HTML, Markdown, or MarkdownV2")
	}
}

// EditMessageTextRequest represents a request to edit message text
type EditMessageTextRequest struct {
	ChatID                string `json:"chat_id"`                          // Channel username or ID
	MessageID             int64  `json:"message_id"`                       // Message ID to edit
	Text                  string `json:"text"`                             // New text
	ParseMode             string `json:"parse_mode,omitempty"`             // "HTML", "Markdown", "MarkdownV2"
	DisableWebPagePreview bool   `json:"disable_web_page_preview,omitempty"`
	InlineMessageID       string `json:"inline_message_id,omitempty"`      // For inline messages
}

// Validate validates the EditMessageTextRequest
func (r *EditMessageTextRequest) Validate(v *validator.Validator) {
	v.CheckField(validator.NotBlank(r.ChatID) || validator.NotBlank(r.InlineMessageID), "chat_id", "chat_id or inline_message_id is required")
	v.CheckField(validator.NotBlank(r.Text), "text", "text is required")
	v.CheckField(validator.MaxChars(r.Text, 4096), "text", "text must be 4096 characters or less")
	if r.ParseMode != "" {
		v.CheckField(validator.PermittedValue(r.ParseMode, "HTML", "Markdown", "MarkdownV2"), "parse_mode", "parse_mode must be HTML, Markdown, or MarkdownV2")
	}
	if r.MessageID == 0 && r.InlineMessageID == "" {
		v.AddNonFieldError("either message_id or inline_message_id must be provided")
	}
}

// EditMessageCaptionRequest represents a request to edit message caption
type EditMessageCaptionRequest struct {
	ChatID          string `json:"chat_id"`                    // Channel username or ID
	MessageID       int64  `json:"message_id"`                 // Message ID to edit
	Caption         string `json:"caption,omitempty"`          // New caption
	ParseMode       string `json:"parse_mode,omitempty"`       // "HTML", "Markdown", "MarkdownV2"
	InlineMessageID string `json:"inline_message_id,omitempty"` // For inline messages
}

// Validate validates the EditMessageCaptionRequest
func (r *EditMessageCaptionRequest) Validate(v *validator.Validator) {
	v.CheckField(validator.NotBlank(r.ChatID) || validator.NotBlank(r.InlineMessageID), "chat_id", "chat_id or inline_message_id is required")
	if r.ParseMode != "" {
		v.CheckField(validator.PermittedValue(r.ParseMode, "HTML", "Markdown", "MarkdownV2"), "parse_mode", "parse_mode must be HTML, Markdown, or MarkdownV2")
	}
	if r.MessageID == 0 && r.InlineMessageID == "" {
		v.AddNonFieldError("either message_id or inline_message_id must be provided")
	}
}

// EditMessageMediaRequest represents a request to edit message media
type EditMessageMediaRequest struct {
	ChatID          string      `json:"chat_id"`                    // Channel username or ID
	MessageID       int64       `json:"message_id"`                 // Message ID to edit
	Media           interface{} `json:"media"`                       // InputMedia object
	InlineMessageID string      `json:"inline_message_id,omitempty"` // For inline messages
}

// DeleteMessageRequest represents a request to delete a message
type DeleteMessageRequest struct {
	ChatID    string `json:"chat_id"`    // Channel username or ID
	MessageID int64  `json:"message_id"` // Message ID to delete
}

// Validate validates the DeleteMessageRequest
func (r *DeleteMessageRequest) Validate(v *validator.Validator) {
	v.CheckField(validator.NotBlank(r.ChatID), "chat_id", "chat_id is required")
	v.CheckField(r.MessageID > 0, "message_id", "message_id must be greater than 0")
}

// ForwardMessageRequest represents a request to forward a message
type ForwardMessageRequest struct {
	ChatID              string `json:"chat_id"`                          // Target channel username or ID
	FromChatID          string `json:"from_chat_id"`                     // Source channel username or ID
	MessageID           int64  `json:"message_id"`                       // Message ID to forward
	DisableNotification bool   `json:"disable_notification,omitempty"`
}

// Validate validates the ForwardMessageRequest
func (r *ForwardMessageRequest) Validate(v *validator.Validator) {
	v.CheckField(validator.NotBlank(r.ChatID), "chat_id", "chat_id is required")
	v.CheckField(validator.NotBlank(r.FromChatID), "from_chat_id", "from_chat_id is required")
	v.CheckField(r.MessageID > 0, "message_id", "message_id must be greater than 0")
}

// CopyMessageRequest represents a request to copy a message
type CopyMessageRequest struct {
	ChatID                string `json:"chat_id"`                          // Target channel username or ID
	FromChatID            string `json:"from_chat_id"`                     // Source channel username or ID
	MessageID             int64  `json:"message_id"`                       // Message ID to copy
	Caption               string `json:"caption,omitempty"`                // New caption (optional)
	ParseMode             string `json:"parse_mode,omitempty"`            // "HTML", "Markdown", "MarkdownV2"
	DisableNotification   bool   `json:"disable_notification,omitempty"`
	ReplyToMessageID      int64  `json:"reply_to_message_id,omitempty"`
}

// Validate validates the CopyMessageRequest
func (r *CopyMessageRequest) Validate(v *validator.Validator) {
	v.CheckField(validator.NotBlank(r.ChatID), "chat_id", "chat_id is required")
	v.CheckField(validator.NotBlank(r.FromChatID), "from_chat_id", "from_chat_id is required")
	v.CheckField(r.MessageID > 0, "message_id", "message_id must be greater than 0")
	if r.ParseMode != "" {
		v.CheckField(validator.PermittedValue(r.ParseMode, "HTML", "Markdown", "MarkdownV2"), "parse_mode", "parse_mode must be HTML, Markdown, or MarkdownV2")
	}
}

// PinChatMessageRequest represents a request to pin a message
type PinChatMessageRequest struct {
	ChatID              string `json:"chat_id"`                          // Channel username or ID
	MessageID           int64  `json:"message_id"`                       // Message ID to pin
	DisableNotification bool   `json:"disable_notification,omitempty"`
}

// Validate validates the PinChatMessageRequest
func (r *PinChatMessageRequest) Validate(v *validator.Validator) {
	v.CheckField(validator.NotBlank(r.ChatID), "chat_id", "chat_id is required")
	v.CheckField(r.MessageID > 0, "message_id", "message_id must be greater than 0")
}

// UnpinChatMessageRequest represents a request to unpin a message
type UnpinChatMessageRequest struct {
	ChatID    string `json:"chat_id"`    // Channel username or ID
	MessageID int64  `json:"message_id"` // Message ID to unpin (optional, if not provided unpins all)
}

// Validate validates the UnpinChatMessageRequest
func (r *UnpinChatMessageRequest) Validate(v *validator.Validator) {
	v.CheckField(validator.NotBlank(r.ChatID), "chat_id", "chat_id is required")
	// MessageID is optional for this request (0 means unpin all)
}

// UnpinAllChatMessagesRequest represents a request to unpin all messages
type UnpinAllChatMessagesRequest struct {
	ChatID string `json:"chat_id"` // Channel username or ID
}

// Validate validates the UnpinAllChatMessagesRequest
func (r *UnpinAllChatMessagesRequest) Validate(v *validator.Validator) {
	v.CheckField(validator.NotBlank(r.ChatID), "chat_id", "chat_id is required")
}
