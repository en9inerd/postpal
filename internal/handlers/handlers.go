package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/en9inerd/postpal/internal/git"
	"github.com/en9inerd/postpal/internal/zola"
	"github.com/en9inerd/telekit"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

// PeerRef identifies a Telegram peer (channel or user) by ID and access hash.
type PeerRef struct {
	ID         int64
	AccessHash int64
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

// New creates a new Handlers instance
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

	h.bot.CommandWithFilter(telekit.CommandDef{
		Name:        "start",
		Description: "Show available commands",
		Scope:       authorScope,
	}, telekit.Filter{Users: []int64{h.author.ID}, Incoming: true}, h.handleStart)

	h.bot.CommandWithFilter(telekit.CommandDef{
		Name:        "delete_post",
		Description: "Delete post(s) from blog",
		Scope:       authorScope,
		Params: telekit.Params{
			"ids":    {Type: telekit.TypeString, Required: true, Description: "Comma-separated post IDs"},
			"revoke": {Type: telekit.TypeBool, Default: false, Description: "Also delete from Telegram"},
		},
		Locked: true,
	}, telekit.Filter{Users: []int64{h.author.ID}, Incoming: true}, h.handleDeletePost)

	h.bot.CommandWithFilter(telekit.CommandDef{
		Name:        "sync_channel_info",
		Description: "Sync channel info",
		Scope:       authorScope,
		Params: telekit.Params{
			"logo": {Type: telekit.TypeBool, Default: true, Description: "Sync channel logo"},
		},
		Locked: true,
	}, telekit.Filter{Users: []int64{h.author.ID}, Incoming: true}, h.handleSyncChannelInfo)
}

func (h *Handlers) handleNewPost(ctx *telekit.Context) error {
	msg := ctx.Message()
	if msg == nil {
		return nil
	}

	// Skip if part of album (handled separately)
	if msg.GroupedID != 0 {
		return nil
	}

	// Skip forwarded messages and service messages
	if _, isFwd := msg.GetFwdFrom(); isFwd || msg.Date == 0 {
		return nil
	}

	// Skip if no content
	if msg.Message == "" && msg.Media == nil {
		return nil
	}

	h.logger.Info("processing new post", "id", msg.ID)

	post, mediaFiles, err := h.processMessages(ctx, []*tg.Message{msg})
	if err != nil {
		h.logger.Error("failed to process message", "error", err)
		return err
	}

	if err := h.zola.CreatePost(ctx, post, mediaFiles); err != nil {
		h.logger.Error("failed to create post", "error", err)
		return err
	}

	if err := h.git.CommitAndPush(ctx, fmt.Sprintf("Add post: %d", post.ID)); err != nil {
		h.logger.Error("failed to commit and push", "error", err)
		return err
	}

	h.logger.Info("created post", "id", post.ID)
	return nil
}

func (h *Handlers) handleEditPost(ctx *telekit.Context) error {
	msg := ctx.Message()
	if msg == nil {
		return nil
	}

	h.logger.Info("processing edited post", "id", msg.ID)

	post, mediaFiles, err := h.processMessages(ctx, []*tg.Message{msg})
	if err != nil {
		h.logger.Error("failed to process edited message", "error", err)
		return err
	}

	var mediaFile []byte
	if len(mediaFiles) > 0 {
		mediaFile = mediaFiles[0]
	}

	if err := h.zola.EditPost(ctx, post, mediaFile); err != nil {
		h.logger.Error("failed to edit post", "error", err)
		return err
	}

	if err := h.git.CommitAndPush(ctx, fmt.Sprintf("Edit post: %d", post.ID)); err != nil {
		h.logger.Error("failed to commit and push", "error", err)
		return err
	}

	h.logger.Info("edited post", "id", post.ID)
	return nil
}

func (h *Handlers) handleAlbum(ctx *telekit.Context) error {
	messages := ctx.Messages()
	if len(messages) < 2 {
		return nil // Single message, not album
	}

	firstMsg := messages[0]
	h.logger.Info("processing album", "id", firstMsg.ID, "count", len(messages))

	post, mediaFiles, err := h.processMessages(ctx, messages)
	if err != nil {
		h.logger.Error("failed to process album", "error", err)
		return err
	}

	if err := h.zola.CreatePost(ctx, post, mediaFiles); err != nil {
		h.logger.Error("failed to create album post", "error", err)
		return err
	}

	if err := h.git.CommitAndPush(ctx, fmt.Sprintf("Add album: %d", post.ID)); err != nil {
		h.logger.Error("failed to commit and push", "error", err)
		return err
	}

	h.logger.Info("created album post", "id", post.ID)
	return nil
}

