package tgbot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/updates"
	updhook "github.com/gotd/td/telegram/updates/hook"
	"github.com/gotd/td/tg"
)

// Bot is the main Telegram bot client.
type Bot struct {
	config     Config
	client     *telegram.Client
	api        *tg.Client
	dispatcher tg.UpdateDispatcher
	gaps       *updates.Manager
	entities   tg.Entities

	// Handlers
	mu               sync.RWMutex
	messageHandlers  []handler
	editHandlers     []handler
	deleteHandlers   []deleteHandler
	callbackHandlers []callbackHandler
	commandHandlers  []commandHandler
	albumHandlers    []handler

	// Command locking
	commandLock *CommandLock

	// Album collector
	albumCollector *albumCollector

	// Lifecycle callbacks
	onReady func(ctx context.Context)

	// State
	running bool
	selfID  int64
}

// New creates a new Bot instance.
func New(cfg Config) (*Bot, error) {
	cfg.setDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(cfg.SessionDir, 0700); err != nil {
		return nil, err
	}

	dispatcher := tg.NewUpdateDispatcher()

	sessionStorage := &session.FileStorage{
		Path: filepath.Join(cfg.SessionDir, "session.json"),
	}

	gaps := updates.New(updates.Config{
		Handler: &dispatcher,
	})

	client := telegram.NewClient(cfg.APIID, cfg.APIHash, telegram.Options{
		Logger:        cfg.zapLogger(),
		UpdateHandler: gaps,
		Middlewares: []telegram.Middleware{
			updhook.UpdateHook(gaps.Handle),
		},
		Device: telegram.DeviceConfig{
			DeviceModel:    cfg.DeviceModel,
			SystemVersion:  cfg.SystemVersion,
			AppVersion:     cfg.AppVersion,
			LangCode:       cfg.LangCode,
			SystemLangCode: cfg.SystemLangCode,
		},
		SessionStorage: sessionStorage,
	})

	bot := &Bot{
		config:      cfg,
		client:      client,
		dispatcher:  dispatcher,
		gaps:        gaps,
		commandLock: NewCommandLock(),
	}

	bot.albumCollector = newAlbumCollector(cfg.AlbumTimeout, bot.handleAlbum)
	bot.registerDispatcherHandlers()

	return bot, nil
}

// OnReady sets a callback that's called when the bot is connected and ready.
func (b *Bot) OnReady(fn func(ctx context.Context)) {
	b.onReady = fn
}

// OnMessage registers a handler for new messages.
func (b *Bot) OnMessage(filter Filter, fn HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.messageHandlers = append(b.messageHandlers, handler{fn: fn, filter: filter})
}

// OnChannelPost registers a handler for new channel posts.
func (b *Bot) OnChannelPost(channelID int64, fn HandlerFunc) {
	b.OnMessage(Filter{
		Chats:    []int64{channelID},
		Incoming: true,
	}, fn)
}

// OnPrivateMessage registers a handler for private messages from specific users.
func (b *Bot) OnPrivateMessage(userIDs []int64, fn HandlerFunc) {
	b.OnMessage(Filter{
		Users:    userIDs,
		Incoming: true,
	}, fn)
}

// OnEdit registers a handler for edited messages.
func (b *Bot) OnEdit(filter Filter, fn HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.editHandlers = append(b.editHandlers, handler{fn: fn, filter: filter})
}

// OnChannelEdit registers a handler for edited channel posts.
func (b *Bot) OnChannelEdit(channelID int64, fn HandlerFunc) {
	b.OnEdit(Filter{
		Chats: []int64{channelID},
	}, fn)
}

// OnAlbum registers a handler for albums (grouped media).
func (b *Bot) OnAlbum(filter Filter, fn HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.albumHandlers = append(b.albumHandlers, handler{fn: fn, filter: filter})
}

// CommandDef defines a command with its metadata.
type CommandDef struct {
	// Name is the command name without the leading slash.
	Name string

	// Description is shown in the bot's command menu.
	Description string

	// Params defines the parameter schema for validation.
	Params Params

	// Locked enables per-user command locking.
	Locked bool

	// Scope defines where the command is available (default: ScopeDefault).
	Scope CommandScope

	// LangCode is the language code for this command's description.
	LangCode string
}

