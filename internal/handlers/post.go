package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/en9inerd/postpal/internal/zola"
	"github.com/en9inerd/telekit"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

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
			post.Content = telekit.EntitiesToHTML(msg.Message, msg.Entities, telekit.Options{
				HashtagHref: func(tag string) string {
					if h.channel.Username == "" {
						return ""
					}
					return "https://t.me/s/" + h.channel.Username + "?q=%23" + url.QueryEscape(tag)
				},
			})
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
		photo, ok := media.Photo.AsNotEmpty()
		if !ok {
			return nil, nil
		}

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
			ThumbSize:     photo.Sizes[thumbIndex].GetType(),
		}

		if _, err := d.Download(api, loc).Stream(ctx, &buf); err != nil {
			return nil, fmt.Errorf("failed to download photo: %w", err)
		}

	case *tg.MessageMediaDocument:
		doc, ok := media.Document.AsNotEmpty()
		if !ok {
			return nil, nil
		}

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
			loc := doc.AsInputDocumentFileLocation()
			loc.ThumbSize = doc.Thumbs[0].GetType()

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