// processMessages converts Telegram messages to a Zola post
func (h *Handlers) processMessages(ctx *telekit.Context, messages []*tg.Message) (zola.Post, [][]byte, error) {
	if len(messages) == 0 {
		return zola.Post{}, nil, errors.New("no messages to process")
	}

	firstMsg := messages[0]
	post := zola.Post{
		ID:   int64(firstMsg.ID),
		Date: time.Unix(int64(firstMsg.Date), 0),
	}

	var mediaFiles [][]byte
	api := ctx.API()

	for _, msg := range messages {
		if msg.Message != "" {
			htmlContent := zola.EntitiesToHTML(msg.Message, msg.Entities)
			post.Content = htmlContent
		}

		if msg.Media != nil {
			mediaData, err := h.downloadMedia(ctx, api, msg)
			if err != nil {
				h.logger.Warn("failed to download media", "error", err)
				continue
			}
			if mediaData != nil {
				mediaFiles = append(mediaFiles, mediaData)
			}
		}
	}

	return post, mediaFiles, nil
}

func (h *Handlers) downloadMedia(ctx context.Context, api *tg.Client, msg *tg.Message) ([]byte, error) {
	if msg.Media == nil {
		return nil, nil
	}

	d := downloader.NewDownloader()
	var buf bytes.Buffer

	switch media := msg.Media.(type) {
	case *tg.MessageMediaPhoto:
		photo, ok := media.Photo.(*tg.Photo)
		if !ok {
			return nil, nil
		}

		// Get appropriate size (thumb index 2 or last available)
		thumbIndex := 2
		if len(photo.Sizes) <= thumbIndex {
			thumbIndex = len(photo.Sizes) - 1
		}
		if thumbIndex < 0 {
			return nil, nil
		}

		loc := &tg.InputPhotoFileLocation{
			ID:            photo.ID,
			AccessHash:    photo.AccessHash,
			FileReference: photo.FileReference,
			ThumbSize:     getPhotoSizeType(photo.Sizes[thumbIndex]),
		}

		if _, err := d.Download(api, loc).Stream(ctx, &buf); err != nil {
			return nil, fmt.Errorf("failed to download photo: %w", err)
		}

	case *tg.MessageMediaDocument:
		doc, ok := media.Document.(*tg.Document)
		if !ok {
			return nil, nil
		}

		// Check if it's a video/gif - download thumbnail
		isVideo := false
		for _, attr := range doc.Attributes {
			if _, ok := attr.(*tg.DocumentAttributeVideo); ok {
				isVideo = true
				break
			}
			if _, ok := attr.(*tg.DocumentAttributeAnimated); ok {
				isVideo = true
				break
			}
		}

		if isVideo && len(doc.Thumbs) > 0 {
			loc := &tg.InputDocumentFileLocation{
				ID:            doc.ID,
				AccessHash:    doc.AccessHash,
				FileReference: doc.FileReference,
				ThumbSize:     getPhotoSizeType(doc.Thumbs[0]),
			}

			if _, err := d.Download(api, loc).Stream(ctx, &buf); err != nil {
				return nil, fmt.Errorf("failed to download video thumb: %w", err)
			}
		}

	default:
		return nil, nil
	}

	if buf.Len() == 0 {
		return nil, nil
	}

	return buf.Bytes(), nil
}

// getPhotoSizeType extracts the size type string from PhotoSizeClass
func getPhotoSizeType(size tg.PhotoSizeClass) string {
	switch s := size.(type) {
	case *tg.PhotoSize:
		return s.Type
	case *tg.PhotoCachedSize:
		return s.Type
	case *tg.PhotoStrippedSize:
		return s.Type
	case *tg.PhotoSizeProgressive:
		return s.Type
	default:
		return "x"
	}
}

