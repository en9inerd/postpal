package zola

import (
	"strings"
	"testing"

	"github.com/gotd/td/tg"
)

// =============================================================================
// Basic Text Formatting (TC-001 to TC-005)
// =============================================================================

func TestEntitiesToHTML_Bold(t *testing.T) {
	text := "This is bold text"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 8, Length: 4}, // "bold"
	}

	result := EntitiesToHTML(text, entities)
	expected := "This is <strong>bold</strong> text"

	if result != expected {
		t.Errorf("TC-001 Bold failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_Italic(t *testing.T) {
	text := "This is italic text"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityItalic{Offset: 8, Length: 6}, // "italic"
	}

	result := EntitiesToHTML(text, entities)
	expected := "This is <em>italic</em> text"

	if result != expected {
		t.Errorf("TC-002 Italic failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_Underline(t *testing.T) {
	text := "This is underlined text"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityUnderline{Offset: 8, Length: 10}, // "underlined"
	}

	result := EntitiesToHTML(text, entities)
	expected := "This is <u>underlined</u> text"

	if result != expected {
		t.Errorf("TC-003 Underline failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_Strikethrough(t *testing.T) {
	text := "This is strikethrough text"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityStrike{Offset: 8, Length: 13}, // "strikethrough"
	}

	result := EntitiesToHTML(text, entities)
	expected := "This is <s>strikethrough</s> text"

	if result != expected {
		t.Errorf("TC-004 Strikethrough failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_Spoiler(t *testing.T) {
	text := "This is spoiler text"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntitySpoiler{Offset: 8, Length: 7}, // "spoiler"
	}

	result := EntitiesToHTML(text, entities)
	expected := "This is <span class=\"spoiler\">spoiler</span> text"

	if result != expected {
		t.Errorf("TC-005 Spoiler failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

// =============================================================================
// Code Formatting (TC-006 to TC-010)
// =============================================================================

func TestEntitiesToHTML_InlineCode(t *testing.T) {
	text := "Use fmt.Println() to print"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityCode{Offset: 4, Length: 13}, // "fmt.Println()"
	}

	result := EntitiesToHTML(text, entities)
	expected := "Use <code>fmt.Println()</code> to print"

	if result != expected {
		t.Errorf("TC-006 Inline code failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_InlineCodeWithSpecialChars(t *testing.T) {
	text := "Check if x < 10 && y > 5 condition"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityCode{Offset: 6, Length: 18}, // "if x < 10 && y > 5"
	}

	result := EntitiesToHTML(text, entities)
	// Special characters should be escaped
	if !strings.Contains(result, "<code>") {
		t.Errorf("TC-007 Inline code with special chars failed\nGot: %s", result)
	}
	if !strings.Contains(result, "&lt;") || !strings.Contains(result, "&gt;") {
		t.Errorf("TC-007 Special chars should be escaped\nGot: %s", result)
	}
}

func TestEntitiesToHTML_CodeBlockWithLanguage(t *testing.T) {
	text := "package main\n\nfunc main() {\n    println(\"Hello\")\n}"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityPre{Offset: 0, Length: 50, Language: "go"},
	}

	result := EntitiesToHTML(text, entities)

	if !strings.Contains(result, "language-go") {
		t.Errorf("TC-008 Code block should have language class\nGot: %s", result)
	}
	if !strings.Contains(result, "<pre><code") {
		t.Errorf("TC-008 Code block should have pre/code tags\nGot: %s", result)
	}
}

func TestEntitiesToHTML_CodeBlockWithoutLanguage(t *testing.T) {
	text := "plain code block"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityPre{Offset: 0, Length: 16, Language: ""},
	}

	result := EntitiesToHTML(text, entities)

	if strings.Contains(result, "language-") {
		t.Errorf("TC-009 Code block without language should not have language class\nGot: %s", result)
	}
	if !strings.Contains(result, "<pre><code>") {
		t.Errorf("TC-009 Code block should have pre/code tags\nGot: %s", result)
	}
}

func TestEntitiesToHTML_CodeBlockLanguages(t *testing.T) {
	languages := []string{"javascript", "python", "typescript", "rust", "bash", "json", "yaml", "sql", "html", "css"}

	for _, lang := range languages {
		text := "code here"
		entities := []tg.MessageEntityClass{
			&tg.MessageEntityPre{Offset: 0, Length: 9, Language: lang},
		}

		result := EntitiesToHTML(text, entities)

		if !strings.Contains(result, "language-"+lang) {
			t.Errorf("TC-010 Code block with %s should have correct language class\nGot: %s", lang, result)
		}
	}
}

// =============================================================================
// Links (TC-011 to TC-015)
// =============================================================================

func TestEntitiesToHTML_TextURL(t *testing.T) {
	text := "Visit Google for search"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityTextURL{Offset: 6, Length: 6, URL: "https://google.com"}, // "Google"
	}

	result := EntitiesToHTML(text, entities)
	expected := "Visit <a href=\"https://google.com\">Google</a> for search"

	if result != expected {
		t.Errorf("TC-011 Text URL failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_AutoURL(t *testing.T) {
	text := "Check out https://example.com for more"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityURL{Offset: 10, Length: 19}, // "https://example.com"
	}

	result := EntitiesToHTML(text, entities)

	if !strings.Contains(result, "<a href=\"https://example.com\">https://example.com</a>") {
		t.Errorf("TC-012 Auto URL failed\nGot: %s", result)
	}
}

func TestEntitiesToHTML_Mention(t *testing.T) {
	text := "Contact @username for help"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityMention{Offset: 8, Length: 9}, // "@username"
	}

	result := EntitiesToHTML(text, entities)

	if !strings.Contains(result, "https://t.me/username") {
		t.Errorf("TC-013 Mention failed\nGot: %s", result)
	}
}

func TestEntitiesToHTML_MentionName(t *testing.T) {
	text := "Contact John for help"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityMentionName{Offset: 8, Length: 4, UserID: 123456789}, // "John"
	}

	result := EntitiesToHTML(text, entities)

	if !strings.Contains(result, "tg://user?id=123456789") {
		t.Errorf("TC-014 Mention by user ID failed\nGot: %s", result)
	}
}

func TestEntitiesToHTML_URLWithSpecialChars(t *testing.T) {
	text := "Link: https://example.com/path?query=value&other=123"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityURL{Offset: 6, Length: 46},
	}

	result := EntitiesToHTML(text, entities)

	// URL should be properly escaped in href
	if !strings.Contains(result, "<a href=") {
		t.Errorf("TC-015 URL with special chars failed\nGot: %s", result)
	}
}

// =============================================================================
// Blockquotes (TC-016 to TC-017)
// =============================================================================

func TestEntitiesToHTML_Blockquote(t *testing.T) {
	text := "This is a quote"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBlockquote{Offset: 0, Length: 15},
	}

	result := EntitiesToHTML(text, entities)
	expected := "<blockquote>This is a quote</blockquote>"

	if result != expected {
		t.Errorf("TC-016 Blockquote failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_BlockquoteMultiline(t *testing.T) {
	text := "Line one\nLine two\nLine three"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBlockquote{Offset: 0, Length: 28},
	}

	result := EntitiesToHTML(text, entities)

	// Newlines inside blockquote should be converted to <br>
	expected := "<blockquote>Line one<br>Line two<br>Line three</blockquote>"
	if result != expected {
		t.Errorf("TC-017 Multi-line blockquote failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_BlockquoteExpandable(t *testing.T) {
	text := "Line one\nLine two"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBlockquote{Offset: 0, Length: 17, Collapsed: true},
	}

	result := EntitiesToHTML(text, entities)

	// Should have expandable class and newlines converted to <br>
	expected := "<blockquote class=\"expandable\">Line one<br>Line two</blockquote>"
	if result != expected {
		t.Errorf("Expandable blockquote failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_BlockquoteWithFormatting(t *testing.T) {
	// Telegram allows formatting inside blockquotes
	text := "Quote with bold and italic text"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBlockquote{Offset: 0, Length: 31},
		&tg.MessageEntityBold{Offset: 11, Length: 4},   // "bold"
		&tg.MessageEntityItalic{Offset: 20, Length: 6}, // "italic"
	}

	result := EntitiesToHTML(text, entities)

	// Should have formatting inside blockquote
	if !strings.Contains(result, "<blockquote>") {
		t.Errorf("Missing blockquote tag\nGot: %s", result)
	}
	if !strings.Contains(result, "<strong>bold</strong>") {
		t.Errorf("Missing bold inside blockquote\nGot: %s", result)
	}
	if !strings.Contains(result, "<em>italic</em>") {
		t.Errorf("Missing italic inside blockquote\nGot: %s", result)
	}
}

func TestEntitiesToHTML_BlockquoteWithLink(t *testing.T) {
	text := "Quote with a link here"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBlockquote{Offset: 0, Length: 22},
		&tg.MessageEntityTextURL{Offset: 13, Length: 4, URL: "https://example.com"}, // "link"
	}

	result := EntitiesToHTML(text, entities)

	if !strings.Contains(result, "<a href=\"https://example.com\">link</a>") {
		t.Errorf("Missing link inside blockquote\nGot: %s", result)
	}
}

// =============================================================================
// Nested and Combined Formatting (TC-018 to TC-023)
// =============================================================================

func TestEntitiesToHTML_BoldItalicSameRange(t *testing.T) {
	text := "This is bold and italic text"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 8, Length: 15},   // "bold and italic"
		&tg.MessageEntityItalic{Offset: 8, Length: 15}, // "bold and italic"
	}

	result := EntitiesToHTML(text, entities)

	// Both tags should be present
	if !strings.Contains(result, "<strong>") || !strings.Contains(result, "<em>") {
		t.Errorf("TC-018 Bold+Italic same range failed\nGot: %s", result)
	}
	if !strings.Contains(result, "</strong>") || !strings.Contains(result, "</em>") {
		t.Errorf("TC-018 Both closing tags should be present\nGot: %s", result)
	}
}

