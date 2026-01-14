package tgbot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

// BotInfo holds bot profile information.
type BotInfo struct {
	// Name is the bot's display name.
	Name string

	// About is the short description shown in the bot's profile.
	About string

	// Description is the longer description shown when starting the bot.
	Description string

	// LangCode is the language code for this info.
	// Empty string means default language.
	LangCode string
}

// UpdateBotInfo updates the bot's profile information.
// Only non-empty fields are updated.
func (b *Bot) UpdateBotInfo(ctx context.Context, info BotInfo) error {
	if b.api == nil {
		return ErrBotNotRunning
	}

	currentInfo, err := b.api.BotsGetBotInfo(ctx, &tg.BotsGetBotInfoRequest{
		LangCode: info.LangCode,
	})
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}

	name := info.Name
	if name == "" {
		name = currentInfo.Name
	}
	about := info.About
	if about == "" {
		about = currentInfo.About
	}
	description := info.Description
	if description == "" {
		description = currentInfo.Description
	}

	if name == currentInfo.Name && about == currentInfo.About && description == currentInfo.Description {
		b.config.Logger.Debug("bot info unchanged, skipping update")
		return nil
	}

	_, err = b.api.BotsSetBotInfo(ctx, &tg.BotsSetBotInfoRequest{
		Name:        name,
		About:       about,
		Description: description,
		LangCode:    info.LangCode,
	})
	if err != nil {
		return fmt.Errorf("failed to set bot info: %w", err)
	}

	b.config.Logger.Info("updated bot info",
		"name", name,
		"about", about,
		"description", description,
		"lang", info.LangCode)
	return nil
}

// SetProfilePhoto sets the bot's profile photo from a URL.
func (b *Bot) SetProfilePhoto(ctx context.Context, photoURL string) error {
	if b.api == nil {
		return ErrBotNotRunning
	}

	resp, err := http.Get(photoURL)
	if err != nil {
		return fmt.Errorf("failed to download photo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download photo: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read photo data: %w", err)
	}

	filename := path.Base(photoURL)
	if filename == "" || filename == "." {
		contentType := resp.Header.Get("Content-Type")
		switch contentType {
		case "image/jpeg":
			filename = "photo.jpg"
		case "image/png":
			filename = "photo.png"
		default:
			filename = "photo.jpg"
		}
	}

	return b.SetProfilePhotoFromBytes(ctx, data, filename)
}

// SetProfilePhotoFromBytes sets the bot's profile photo from raw bytes.
func (b *Bot) SetProfilePhotoFromBytes(ctx context.Context, data []byte, filename string) error {
	if b.api == nil {
		return ErrBotNotRunning
	}

	u := uploader.NewUploader(b.api)
	file, err := u.FromBytes(ctx, filename, data)
	if err != nil {
		return fmt.Errorf("failed to upload photo: %w", err)
	}

	_, err = b.api.PhotosUploadProfilePhoto(ctx, &tg.PhotosUploadProfilePhotoRequest{
		File: file,
	})
	if err != nil {
		return fmt.Errorf("failed to set profile photo: %w", err)
	}

	b.config.Logger.Info("updated bot profile photo")
	return nil
}

// DeleteProfilePhotos deletes the bot's profile photos.
func (b *Bot) DeleteProfilePhotos(ctx context.Context) error {
	if b.api == nil {
		return ErrBotNotRunning
	}

	photos, err := b.api.PhotosGetUserPhotos(ctx, &tg.PhotosGetUserPhotosRequest{
		UserID: &tg.InputUserSelf{},
		Limit:  100,
	})
	if err != nil {
		return fmt.Errorf("failed to get photos: %w", err)
	}

	var inputPhotos []tg.InputPhotoClass
	switch p := photos.(type) {
	case *tg.PhotosPhotos:
		for _, photo := range p.Photos {
			if ph, ok := photo.(*tg.Photo); ok {
				inputPhotos = append(inputPhotos, &tg.InputPhoto{
					ID:            ph.ID,
					AccessHash:    ph.AccessHash,
					FileReference: ph.FileReference,
				})
			}
		}
	case *tg.PhotosPhotosSlice:
		for _, photo := range p.Photos {
			if ph, ok := photo.(*tg.Photo); ok {
				inputPhotos = append(inputPhotos, &tg.InputPhoto{
					ID:            ph.ID,
					AccessHash:    ph.AccessHash,
					FileReference: ph.FileReference,
				})
			}
		}
	}

	if len(inputPhotos) == 0 {
		return nil
	}

	_, err = b.api.PhotosDeletePhotos(ctx, inputPhotos)
	if err != nil {
		return fmt.Errorf("failed to delete photos: %w", err)
	}

	b.config.Logger.Info("deleted bot profile photos", "count", len(inputPhotos))
	return nil
}