func (h *Handlers) handleStart(ctx *telekit.Context) error {
	msg := `Available commands:
/delete_post ids=123,456 [revoke=true] - Delete post(s) from blog
/sync_channel_info [logo=true] - Sync channel info`

	return ctx.Reply(msg)
}

func (h *Handlers) handleDeletePost(ctx *telekit.Context) error {
	ids := ctx.Params().String("ids")
	revoke := ctx.Params().Bool("revoke")

	h.logger.Info("deleting posts", "ids", ids, "revoke", revoke)

	if err := h.zola.DeletePost(ctx, ids); err != nil {
		if errors.Is(err, zola.ErrNoPostsDeleted) {
			return ctx.Reply("No matching posts found")
		}
		h.logger.Error("failed to delete posts", "ids", ids, "error", err)
		return ctx.Reply(fmt.Sprintf("Error deleting post(s): %v", err))
	}

	if err := h.git.CommitAndPush(ctx, fmt.Sprintf("Delete post(s): %s", ids)); err != nil {
		h.logger.Error("failed to commit and push", "error", err)
		return ctx.Reply(fmt.Sprintf("Error committing deletion: %v", err))
	}

	// Optionally revoke from Telegram
	if revoke {
		var msgIDs []int
		for idStr := range strings.SplitSeq(ids, ",") {
			idStr = strings.TrimSpace(idStr)
			if id, err := strconv.Atoi(idStr); err == nil {
				msgIDs = append(msgIDs, id)
			}
		}

		if len(msgIDs) > 0 {
			_, err := ctx.API().ChannelsDeleteMessages(ctx, &tg.ChannelsDeleteMessagesRequest{
				Channel: &tg.InputChannel{ChannelID: h.channel.ID, AccessHash: h.channel.AccessHash},
				ID:      msgIDs,
			})
			if err != nil {
				h.logger.Warn("failed to delete messages from channel", "error", err)
			}
		}
	}

	return ctx.Reply(fmt.Sprintf("Deleted post(s): %s", ids))
}

func (h *Handlers) handleSyncChannelInfo(ctx *telekit.Context) error {
	syncLogo := ctx.Params().Bool("logo")

	if syncLogo {
		h.logger.Info("syncing channel logo")

		channelFull, err := ctx.API().ChannelsGetFullChannel(ctx, &tg.InputChannel{
			ChannelID:  h.channel.ID,
			AccessHash: h.channel.AccessHash,
		})
		if err != nil {
			errMsg := fmt.Sprintf("Error getting channel info: %s", err.Error())
			return ctx.Reply(errMsg)
		}

		if full, ok := channelFull.FullChat.(*tg.ChannelFull); ok {
			if photo, ok := full.ChatPhoto.(*tg.Photo); ok {
				logoData, err := h.downloadChannelPhoto(ctx, ctx.API(), photo)
				if err != nil {
					h.logger.Error("failed to download channel photo", "error", err)
					return ctx.Reply("Error downloading logo: " + err.Error())
				}

				if logoData != nil {
					if err := h.zola.SaveChannelLogo(logoData); err != nil {
						h.logger.Error("failed to save channel logo", "error", err)
						return ctx.Reply("Error saving logo: " + err.Error())
					}
					h.logger.Info("saved channel logo")
				}
			}
		}

		if err := h.git.CommitAndPush(ctx, "Update channel info"); err != nil {
			h.logger.Error("failed to commit channel info", "error", err)
			return ctx.Reply("Error: " + err.Error())
		}
	}

	return ctx.Reply("Channel info synced")
}

func (h *Handlers) downloadChannelPhoto(ctx context.Context, api *tg.Client, photo *tg.Photo) ([]byte, error) {
	if photo == nil || len(photo.Sizes) == 0 {
		return nil, nil
	}

	// Get the largest size
	lastIdx := len(photo.Sizes) - 1
	loc := &tg.InputPhotoFileLocation{
		ID:            photo.ID,
		AccessHash:    photo.AccessHash,
		FileReference: photo.FileReference,
		ThumbSize:     getPhotoSizeType(photo.Sizes[lastIdx]),
	}

	d := downloader.NewDownloader()
	var buf bytes.Buffer

	if _, err := d.Download(api, loc).Stream(ctx, &buf); err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}

	return buf.Bytes(), nil
}