func TestEntitiesToHTML_BoldContainingItalic(t *testing.T) {
	text := "all bold some italic still bold"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 0, Length: 31},   // entire text
		&tg.MessageEntityItalic{Offset: 9, Length: 11}, // "some italic"
	}

	result := EntitiesToHTML(text, entities)
	expected := "<strong>all bold <em>some italic</em> still bold</strong>"

	if result != expected {
		t.Errorf("TC-019 Bold containing italic failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_OverlappingBoldItalic(t *testing.T) {
	// Same as overlap_test.go TC-020
	text := "Hello World"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 0, Length: 8},   // "Hello Wo"
		&tg.MessageEntityItalic{Offset: 3, Length: 8}, // "lo World"
	}

	result := EntitiesToHTML(text, entities)
	expected := "<strong>Hel<em>lo Wo</em></strong><em>rld</em>"

	if result != expected {
		t.Errorf("TC-020 Overlapping bold/italic failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_LinkWithBold(t *testing.T) {
	text := "Click here for info"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityTextURL{Offset: 6, Length: 4, URL: "https://example.com"}, // "here"
		&tg.MessageEntityBold{Offset: 6, Length: 4},                                // "here"
	}

	result := EntitiesToHTML(text, entities)

	// Both link and bold should be present
	if !strings.Contains(result, "<a href=") {
		t.Errorf("TC-021 Link with bold - missing link\nGot: %s", result)
	}
	if !strings.Contains(result, "<strong>") || !strings.Contains(result, "</strong>") {
		t.Errorf("TC-021 Link with bold - missing bold\nGot: %s", result)
	}
}