// Command registers a command handler with optional parameter schema.
func (b *Bot) Command(name string, params Params, fn HandlerFunc) {
	b.CommandWithFilter(CommandDef{Name: name, Params: params}, Filter{Incoming: true}, fn)
}

// CommandWithDesc registers a command with description (for menu sync).
func (b *Bot) CommandWithDesc(def CommandDef, fn HandlerFunc) {
	b.CommandWithFilter(def, Filter{Incoming: true}, fn)
}

// CommandWithFilter registers a command handler with a custom filter.
func (b *Bot) CommandWithFilter(def CommandDef, filter Filter, fn HandlerFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.commandHandlers = append(b.commandHandlers, commandHandler{
		name:        def.Name,
		description: def.Description,
		params:      def.Params,
		fn:          fn,
		filter:      filter,
		locked:      def.Locked,
		scope:       def.Scope,
		langCode:    def.LangCode,
	})
}

// CommandFrom registers a command handler that only responds to specific users.
func (b *Bot) CommandFrom(name string, params Params, userIDs []int64, fn HandlerFunc) {
	b.CommandWithFilter(CommandDef{Name: name, Params: params}, Filter{
		Users:    userIDs,
		Incoming: true,
	}, fn)
}

// LockedCommand registers a command handler with per-user locking.
// Only one instance of a locked command can run per user at a time.
func (b *Bot) LockedCommand(name string, params Params, fn HandlerFunc) {
	b.CommandWithFilter(CommandDef{Name: name, Params: params, Locked: true}, Filter{Incoming: true}, fn)
}

// LockedCommandWithDesc registers a locked command with description.
func (b *Bot) LockedCommandWithDesc(def CommandDef, fn HandlerFunc) {
	def.Locked = true
	b.CommandWithFilter(def, Filter{Incoming: true}, fn)
}

// LockedCommandFrom registers a locked command handler for specific users.
func (b *Bot) LockedCommandFrom(name string, params Params, userIDs []int64, fn HandlerFunc) {
	b.CommandWithFilter(CommandDef{Name: name, Params: params, Locked: true}, Filter{
		Users:    userIDs,
		Incoming: true,
	}, fn)
}

// SyncCommands registers all commands with Telegram so they appear in the bot menu.
// It resets previous command scopes and sets new ones based on registered commands.
// Should be called after all commands are registered and the bot is running.
func (b *Bot) SyncCommands(ctx context.Context) error {
	if b.api == nil {
		return ErrBotNotRunning
	}

	b.mu.RLock()
	handlers := b.commandHandlers
	b.mu.RUnlock()

	// Group commands by scope and language
	grouped := make(map[string][]tg.BotCommand)
	scopes := make(map[string]CommandScope)

	for _, h := range handlers {
		if h.description == "" {
			continue
		}

		scope := h.scope
		if scope == nil {
			scope = ScopeDefault{}
		}

		key := scopeKeyString(scope, h.langCode)
		scopes[key] = scope

		// Check for duplicate command in same scope
		exists := false
		for _, cmd := range grouped[key] {
			if cmd.Command == h.name {
				exists = true
				break
			}
		}
		if !exists {
			grouped[key] = append(grouped[key], tg.BotCommand{
				Command:     h.name,
				Description: h.description,
			})
		}
	}

	if len(grouped) == 0 {
		b.config.Logger.Debug("no commands with descriptions to sync")
		return nil
	}

	// Reset previous scopes first
	if err := b.ResetCommands(ctx); err != nil {
		b.config.Logger.Warn("failed to reset previous commands", "error", err)
	}

	for key, commands := range grouped {
		scope := scopes[key]
		langCode := extractLangCode(key)

		// Resolve scope if needed (e.g., ScopeChannelUsername)
		resolvedScope, err := b.resolveScope(ctx, scope)
		if err != nil {
			b.config.Logger.Error("failed to resolve scope",
				"scope", key,
				"error", err)
			continue
		}

		_, err = b.api.BotsSetBotCommands(ctx, &tg.BotsSetBotCommandsRequest{
			Scope:    resolvedScope,
			LangCode: langCode,
			Commands: commands,
		})
		if err != nil {
			b.config.Logger.Error("failed to set commands",
				"scope", key,
				"error", err)
			continue
		}

		b.config.Logger.Debug("set commands for scope",
			"scope", key,
			"count", len(commands))
	}

	b.config.Logger.Info("synced commands to Telegram",
		"scopes", len(grouped))
	return nil
}

