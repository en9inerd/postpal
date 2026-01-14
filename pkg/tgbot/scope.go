package tgbot

import "github.com/gotd/td/tg"

// CommandScope defines where a command should be available.
type CommandScope interface {
	toTG() tg.BotCommandScopeClass
}

// ScopeDefault makes the command available in all private chats.
type ScopeDefault struct{}

func (s ScopeDefault) toTG() tg.BotCommandScopeClass {
	return &tg.BotCommandScopeDefault{}
}

// ScopeAllPrivate makes the command available in all private chats.
type ScopeAllPrivate struct{}

func (s ScopeAllPrivate) toTG() tg.BotCommandScopeClass {
	return &tg.BotCommandScopeUsers{}
}

// ScopeAllGroups makes the command available in all group chats.
type ScopeAllGroups struct{}

func (s ScopeAllGroups) toTG() tg.BotCommandScopeClass {
	return &tg.BotCommandScopeChats{}
}

// ScopeAllGroupAdmins makes the command available to all group admins.
type ScopeAllGroupAdmins struct{}

func (s ScopeAllGroupAdmins) toTG() tg.BotCommandScopeClass {
	return &tg.BotCommandScopeChatAdmins{}
}

// ScopeChat makes the command available in a specific chat.
type ScopeChat struct {
	ChatID int64
}

func (s ScopeChat) toTG() tg.BotCommandScopeClass {
	return &tg.BotCommandScopePeer{
		Peer: &tg.InputPeerChat{ChatID: s.ChatID},
	}
}

// ScopeChannel makes the command available in a specific channel.
// Note: The bot must be an admin in the channel.
type ScopeChannel struct {
	ChannelID  int64
	AccessHash int64 // Optional - will be resolved if zero
}

func (s ScopeChannel) toTG() tg.BotCommandScopeClass {
	return &tg.BotCommandScopePeer{
		Peer: &tg.InputPeerChannel{ChannelID: s.ChannelID, AccessHash: s.AccessHash},
	}
}

// ScopeChatAdmins makes the command available to admins in a specific chat.
type ScopeChatAdmins struct {
	ChatID int64
}

func (s ScopeChatAdmins) toTG() tg.BotCommandScopeClass {
	return &tg.BotCommandScopePeerAdmins{
		Peer: &tg.InputPeerChat{ChatID: s.ChatID},
	}
}

// ScopeChatMember makes the command available to a specific user in a chat.
type ScopeChatMember struct {
	ChatID int64
	UserID int64
}

func (s ScopeChatMember) toTG() tg.BotCommandScopeClass {
	return &tg.BotCommandScopePeerUser{
		Peer:   &tg.InputPeerChat{ChatID: s.ChatID},
		UserID: &tg.InputUser{UserID: s.UserID},
	}
}

// CommandRegistration holds command info for syncing to Telegram.
type CommandRegistration struct {
	// Name is the command name without leading slash.
	Name string

	// Description is shown in the bot menu.
	Description string

	// Scope defines where the command is available.
	// Defaults to ScopeDefault if nil.
	Scope CommandScope

	// LangCode is the language code for this registration.
	// Empty string means all languages.
	LangCode string
}

// ScopeUser makes the command available to a specific user in private chat.
type ScopeUser struct {
	UserID int64
}

func (s ScopeUser) toTG() tg.BotCommandScopeClass {
	return &tg.BotCommandScopePeer{
		Peer: &tg.InputPeerUser{UserID: s.UserID},
	}
}

// ScopeChannelUsername makes the command available in a channel by username.
// This is resolved at sync time.
type ScopeChannelUsername struct {
	Username string // Without @ prefix
}

func (s ScopeChannelUsername) toTG() tg.BotCommandScopeClass {
	// Resolved by Bot.resolveScope() before use
	return nil
}

// ScopeUsername makes the command available to a user by username.
// This is resolved at sync time.
type ScopeUsername struct {
	Username string // Without @ prefix
}

func (s ScopeUsername) toTG() tg.BotCommandScopeClass {
	// Resolved by Bot.resolveScope() before use
	return nil
}
