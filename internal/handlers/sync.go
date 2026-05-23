package handlers

import (
	"bytes"
	"context"
	"fmt"

	"github.com/en9inerd/telekit"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
)

func (h *Handlers) handleSyncChannelInfo(ctx *telekit.Context) error {
	syncLogo := ctx.Params().Bool("logo")

	changed := false

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
			if photo, ok := full.ChatPhoto.AsNotEmpty(); ok {
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
					changed = true
				}
			}
		}
	}

	if !changed {
		return ctx.Reply("No changes to sync")
	}

	if err := h.git.CommitAndPush(ctx, "Update channel info"); err != nil {
		h.logger.Error("failed to commit channel info", "error", err)
		return ctx.Reply("Error: " + err.Error())
	}

	return ctx.Reply("Channel info synced")
}

func (h *Handlers) downloadChannelPhoto(ctx context.Context, api *tg.Client, photo *tg.Photo) ([]byte, error) {
	if photo == nil {
		return nil, nil
	}

	last, ok := tg.PhotoSizeClassArray(photo.Sizes).Last()
	if !ok {
		return nil, nil
	}

	loc := &tg.InputPhotoFileLocation{
		ID:            photo.ID,
		AccessHash:    photo.AccessHash,
		FileReference: photo.FileReference,
		ThumbSize:     last.GetType(),
	}

	d := downloader.NewDownloader()
	var buf bytes.Buffer

	if _, err := d.Download(api, loc).Stream(ctx, &buf); err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}

	return buf.Bytes(), nil
}