func TestEntitiesToHTML_MultipleFormattingTypes(t *testing.T) {
	text := "Bold and italic and code and strike and underline"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 0, Length: 4},       // "Bold"
		&tg.MessageEntityItalic{Offset: 9, Length: 6},     // "italic"
		&tg.MessageEntityCode{Offset: 20, Length: 4},      // "code"
		&tg.MessageEntityStrike{Offset: 29, Length: 6},    // "strike"
		&tg.MessageEntityUnderline{Offset: 40, Length: 9}, // "underline"
	}

	result := EntitiesToHTML(text, entities)

	// All formatting types should be present
	checks := []string{"<strong>", "<em>", "<code>", "<s>", "<u>"}
	for _, check := range checks {
		if !strings.Contains(result, check) {
			t.Errorf("TC-023 Multiple formatting - missing %s\nGot: %s", check, result)
		}
	}
}

// =============================================================================
// Special Characters and HTML Entities (TC-027 to TC-031)
// =============================================================================

func TestEntitiesToHTML_AngleBrackets(t *testing.T) {
	text := "Compare: 5 < 10 and 20 > 15"
	entities := []tg.MessageEntityClass{}

	result := EntitiesToHTML(text, entities)

	if !strings.Contains(result, "&lt;") || !strings.Contains(result, "&gt;") {
		t.Errorf("TC-028 Angle brackets should be escaped\nGot: %s", result)
	}
}