// ResetCommands removes all bot commands from Telegram.
func (b *Bot) ResetCommands(ctx context.Context) error {
	if b.api == nil {
		return ErrBotNotRunning
	}

	scopesToReset := []tg.BotCommandScopeClass{
		&tg.BotCommandScopeDefault{},
		&tg.BotCommandScopeUsers{},
		&tg.BotCommandScopeChats{},
		&tg.BotCommandScopeChatAdmins{},
	}

	for _, scope := range scopesToReset {
		_, err := b.api.BotsResetBotCommands(ctx, &tg.BotsResetBotCommandsRequest{
			Scope:    scope,
			LangCode: "",
		})
		if err != nil {
			// Ignore errors for scopes that may not have commands
			b.config.Logger.Debug("failed to reset scope", "error", err)
		}
	}

	b.config.Logger.Debug("reset bot commands")
	return nil
}

// SetCommandsForScope sets commands for a specific scope and language.
// This is useful for setting commands dynamically without affecting other scopes.
func (b *Bot) SetCommandsForScope(ctx context.Context, scope CommandScope, langCode string, commands []CommandRegistration) error {
	if b.api == nil {
		return ErrBotNotRunning
	}

	if scope == nil {
		scope = ScopeDefault{}
	}

	var tgCommands []tg.BotCommand
	for _, cmd := range commands {
		tgCommands = append(tgCommands, tg.BotCommand{
			Command:     cmd.Name,
			Description: cmd.Description,
		})
	}

	_, err := b.api.BotsSetBotCommands(ctx, &tg.BotsSetBotCommandsRequest{
		Scope:    scope.toTG(),
		LangCode: langCode,
		Commands: tgCommands,
	})
	return err
}

// resolveScope resolves scopes that need lookup (e.g., ScopeChannelUsername, ScopeUsername).
func (b *Bot) resolveScope(ctx context.Context, scope CommandScope) (tg.BotCommandScopeClass, error) {
	switch s := scope.(type) {
	case ScopeChannelUsername:
		resolved, err := b.api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
			Username: s.Username,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to resolve @%s: %w", s.Username, err)
		}

		for _, chat := range resolved.Chats {
			if channel, ok := chat.(*tg.Channel); ok {
				return &tg.BotCommandScopePeer{
					Peer: &tg.InputPeerChannel{
						ChannelID:  channel.ID,
						AccessHash: channel.AccessHash,
					},
				}, nil
			}
		}
		return nil, fmt.Errorf("@%s is not a channel", s.Username)

	case ScopeUsername:
		resolved, err := b.api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
			Username: s.Username,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to resolve @%s: %w", s.Username, err)
		}

		for _, user := range resolved.Users {
			if u, ok := user.(*tg.User); ok {
				return &tg.BotCommandScopePeer{
					Peer: &tg.InputPeerUser{
						UserID:     u.ID,
						AccessHash: u.AccessHash,
					},
				}, nil
			}
		}
		return nil, fmt.Errorf("@%s is not a user", s.Username)

	default:
		return scope.toTG(), nil
	}
}

// ResolveChannel resolves a channel by username and returns its ID and access hash.
// Useful for getting the access hash needed for ScopeChannel.
func (b *Bot) ResolveChannel(ctx context.Context, username string) (channelID, accessHash int64, err error) {
	if b.api == nil {
		return 0, 0, ErrBotNotRunning
	}

	resolved, err := b.api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to resolve @%s: %w", username, err)
	}

	for _, chat := range resolved.Chats {
		if channel, ok := chat.(*tg.Channel); ok {
			return channel.ID, channel.AccessHash, nil
		}
	}
	return 0, 0, fmt.Errorf("@%s is not a channel", username)
}

