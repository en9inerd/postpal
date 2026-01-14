// Package tgbot provides a high-level framework for building Telegram bots
// using the MTProto protocol via gotd/td.
//
// It simplifies common bot development tasks such as:
//   - Event handling with filters (channels, users, message types)
//   - Command parsing with typed parameter validation
//   - Album (grouped media) handling
//   - Session management
//
// Basic usage:
//
//	bot, err := tgbot.New(tgbot.Config{
//	    APIID:    12345,
//	    APIHash:  "your-api-hash",
//	    BotToken: "your-bot-token",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	bot.OnChannelPost(channelID, func(ctx *tgbot.Context) error {
//	    // Handle new channel post
//	    return nil
//	})
//
//	bot.Command("start", nil, func(ctx *tgbot.Context) error {
//	    return ctx.Reply("Hello!")
//	})
//
//	if err := bot.Run(context.Background()); err != nil {
//	    log.Fatal(err)
//	}
package tgbot