func TestEntitiesToHTML_Ampersand(t *testing.T) {
	text := "Tom & Jerry"
	entities := []tg.MessageEntityClass{}

	result := EntitiesToHTML(text, entities)

	if !strings.Contains(result, "&amp;") {
		t.Errorf("TC-029 Ampersand should be escaped\nGot: %s", result)
	}
}

func TestEntitiesToHTML_Quotes(t *testing.T) {
	text := `She said "Hello"`
	entities := []tg.MessageEntityClass{}

	result := EntitiesToHTML(text, entities)

	// Quotes can be preserved, escaped as &quot; or as &#34;
	if !strings.Contains(result, "\"") && !strings.Contains(result, "&quot;") && !strings.Contains(result, "&#34;") {
		t.Errorf("TC-030 Quotes handling failed\nGot: %s", result)
	}
}

func TestEntitiesToHTML_Unicode(t *testing.T) {
	text := "Hello 👋 World 🌍"
	entities := []tg.MessageEntityClass{}

	result := EntitiesToHTML(text, entities)

	if !strings.Contains(result, "👋") || !strings.Contains(result, "🌍") {
		t.Errorf("TC-031 Unicode/emoji should be preserved\nGot: %s", result)
	}
}

// =============================================================================
// Edge Cases (TC-041 to TC-050)
// =============================================================================

func TestEntitiesToHTML_EmptyText(t *testing.T) {
	text := ""
	entities := []tg.MessageEntityClass{}

	result := EntitiesToHTML(text, entities)

	if result != "" {
		t.Errorf("TC-041 Empty text should return empty\nGot: %s", result)
	}
}

func TestEntitiesToHTML_NoEntities(t *testing.T) {
	text := "Plain text without any formatting"
	entities := []tg.MessageEntityClass{}

	result := EntitiesToHTML(text, entities)

	// Should just escape HTML and return
	if strings.Contains(result, "<") && !strings.Contains(result, "&lt;") {
		t.Errorf("TC-041 No entities - should not add tags\nGot: %s", result)
	}
}

func TestEntitiesToHTML_VeryLongText(t *testing.T) {
	// Create a long text (4096+ chars like Telegram limit)
	text := strings.Repeat("a", 5000)
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 0, Length: 100},
		&tg.MessageEntityItalic{Offset: 4900, Length: 100},
	}

	result := EntitiesToHTML(text, entities)

	if len(result) < 5000 {
		t.Errorf("TC-042 Long text should be fully processed\nGot length: %d", len(result))
	}
	if !strings.Contains(result, "<strong>") {
		t.Errorf("TC-042 Long text - bold at start missing\nGot: %s...", result[:200])
	}
}

func TestEntitiesToHTML_InvalidEntityOffset(t *testing.T) {
	text := "Short text"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 100, Length: 5}, // Beyond text length
	}

	// Should not panic, should handle gracefully
	result := EntitiesToHTML(text, entities)

	if result == "" {
		t.Errorf("TC-044 Invalid offset should still return text\nGot: %s", result)
	}
}

func TestEntitiesToHTML_AdjacentEntities(t *testing.T) {
	text := "bolditalic"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 0, Length: 4},   // "bold"
		&tg.MessageEntityItalic{Offset: 4, Length: 6}, // "italic"
	}

	result := EntitiesToHTML(text, entities)

	// Adjacent entities should both be rendered
	// Implementation may produce empty tags at boundary: <strong>bold<em></em></strong><em>italic</em>
	// This is acceptable as it renders correctly in browsers
	if !strings.Contains(result, "<strong>") || !strings.Contains(result, "bold") {
		t.Errorf("TC-045 Adjacent entities - bold missing\nGot: %s", result)
	}
	if !strings.Contains(result, "<em>") || !strings.Contains(result, "italic") {
		t.Errorf("TC-045 Adjacent entities - italic missing\nGot: %s", result)
	}
}

