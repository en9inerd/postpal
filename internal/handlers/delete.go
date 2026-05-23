package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/en9inerd/postpal/internal/zola"
	"github.com/en9inerd/telekit"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
)

func (h *Handlers) handleDeletePost(ctx *telekit.Context) error {
	ids := ctx.Params().String("ids")
	revoke := ctx.Params().Bool("revoke")

	h.logger.Info("deleting posts", "ids", ids, "revoke", revoke)

	if err := h.git.Pull(ctx); err != nil {
		h.logger.Warn("pre-operation pull failed", "error", err)
	}

	// Revoke from Telegram first (independent of blog post existence)
	if revoke {
		var msgIDs []int
		for idStr := range strings.SplitSeq(ids, ",") {
			idStr = strings.TrimSpace(idStr)
			if id, err := strconv.Atoi(idStr); err == nil {
				msgIDs = append(msgIDs, id)
			}
		}

		if len(msgIDs) > 0 {
			sender := message.NewSender(ctx.API())
			peer := &tg.InputPeerChannel{ChannelID: h.channel.ID, AccessHash: h.channel.AccessHash}
			if _, err := sender.To(peer).Revoke().Messages(ctx, msgIDs...); err != nil {
				h.logger.Warn("failed to delete messages from channel", "error", err)
			}
		}
	}

	if err := h.zola.DeletePost(ctx, ids); err != nil {
		if errors.Is(err, zola.ErrNoPostsDeleted) {
			if revoke {
				return ctx.Reply(fmt.Sprintf("Revoked post(s) from Telegram: %s (already deleted from blog)", ids))
			}
			return ctx.Reply("No matching posts found")
		}
		h.logger.Error("failed to delete posts", "ids", ids, "error", err)
		return ctx.Reply(fmt.Sprintf("Error deleting post(s): %v", err))
	}

	if err := h.git.CommitAndPush(ctx, fmt.Sprintf("Delete post(s): %s", ids)); err != nil {
		h.logger.Error("failed to commit and push", "error", err)
		return ctx.Reply(fmt.Sprintf("Error committing deletion: %v", err))
	}

	return ctx.Reply(fmt.Sprintf("Deleted post(s): %s", ids))
}
