package zola

import "testing"

func TestGetImageFormat(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"jpeg", []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}, "jpg"},
		{"png", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D}, "png"},
		{"gif", []byte{0x47, 0x49, 0x46, 0x38, 0x39}, "gif"},
		{"webp", []byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50}, "webp"},
		{"unknown", []byte{0x00, 0x00, 0x00, 0x00, 0x00}, "jpg"},
		{"short", []byte{0xFF, 0xD8}, "jpg"},
		{"empty", []byte{}, "jpg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getImageFormat(tt.data); got != tt.want {
				t.Errorf("getImageFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}
