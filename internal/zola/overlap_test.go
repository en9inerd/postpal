package zola

import (
	"testing"

	"github.com/gotd/td/tg"
)

func TestEntitiesToHTML_Overlapping(t *testing.T) {
	// Test: "Hello World" with bold on "Hello Wo" and italic on "lo World"
	text := "Hello World"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 0, Length: 8},   // "Hello Wo"
		&tg.MessageEntityItalic{Offset: 3, Length: 8}, // "lo World"
	}

	result := EntitiesToHTML(text, entities)

	// Should produce properly nested tags
	// "Hel" = bold only, "lo Wo" = bold+italic (nested), "rld" = italic only
	// Valid output: <strong>Hel<em>lo Wo</em></strong><em>rld</em>
	expected := "<strong>Hel<em>lo Wo</em></strong><em>rld</em>"
	if result != expected {
		t.Errorf("Overlapping entities failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_Nested(t *testing.T) {
	// Test: Nested - bold wrapping italic
	text := "all bold some italic"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 0, Length: 20},  // entire text
		&tg.MessageEntityItalic{Offset: 9, Length: 4}, // "some"
	}

	result := EntitiesToHTML(text, entities)

	// Should produce: <strong>all bold <em>some</em> italic</strong>
	expected := "<strong>all bold <em>some</em> italic</strong>"
	if result != expected {
		t.Errorf("Nested entities failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_SamePosition(t *testing.T) {
	// Test: Bold and italic starting at same position
	text := "styled text"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 0, Length: 6},   // "styled"
		&tg.MessageEntityItalic{Offset: 0, Length: 6}, // "styled"
	}

	result := EntitiesToHTML(text, entities)

	// Both should be applied
	// Could be <strong><em>styled</em></strong> or <em><strong>styled</strong></em>
	// Either is valid, just check both tags are present
	if result != "<strong><em>styled</em></strong> text" && result != "<em><strong>styled</strong></em> text" {
		t.Errorf("Same position entities failed\nGot: %s", result)
	}
}