// ResolveUser resolves a user by username and returns their ID and access hash.
func (b *Bot) ResolveUser(ctx context.Context, username string) (userID, accessHash int64, err error) {
	if b.api == nil {
		return 0, 0, ErrBotNotRunning
	}

	resolved, err := b.api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: username,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to resolve @%s: %w", username, err)
	}

	for _, user := range resolved.Users {
		if u, ok := user.(*tg.User); ok {
			return u.ID, u.AccessHash, nil
		}
	}
	return 0, 0, fmt.Errorf("@%s is not a user", username)
}

// scopeKeyString generates a unique string key for scope + language.
func scopeKeyString(scope CommandScope, langCode string) string {
	if scope == nil {
		scope = ScopeDefault{}
	}

	var scopeName string
	switch s := scope.(type) {
	case ScopeDefault:
		scopeName = "default"
	case ScopeAllPrivate:
		scopeName = "users"
	case ScopeAllGroups:
		scopeName = "chats"
	case ScopeAllGroupAdmins:
		scopeName = "chat_admins"
	case ScopeChat:
		scopeName = fmt.Sprintf("chat:%d", s.ChatID)
	case ScopeChannel:
		scopeName = fmt.Sprintf("channel:%d", s.ChannelID)
	case ScopeChatAdmins:
		scopeName = fmt.Sprintf("chat_admins:%d", s.ChatID)
	case ScopeChatMember:
		scopeName = fmt.Sprintf("chat_member:%d:%d", s.ChatID, s.UserID)
	case ScopeUser:
		scopeName = fmt.Sprintf("user:%d", s.UserID)
	case ScopeUsername:
		scopeName = fmt.Sprintf("user:@%s", s.Username)
	case ScopeChannelUsername:
		scopeName = fmt.Sprintf("channel:@%s", s.Username)
	default:
		scopeName = "unknown"
	}

	return scopeName + "|" + langCode
}

// extractLangCode extracts the language code from a scope key string.
func extractLangCode(key string) string {
	for i := len(key) - 1; i >= 0; i-- {
		if key[i] == '|' {
			return key[i+1:]
		}
	}
	return ""
}

// OnCallback registers a handler for callback queries (inline button clicks).
func (b *Bot) OnCallback(filter CallbackFilter, fn CallbackFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.callbackHandlers = append(b.callbackHandlers, callbackHandler{fn: fn, filter: filter})
}

// OnCallbackPrefix registers a handler for callback queries with a specific data prefix.
func (b *Bot) OnCallbackPrefix(prefix string, fn CallbackFunc) {
	b.OnCallback(CallbackFilter{DataPrefix: prefix}, fn)
}

// OnDelete registers a handler for deleted messages.
func (b *Bot) OnDelete(filter DeleteFilter, fn DeleteFunc) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.deleteHandlers = append(b.deleteHandlers, deleteHandler{fn: fn, filter: filter})
}

// OnChannelDelete registers a handler for deleted channel messages.
func (b *Bot) OnChannelDelete(channelID int64, fn DeleteFunc) {
	b.OnDelete(DeleteFilter{Chats: []int64{channelID}}, fn)
}

// Run starts the bot and blocks until the context is cancelled.
func (b *Bot) Run(ctx context.Context) error {
	if b.running {
		return ErrAlreadyRunning
	}

	return b.client.Run(ctx, func(ctx context.Context) error {
		status, err := b.client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			if _, err := b.client.Auth().Bot(ctx, b.config.BotToken); err != nil {
				return err
			}
		}

		self, err := b.client.Self(ctx)
		if err != nil {
			return err
		}
		b.selfID = self.ID
		b.api = tg.NewClient(b.client)

		b.running = true
		defer func() {
			b.running = false
			b.albumCollector.stop()
		}()

		if b.config.ProfilePhotoURL != "" {
			if err := b.SetProfilePhoto(ctx, b.config.ProfilePhotoURL); err != nil {
				b.config.Logger.Warn("failed to set profile photo", "error", err)
			}
		}

		if b.config.BotInfo != nil {
			if err := b.UpdateBotInfo(ctx, *b.config.BotInfo); err != nil {
				b.config.Logger.Warn("failed to update bot info", "error", err)
			}
		}

		if b.config.SyncCommands {
			if err := b.SyncCommands(ctx); err != nil {
				b.config.Logger.Warn("failed to sync commands", "error", err)
			}
		}

		if b.onReady != nil {
			b.onReady(ctx)
		}

		b.config.Logger.Info("bot started", "id", self.ID, "username", self.Username)

		return b.gaps.Run(ctx, b.api, self.ID, updates.AuthOptions{
			OnStart: func(ctx context.Context) {
				b.config.Logger.Info("listening for updates")
			},
		})
	})
}

