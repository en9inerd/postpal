package tgbot

// HandlerFunc is the function signature for event handlers.
type HandlerFunc func(ctx *Context) error

// Filter defines conditions for when a handler should be invoked.
type Filter struct {
	// Chats filters by chat IDs (channels, groups, users).
	// Empty means all chats.
	Chats []int64

	// Users filters by user IDs.
	// Empty means all users.
	Users []int64

	// Incoming filters for incoming messages only.
	Incoming bool

	// Outgoing filters for outgoing messages only.
	Outgoing bool

	// Custom is a custom filter function.
	// Return true to process the message, false to skip.
	Custom func(ctx *Context) bool
}

// handler wraps a handler function with its filter and metadata.
type handler struct {
	fn     HandlerFunc
	filter Filter
}

// commandHandler wraps a command handler with its schema.
type commandHandler struct {
	name        string
	description string
	params      Params
	fn          HandlerFunc
	filter      Filter
	locked      bool
	scope       CommandScope
	langCode    string
}

// callbackHandler wraps a callback query handler.
type callbackHandler struct {
	fn     CallbackFunc
	filter CallbackFilter
}

// CallbackFunc is the function signature for callback query handlers.
type CallbackFunc func(ctx *CallbackContext) error

// CallbackFilter defines conditions for callback query handlers.
type CallbackFilter struct {
	// Data filters by callback data prefix.
	DataPrefix string

	// Users filters by user IDs.
	Users []int64

	// Custom is a custom filter function.
	Custom func(ctx *CallbackContext) bool
}

// deleteHandler wraps a deleted message handler.
type deleteHandler struct {
	fn     DeleteFunc
	filter DeleteFilter
}

// DeleteFunc is the function signature for deleted message handlers.
type DeleteFunc func(ctx *DeleteContext) error

// DeleteFilter defines conditions for delete handlers.
type DeleteFilter struct {
	// Chats filters by chat IDs.
	Chats []int64

	// Custom is a custom filter function.
	Custom func(ctx *DeleteContext) bool
}

// matches checks if the filter matches the given context.
func (f *Filter) matches(ctx *Context) bool {
	if len(f.Chats) > 0 {
		found := false
		for _, chatID := range f.Chats {
			if ctx.ChatID() == chatID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(f.Users) > 0 {
		found := false
		for _, userID := range f.Users {
			if ctx.SenderID() == userID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if f.Incoming && ctx.IsOutgoing() {
		return false
	}
	if f.Outgoing && !ctx.IsOutgoing() {
		return false
	}

	if f.Custom != nil && !f.Custom(ctx) {
		return false
	}

	return true
}