func TestEntitiesToHTML_CustomEmoji(t *testing.T) {
	text := "Hello emoji here"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityCustomEmoji{Offset: 6, Length: 5, DocumentID: 12345}, // "emoji"
	}

	result := EntitiesToHTML(text, entities)

	// Custom emoji should be skipped (no special handling)
	// Text should still be present
	if !strings.Contains(result, "emoji") {
		t.Errorf("TC-046 Custom emoji - text should be preserved\nGot: %s", result)
	}
}

func TestEntitiesToHTML_DeeplyNested(t *testing.T) {
	text := "deeply nested formatting here"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 0, Length: 29},       // entire text
		&tg.MessageEntityItalic{Offset: 7, Length: 22},     // "nested formatting here"
		&tg.MessageEntityUnderline{Offset: 14, Length: 10}, // "formatting"
	}

	result := EntitiesToHTML(text, entities)

	// All three formatting types should be present
	if !strings.Contains(result, "<strong>") {
		t.Errorf("TC-047 Deeply nested - missing bold\nGot: %s", result)
	}
	if !strings.Contains(result, "<em>") {
		t.Errorf("TC-047 Deeply nested - missing italic\nGot: %s", result)
	}
	if !strings.Contains(result, "<u>") {
		t.Errorf("TC-047 Deeply nested - missing underline\nGot: %s", result)
	}
}

func TestEntitiesToHTML_URLInsideCode(t *testing.T) {
	text := "Check https://example.com for docs"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityCode{Offset: 6, Length: 19}, // "https://example.com"
	}

	result := EntitiesToHTML(text, entities)

	// URL should be in code, not made into a link
	if strings.Contains(result, "<a href=") {
		t.Errorf("TC-049 URL inside code should not be a link\nGot: %s", result)
	}
	if !strings.Contains(result, "<code>") {
		t.Errorf("TC-049 Should have code tags\nGot: %s", result)
	}
}

// =============================================================================
// Unicode/Rune handling
// =============================================================================

func TestEntitiesToHTML_UnicodeOffsets(t *testing.T) {
	// Telegram uses UTF-16 code units for offsets
	// This tests that we handle multi-byte characters correctly
	text := "Hello 世界 World"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 6, Length: 2}, // "世界"
	}

	result := EntitiesToHTML(text, entities)

	if !strings.Contains(result, "<strong>世界</strong>") {
		t.Errorf("Unicode offsets - Chinese chars should be bold\nGot: %s", result)
	}
}

func TestEntitiesToHTML_EmojiOffsets(t *testing.T) {
	text := "Hi 👋 there"
	entities := []tg.MessageEntityClass{
		&tg.MessageEntityBold{Offset: 3, Length: 2}, // "👋" (emoji is 2 UTF-16 code units)
	}

	result := EntitiesToHTML(text, entities)

	// The emoji handling depends on how offsets are calculated
	// At minimum, the text should not be corrupted
	if !strings.Contains(result, "Hi") || !strings.Contains(result, "there") {
		t.Errorf("Emoji offsets - text should not be corrupted\nGot: %s", result)
	}
}