// API returns the raw tg.Client for advanced operations.
func (b *Bot) API() *tg.Client {
	return b.api
}

// SelfID returns the bot's user ID.
func (b *Bot) SelfID() int64 {
	return b.selfID
}

// registerDispatcherHandlers sets up the gotd dispatcher handlers.
func (b *Bot) registerDispatcherHandlers() {
	// New channel messages
	b.dispatcher.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewChannelMessage) error {
		b.entities = e
		msg, ok := u.Message.(*tg.Message)
		if !ok {
			return nil
		}
		return b.handleMessage(ctx, msg, u)
	})

	// New private/group messages
	b.dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
		b.entities = e
		msg, ok := u.Message.(*tg.Message)
		if !ok {
			return nil
		}
		return b.handleMessage(ctx, msg, u)
	})

	// Edited channel messages
	b.dispatcher.OnEditChannelMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateEditChannelMessage) error {
		b.entities = e
		msg, ok := u.Message.(*tg.Message)
		if !ok {
			return nil
		}
		return b.handleEdit(ctx, msg, u)
	})

	// Edited messages
	b.dispatcher.OnEditMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateEditMessage) error {
		b.entities = e
		msg, ok := u.Message.(*tg.Message)
		if !ok {
			return nil
		}
		return b.handleEdit(ctx, msg, u)
	})

	// Callback queries (inline button clicks)
	b.dispatcher.OnBotCallbackQuery(func(ctx context.Context, e tg.Entities, u *tg.UpdateBotCallbackQuery) error {
		b.entities = e
		return b.handleCallback(ctx, u)
	})

	// Deleted channel messages
	b.dispatcher.OnDeleteChannelMessages(func(ctx context.Context, e tg.Entities, u *tg.UpdateDeleteChannelMessages) error {
		b.entities = e
		return b.handleDelete(ctx, u.Messages, 0, u.ChannelID)
	})

	// Deleted messages (private/group)
	b.dispatcher.OnDeleteMessages(func(ctx context.Context, e tg.Entities, u *tg.UpdateDeleteMessages) error {
		b.entities = e
		return b.handleDelete(ctx, u.Messages, 0, 0)
	})
}

// handleMessage processes a new message.
func (b *Bot) handleMessage(ctx context.Context, msg *tg.Message, update tg.UpdateClass) error {
	if msg.Out {
		return nil
	}

	if b.albumCollector.add(ctx, msg) {
		return nil
	}

	botCtx := &Context{
		Context: ctx,
		bot:     b,
		message: msg,
		update:  update,
	}

	if strings.HasPrefix(msg.Message, "/") {
		if err := b.handleCommand(botCtx); err != nil {
			b.config.Logger.Error("command handler error", "error", err)
		}
		return nil
	}

	b.mu.RLock()
	handlers := b.messageHandlers
	b.mu.RUnlock()

	for _, h := range handlers {
		if h.filter.matches(botCtx) {
			if err := h.fn(botCtx); err != nil {
				b.config.Logger.Error("message handler error", "error", err)
			}
		}
	}

	return nil
}

// handleEdit processes an edited message.
func (b *Bot) handleEdit(ctx context.Context, msg *tg.Message, update tg.UpdateClass) error {
	botCtx := &Context{
		Context: ctx,
		bot:     b,
		message: msg,
		update:  update,
	}

	b.mu.RLock()
	handlers := b.editHandlers
	b.mu.RUnlock()

	for _, h := range handlers {
		if h.filter.matches(botCtx) {
			if err := h.fn(botCtx); err != nil {
				b.config.Logger.Error("edit handler error", "error", err)
			}
		}
	}

	return nil
}

