package zola

import (
	"strings"
	"testing"
)

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
