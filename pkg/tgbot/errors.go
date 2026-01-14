package tgbot

import "errors"

// Configuration errors
var (
	ErrMissingAPIID    = errors.New("tgbot: API ID is required")
	ErrMissingAPIHash  = errors.New("tgbot: API hash is required")
	ErrMissingBotToken = errors.New("tgbot: bot token is required")
)

// Runtime errors
var (
	ErrBotNotRunning  = errors.New("tgbot: bot is not running")
	ErrAlreadyRunning = errors.New("tgbot: bot is already running")
)
