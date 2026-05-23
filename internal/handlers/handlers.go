package handlers

import (
	"log/slog"

	"github.com/en9inerd/postpal/internal/git"
	"github.com/en9inerd/postpal/internal/zola"
	"github.com/en9inerd/telekit"
)

// PeerRef identifies a Telegram peer (channel or user) by ID and access hash.
type PeerRef struct {
	ID         int64
	AccessHash int64
	Username   string
}

// Handlers manages bot event handlers
type Handlers struct {
	bot     *telekit.Bot
	git     *git.Service
	zola    *zola.Service
	channel PeerRef
	author  PeerRef
	logger  *slog.Logger
}

func New(bot *telekit.Bot, gitSvc *git.Service, zolaSvc *zola.Service, channel, author PeerRef, logger *slog.Logger) *Handlers {
	return &Handlers{
		bot:     bot,
		git:     gitSvc,
		zola:    zolaSvc,
		channel: channel,
		author:  author,
		logger:  logger,
	}
}

// Register registers all event handlers with the bot
func (h *Handlers) Register() {
	h.bot.OnChannelPost(h.channel.ID, h.handleNewPost)
	h.bot.OnChannelEdit(h.channel.ID, h.handleEditPost)
	h.bot.OnAlbum(telekit.Filter{Chats: []int64{h.channel.ID}, Incoming: true}, h.handleAlbum)

	// Commands (visible only to author in menu, executable only by author)
	authorScope := telekit.ScopeUser{UserID: h.author.ID, AccessHash: h.author.AccessHash}
	authorFilter := telekit.Filter{Users: []int64{h.author.ID}, Incoming: true}

	h.bot.CommandWithFilter(telekit.CommandDef{
		Name:        "start",
		Description: "Show available commands",
		Scope:       authorScope,
	}, authorFilter, h.handleStart)

	h.bot.CommandWithFilter(telekit.CommandDef{
		Name:        "delete_post",
		Description: "Delete post(s) from blog",
		Scope:       authorScope,
		Params: telekit.Params{
			"ids":    {Type: telekit.TypeString, Required: true, Description: "Comma-separated post IDs"},
			"revoke": {Type: telekit.TypeBool, Default: false, Description: "Also delete from Telegram"},
		},
		Locked: true,
	}, authorFilter, h.handleDeletePost)

	h.bot.CommandWithFilter(telekit.CommandDef{
		Name:        "address",
		Description: "Show current and next post address",
		Scope:       authorScope,
	}, authorFilter, h.handleAddress)

	h.bot.CommandWithFilter(telekit.CommandDef{
		Name:        "sync_channel_info",
		Description: "Sync channel info",
		Scope:       authorScope,
		Params: telekit.Params{
			"logo": {Type: telekit.TypeBool, Default: true, Description: "Sync channel logo"},
		},
		Locked: true,
	}, authorFilter, h.handleSyncChannelInfo)
}
