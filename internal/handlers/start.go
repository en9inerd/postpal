package handlers

import "github.com/en9inerd/telekit"

func (h *Handlers) handleStart(ctx *telekit.Context) error {
	msg := `Available commands:
/address - Show current and next post address
/delete_post ids=123,456 [revoke=true] - Delete post(s) from blog
/sync_channel_info [logo=true] - Sync channel info`

	return ctx.Reply(msg)
}
