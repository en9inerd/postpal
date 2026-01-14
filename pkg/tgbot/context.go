package tgbot

import (
	"context"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

// Context provides access to the current update and utility methods.
type Context struct {
	context.Context

	bot     *Bot
	message *tg.Message
	update  tg.UpdateClass

	// Parsed command parameters (nil if not a command)
	params ParsedParams

	// For album handling
	messages []*tg.Message
}

// Message returns the current message.
func (c *Context) Message() *tg.Message {
	return c.message
}

// Messages returns all messages (for albums, otherwise single message).
func (c *Context) Messages() []*tg.Message {
	if len(c.messages) > 0 {
		return c.messages
	}
	if c.message != nil {
		return []*tg.Message{c.message}
	}
	return nil
}

// Update returns the raw update.
func (c *Context) Update() tg.UpdateClass {
	return c.update
}

// Text returns the message text.
func (c *Context) Text() string {
	if c.message != nil {
		return c.message.Message
	}
	return ""
}

// MessageID returns the message ID.
func (c *Context) MessageID() int {
	if c.message != nil {
		return c.message.ID
	}
	return 0
}

// ChatID returns the chat ID where the message was sent.
func (c *Context) ChatID() int64 {
	if c.message == nil {
		return 0
	}
	switch peer := c.message.PeerID.(type) {
	case *tg.PeerChannel:
		return peer.ChannelID
	case *tg.PeerChat:
		return peer.ChatID
	case *tg.PeerUser:
		return peer.UserID
	}
	return 0
}

// SenderID returns the sender's user ID.
func (c *Context) SenderID() int64 {
	if c.message != nil && c.message.FromID != nil {
		if user, ok := c.message.FromID.(*tg.PeerUser); ok {
			return user.UserID
		}
	}
	return 0
}

// IsOutgoing returns true if this is an outgoing message.
func (c *Context) IsOutgoing() bool {
	if c.message != nil {
		return c.message.Out
	}
	return false
}

// IsChannel returns true if the message is from a channel.
func (c *Context) IsChannel() bool {
	if c.message != nil {
		_, ok := c.message.PeerID.(*tg.PeerChannel)
		return ok
	}
	return false
}

// IsPrivate returns true if the message is from a private chat.
func (c *Context) IsPrivate() bool {
	if c.message != nil {
		_, ok := c.message.PeerID.(*tg.PeerUser)
		return ok
	}
	return false
}

// IsGroup returns true if the message is from a group.
func (c *Context) IsGroup() bool {
	if c.message != nil {
		_, ok := c.message.PeerID.(*tg.PeerChat)
		return ok
	}
	return false
}

// Entities returns the message entities (formatting).
func (c *Context) Entities() []tg.MessageEntityClass {
	if c.message != nil {
		return c.message.Entities
	}
	return nil
}

// Media returns the message media (photo, video, etc.).
func (c *Context) Media() tg.MessageMediaClass {
	if c.message != nil {
		return c.message.Media
	}
	return nil
}

// Params returns the parsed command parameters.
func (c *Context) Params() ParsedParams {
	return c.params
}

// Param returns a single parameter value.
func (c *Context) Param(key string) any {
	if c.params != nil {
		return c.params[key]
	}
	return nil
}

// API returns the raw tg.Client for advanced operations.
func (c *Context) API() *tg.Client {
	return c.bot.api
}

// Reply sends a reply to the current message.
func (c *Context) Reply(text string) error {
	if c.message == nil {
		return nil
	}
	sender := message.NewSender(c.bot.api)
	_, err := sender.Reply(c.bot.entities, c.updateToMessage()).Text(c, text)
	return err
}

// Send sends a message to the current chat.
func (c *Context) Send(text string) error {
	if c.message == nil {
		return nil
	}
	sender := message.NewSender(c.bot.api)
	_, err := sender.To(c.inputPeer()).Text(c, text)
	return err
}

// SendTo sends a message to a specific user ID.
func (c *Context) SendTo(userID int64, text string) error {
	sender := message.NewSender(c.bot.api)
	_, err := sender.To(&tg.InputPeerUser{UserID: userID}).Text(c, text)
	return err
}

// inputPeer returns the input peer for the current chat.
func (c *Context) inputPeer() tg.InputPeerClass {
	if c.message == nil {
		return nil
	}
	switch peer := c.message.PeerID.(type) {
	case *tg.PeerChannel:
		return &tg.InputPeerChannel{ChannelID: peer.ChannelID}
	case *tg.PeerChat:
		return &tg.InputPeerChat{ChatID: peer.ChatID}
	case *tg.PeerUser:
		return &tg.InputPeerUser{UserID: peer.UserID}
	}
	return nil
}

// updateToMessage converts the update to a message interface for reply.
func (c *Context) updateToMessage() *tg.UpdateNewMessage {
	return &tg.UpdateNewMessage{Message: c.message}
}

// CallbackContext provides access to callback query data.
type CallbackContext struct {
	context.Context

	bot    *Bot
	query  *tg.UpdateBotCallbackQuery
	data   string
	userID int64
	msgID  int
	chatID int64
}

// Query returns the raw callback query.
func (c *CallbackContext) Query() *tg.UpdateBotCallbackQuery {
	return c.query
}

// Data returns the callback data string.
func (c *CallbackContext) Data() string {
	return c.data
}

// UserID returns the user who clicked the button.
func (c *CallbackContext) UserID() int64 {
	return c.userID
}

// MessageID returns the message ID containing the button.
func (c *CallbackContext) MessageID() int {
	return c.msgID
}

// ChatID returns the chat ID where the button was clicked.
func (c *CallbackContext) ChatID() int64 {
	return c.chatID
}

// Answer sends an answer to the callback query (toast/alert).
func (c *CallbackContext) Answer(text string, alert bool) error {
	_, err := c.bot.api.MessagesSetBotCallbackAnswer(c, &tg.MessagesSetBotCallbackAnswerRequest{
		QueryID:   c.query.QueryID,
		Message:   text,
		Alert:     alert,
		CacheTime: 0,
	})
	return err
}

// AnswerEmpty acknowledges the callback without showing anything.
func (c *CallbackContext) AnswerEmpty() error {
	return c.Answer("", false)
}

// API returns the raw tg.Client for advanced operations.
func (c *CallbackContext) API() *tg.Client {
	return c.bot.api
}

// DeleteContext provides access to deleted message info.
type DeleteContext struct {
	context.Context

	bot        *Bot
	messageIDs []int
	chatID     int64
	channelID  int64
}

// MessageIDs returns the IDs of deleted messages.
func (c *DeleteContext) MessageIDs() []int {
	return c.messageIDs
}

// ChatID returns the chat ID (for non-channel deletes).
func (c *DeleteContext) ChatID() int64 {
	return c.chatID
}

// ChannelID returns the channel ID (for channel deletes).
func (c *DeleteContext) ChannelID() int64 {
	return c.channelID
}

// API returns the raw tg.Client for advanced operations.
func (c *DeleteContext) API() *tg.Client {
	return c.bot.api
}
