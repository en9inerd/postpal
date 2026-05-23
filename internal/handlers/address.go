package handlers

import (
	"fmt"

	"github.com/en9inerd/telekit"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (h *Handlers) handleAddress(ctx *telekit.Context) error {
	current, next, err := h.zola.GetLatestAddress()
	if err != nil {
		return ctx.Reply(fmt.Sprintf("Error: %v", err))
	}

	sender := message.NewSender(ctx.API())
	peer := &tg.InputPeerUser{UserID: h.author.ID, AccessHash: h.author.AccessHash}

	_, err = sender.To(peer).Reply(ctx.MessageID()).StyledText(ctx,
		styling.Plain("Current: "),
		styling.Code(current),
		styling.Plain("\nNext:    "),
		styling.Code(next),
	)
	return err
}
