package zola

import (
	"fmt"
	"os"
	"path/filepath"
)

// SaveChannelLogo saves the channel logo to the static directory
func (s *Service) SaveChannelLogo(logoData []byte) error {
	format := getImageFormat(logoData)
	logoPath := filepath.Join(s.repoDir, "static", "logo."+format)

	staticDir := filepath.Join(s.repoDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("failed to create static directory: %w", err)
	}

	if err := os.WriteFile(logoPath, logoData, 0644); err != nil {
		return fmt.Errorf("failed to write logo file: %w", err)
	}

	relLogoPath := filepath.Join("static", "logo."+format)
	if err := s.gitService.Add(relLogoPath); err != nil {
		return fmt.Errorf("failed to add logo to git: %w", err)
	}

	return nil
}

func getImageFormat(data []byte) string {
	if len(data) < 4 {
		return "jpg"
	}

	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpg"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "gif"
	}
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 && string(data[8:12]) == "WEBP" {
		return "webp"
	}

	return "jpg"
}