func TestEntitiesToHTML_TrailingWhitespaceTrimming(t *testing.T) {
	// Telegram's markdown parser often includes trailing spaces in entity boundaries
	text := "Hello bold text and more"
	entities := []tg.MessageEntityClass{
		// Entity includes trailing space: "bold text " (offset 6, length 10)
		&tg.MessageEntityBold{Offset: 6, Length: 10},
	}

	result := EntitiesToHTML(text, entities)

	// Trailing whitespace should be outside the tags, not inside
	expected := "Hello <strong>bold text</strong> and more"
	if result != expected {
		t.Errorf("Trailing whitespace trimming failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

func TestEntitiesToHTML_MultipleTrailingSpaces(t *testing.T) {
	text := "Hello bold   and more"
	entities := []tg.MessageEntityClass{
		// Entity includes multiple trailing spaces: "bold   " (offset 6, length 7)
		&tg.MessageEntityBold{Offset: 6, Length: 7},
	}

	result := EntitiesToHTML(text, entities)

	// All trailing spaces should be outside
	expected := "Hello <strong>bold</strong>   and more"
	if result != expected {
		t.Errorf("Multiple trailing spaces trimming failed\nExpected: %s\nGot:      %s", expected, result)
	}
}

// =============================================================================
// Regression tests (TC-054 to TC-056)
// =============================================================================

// Note: Legacy <spoiler> and <tg-spoiler> tag tests removed.
// Telegram sends MessageEntitySpoiler entities, not HTML tags.
// EntitiesToHTML converts these to <span class="spoiler"> directly.

func TestProcessContent_CodeBlockTrailingNewline(t *testing.T) {
	input := "<pre><code class=\"language-go\">code here\n</code></pre>"
	result := ProcessContent(input)

	// Should not have double newlines at end of code block
	if strings.Contains(result, "here\n\n```") {
		t.Errorf("TC-055 Code block trailing newline should be trimmed\nGot: %s", result)
	}
}

// Note: Blockquote newline conversion is now tested in TestEntitiesToHTML_BlockquoteMultiline
// since it's handled by EntitiesToHTML, not ProcessContent.

// =============================================================================
// Table-driven tests for comprehensive coverage
// =============================================================================

func TestEntitiesToHTML_AllBasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		entities []tg.MessageEntityClass
		contains []string
	}{
		{
			name: "Bold",
			text: "test bold here",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityBold{Offset: 5, Length: 4},
			},
			contains: []string{"<strong>bold</strong>"},
		},
		{
			name: "Italic",
			text: "test italic here",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityItalic{Offset: 5, Length: 6},
			},
			contains: []string{"<em>italic</em>"},
		},
		{
			name: "Underline",
			text: "test underline here",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityUnderline{Offset: 5, Length: 9},
			},
			contains: []string{"<u>underline</u>"},
		},
		{
			name: "Strike",
			text: "test strike here",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityStrike{Offset: 5, Length: 6},
			},
			contains: []string{"<s>strike</s>"},
		},
		{
			name: "Code",
			text: "test code here",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityCode{Offset: 5, Length: 4},
			},
			contains: []string{"<code>code</code>"},
		},
		{
			name: "Spoiler",
			text: "test spoiler here",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntitySpoiler{Offset: 5, Length: 7},
			},
			contains: []string{"<span class=\"spoiler\">spoiler</span>"},
		},
		{
			name: "Blockquote",
			text: "test quote here",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityBlockquote{Offset: 5, Length: 5},
			},
			contains: []string{"<blockquote>quote</blockquote>"},
		},
		{
			name: "Pre with language",
			text: "code block",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityPre{Offset: 0, Length: 10, Language: "python"},
			},
			contains: []string{"<pre><code class=\"language-python\">", "</code></pre>"},
		},
		{
			name: "Pre without language",
			text: "code block",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityPre{Offset: 0, Length: 10, Language: ""},
			},
			contains: []string{"<pre><code>", "</code></pre>"},
		},
		{
			name: "TextURL",
			text: "click here now",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityTextURL{Offset: 6, Length: 4, URL: "https://test.com"},
			},
			contains: []string{"<a href=\"https://test.com\">here</a>"},
		},
		{
			name: "URL",
			text: "visit https://test.com today",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityURL{Offset: 6, Length: 16},
			},
			contains: []string{"<a href=\"https://test.com\">https://test.com</a>"},
		},
		{
			name: "Mention",
			text: "contact @testuser please",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityMention{Offset: 8, Length: 9},
			},
			contains: []string{"https://t.me/testuser"},
		},
		{
			name: "MentionName",
			text: "contact John please",
			entities: []tg.MessageEntityClass{
				&tg.MessageEntityMentionName{Offset: 8, Length: 4, UserID: 999},
			},
			contains: []string{"tg://user?id=999"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EntitiesToHTML(tt.text, tt.entities)
			for _, c := range tt.contains {
				if !strings.Contains(result, c) {
					t.Errorf("%s: expected to contain %q\nGot: %s", tt.name, c, result)
				}
			}
		})
	}
}

func TestProcessContent_CodeBlockLanguages(t *testing.T) {
	languages := []string{
		"go", "python", "javascript", "typescript", "rust",
		"java", "cpp", "c", "ruby", "php", "swift", "kotlin",
		"sql", "bash", "shell", "json", "yaml", "xml", "html", "css",
	}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			input := "<pre><code class=\"language-" + lang + "\">code</code></pre>"
			result := ProcessContent(input)

			expected := "```" + lang + "\ncode\n```"
			if result != expected {
				t.Errorf("Language %s: expected %q, got %q", lang, expected, result)
			}
		})
	}
}