// handleCommand processes a command message.
func (b *Bot) handleCommand(ctx *Context) error {
	text := ctx.Text()
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return nil
	}

	cmdName := strings.TrimPrefix(parts[0], "/")
	// Handle commands like /cmd@botname
	if idx := strings.Index(cmdName, "@"); idx > 0 {
		cmdName = cmdName[:idx]
	}

	b.mu.RLock()
	handlers := b.commandHandlers
	b.mu.RUnlock()

	for _, h := range handlers {
		if h.name == cmdName && h.filter.matches(ctx) {
			if h.locked {
				userID := ctx.SenderID()
				if userID == 0 {
					continue
				}
				if !b.commandLock.TryLock(userID) {
					return nil
				}
				defer b.commandLock.Unlock(userID)
			}

			params, err := parseParams(text, h.params)
			if err != nil {
				if ctx.SenderID() != 0 {
					_ = ctx.SendTo(ctx.SenderID(), "Error: "+err.Error())
				}
				return nil
			}
			ctx.params = params

			if err := h.fn(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// handleAlbum processes a completed album.
func (b *Bot) handleAlbum(ctx context.Context, messages []*tg.Message) {
	if len(messages) == 0 {
		return
	}

	botCtx := &Context{
		Context:  ctx,
		bot:      b,
		message:  messages[0],
		messages: messages,
	}

	b.mu.RLock()
	handlers := b.albumHandlers
	b.mu.RUnlock()

	for _, h := range handlers {
		if h.filter.matches(botCtx) {
			if err := h.fn(botCtx); err != nil {
				b.config.Logger.Error("album handler error", "error", err)
			}
		}
	}
}

// handleCallback processes a callback query.
func (b *Bot) handleCallback(ctx context.Context, query *tg.UpdateBotCallbackQuery) error {
	data := string(query.Data)

	var chatID int64
	var msgID int
	if peer := query.Peer; peer != nil {
		switch p := peer.(type) {
		case *tg.PeerUser:
			chatID = p.UserID
		case *tg.PeerChat:
			chatID = p.ChatID
		case *tg.PeerChannel:
			chatID = p.ChannelID
		}
	}
	msgID = query.MsgID

	cbCtx := &CallbackContext{
		Context: ctx,
		bot:     b,
		query:   query,
		data:    data,
		userID:  query.UserID,
		msgID:   msgID,
		chatID:  chatID,
	}

	b.mu.RLock()
	handlers := b.callbackHandlers
	b.mu.RUnlock()

	for _, h := range handlers {
		if h.filter.matches(cbCtx) {
			if err := h.fn(cbCtx); err != nil {
				b.config.Logger.Error("callback handler error", "error", err)
			}
		}
	}

	return nil
}

// matches checks if the callback filter matches the given context.
func (f *CallbackFilter) matches(ctx *CallbackContext) bool {
	if f.DataPrefix != "" && !strings.HasPrefix(ctx.data, f.DataPrefix) {
		return false
	}

	if len(f.Users) > 0 {
		found := false
		for _, userID := range f.Users {
			if ctx.userID == userID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if f.Custom != nil && !f.Custom(ctx) {
		return false
	}

	return true
}

// handleDelete processes deleted messages.
func (b *Bot) handleDelete(ctx context.Context, messageIDs []int, chatID, channelID int64) error {
	delCtx := &DeleteContext{
		Context:    ctx,
		bot:        b,
		messageIDs: messageIDs,
		chatID:     chatID,
		channelID:  channelID,
	}

	b.mu.RLock()
	handlers := b.deleteHandlers
	b.mu.RUnlock()

	for _, h := range handlers {
		if h.filter.matches(delCtx) {
			if err := h.fn(delCtx); err != nil {
				b.config.Logger.Error("delete handler error", "error", err)
			}
		}
	}

	return nil
}

// matches checks if the delete filter matches the given context.
func (f *DeleteFilter) matches(ctx *DeleteContext) bool {
	if len(f.Chats) > 0 {
		found := false
		targetID := ctx.channelID
		if targetID == 0 {
			targetID = ctx.chatID
		}
		for _, chatID := range f.Chats {
			if targetID == chatID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if f.Custom != nil && !f.Custom(ctx) {
		return false
	}

	return true
}
