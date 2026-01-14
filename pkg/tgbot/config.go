package tgbot

import (
	"log/slog"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the configuration for the bot.
type Config struct {
	// APIID is the Telegram API ID from https://my.telegram.org
	APIID int

	// APIHash is the Telegram API hash from https://my.telegram.org
	APIHash string

	// BotToken is the bot token from @BotFather
	BotToken string

	// SessionDir is the directory for storing session data.
	// Defaults to "./session" if empty.
	SessionDir string

	// Logger is the logger to use. If nil, a default logger is created.
	Logger *slog.Logger

	// DeviceModel is the device model to report to Telegram.
	// Defaults to "tgbot" if empty.
	DeviceModel string

	// SystemVersion is the system version to report to Telegram.
	// Defaults to "1.0" if empty.
	SystemVersion string

	// AppVersion is the app version to report to Telegram.
	// Defaults to "1.0.0" if empty.
	AppVersion string

	// LangCode is the language code to report to Telegram.
	// Defaults to "en" if empty.
	LangCode string

	// SystemLangCode is the system language code.
	// Defaults to "en" if empty.
	SystemLangCode string

	// AlbumTimeout is the duration to wait for grouped messages.
	// Defaults to 500ms if zero.
	AlbumTimeout time.Duration

	// SyncCommands automatically syncs commands to Telegram on startup.
	// Commands must have descriptions set to be synced.
	SyncCommands bool

	// BotInfo is the bot profile information to set on startup.
	// If set, the bot info will be updated when the bot starts.
	BotInfo *BotInfo

	// ProfilePhotoURL is the URL of the bot's profile photo.
	// If set, the profile photo will be updated when the bot starts.
	ProfilePhotoURL string

	// Verbose enables debug logging for the MTProto client.
	Verbose bool
}

func (c *Config) setDefaults() {
	if c.SessionDir == "" {
		c.SessionDir = "./session"
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
	if c.DeviceModel == "" {
		c.DeviceModel = "tgbot"
	}
	if c.SystemVersion == "" {
		c.SystemVersion = "1.0"
	}
	if c.AppVersion == "" {
		c.AppVersion = "1.0.0"
	}
	if c.LangCode == "" {
		c.LangCode = "en"
	}
	if c.SystemLangCode == "" {
		c.SystemLangCode = "en"
	}
	if c.AlbumTimeout == 0 {
		c.AlbumTimeout = 500 * time.Millisecond
	}
}

func (c *Config) validate() error {
	if c.APIID == 0 {
		return ErrMissingAPIID
	}
	if c.APIHash == "" {
		return ErrMissingAPIHash
	}
	if c.BotToken == "" {
		return ErrMissingBotToken
	}
	return nil
}

// zapLogger creates a zap logger matching the Verbose setting.
func (c *Config) zapLogger() *zap.Logger {
	var level zapcore.Level
	if c.Verbose {
		level = zapcore.DebugLevel
	} else {
		level = zapcore.InfoLevel
	}

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Encoding:         "console",
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			MessageKey:     "msg",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
	}

	logger, _ := cfg.Build()
	return logger
}
