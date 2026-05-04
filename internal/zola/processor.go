package zola

import (
	"html"
	"strconv"
	"strings"
	"time"
)

const maxTitleLength = 50

const (
	preCodeOpen  = "<pre><code"
	preCodeClose = "</code></pre>"
	langAttr     = `class="language-`
)

// ProcessContent converts HTML content (from EntitiesToHTML) to Zola-compatible format.
// Converts code blocks to markdown and handles line breaks.
func ProcessContent(content string) string {
	if content == "" {
		return ""
	}

	content, placeholders := replaceCodeBlocks(content)
	content = convertNewlines(content)

	for placeholder, block := range placeholders {
		content = strings.ReplaceAll(content, placeholder, block)
	}

	return content
}

// replaceCodeBlocks replaces <pre><code> HTML blocks with markdown code block
// placeholders. Handles both language-tagged and plain code blocks in a single pass.
func replaceCodeBlocks(content string) (string, map[string]string) {
	placeholders := make(map[string]string)
	var b strings.Builder
	b.Grow(len(content))
	idx := 0
	pos := 0

	for {
		start := strings.Index(content[pos:], preCodeOpen)
		if start == -1 {
			b.WriteString(content[pos:])
			break
		}
		b.WriteString(content[pos : pos+start])

		searchFrom := pos + start + len(preCodeOpen)
		tagEnd := strings.IndexByte(content[searchFrom:], '>')
		if tagEnd == -1 {
			b.WriteString(content[pos:])
			break
		}

		openTag := content[pos+start : searchFrom+tagEnd+1]

		var language string
		if li := strings.Index(openTag, langAttr); li >= 0 {
			ls := li + len(langAttr)
			if le := strings.IndexByte(openTag[ls:], '"'); le >= 0 {
				language = openTag[ls : ls+le]
			}
		}

		afterOpen := searchFrom + tagEnd + 1
		closeIdx := strings.Index(content[afterOpen:], preCodeClose)
		if closeIdx == -1 {
			b.WriteString(content[pos:])
			break
		}

		code := strings.TrimRight(content[afterOpen:afterOpen+closeIdx], "\n")

		var block string
		if language != "" {
			block = "```" + language + "\n" + code + "\n```"
		} else {
			block = "```\n" + code + "\n```"
		}

		placeholder := "___CODEBLOCK_" + strconv.Itoa(idx) + "___"
		placeholders[placeholder] = block
		b.WriteString(placeholder)
		idx++

		pos = afterOpen + closeIdx + len(preCodeClose)
	}

	return b.String(), placeholders
}

// convertNewlines converts newline runs to Zola-compatible format:
//   - Single \n -> trailing two spaces + \n (markdown hard line break)
//   - Double \n\n -> paragraph break
//   - Triple+ \n{3,} -> paragraph break + <br> prepended to next paragraph
//     (CommonMark collapses multiple blank lines into one paragraph break;
//     <br> tags at the END of a <p> lose one visible line because the
//     paragraph's bottom margin overlaps the last <br>, so we place them
//     at the START of the next paragraph where they reliably render)
func convertNewlines(content string) string {
	var b strings.Builder
	b.Grow(len(content) * 2)
	i := 0
	for i < len(content) {
		if content[i] != '\n' {
			b.WriteByte(content[i])
			i++
			continue
		}
		n := 0
		for i < len(content) && content[i] == '\n' {
			n++
			i++
		}
		switch {
		case n == 1:
			b.WriteString("  \n")
		case n == 2:
			b.WriteString("  \n  \n")
		default:
			b.WriteString("  \n  \n")
			b.WriteString(strings.Repeat("<br>", n-1))
		}
	}
	return b.String()
}

// ExtractTitle derives a post title from content.
// If a hex address (0x...) is found at the end, returns "channelID [address]".
// Otherwise uses the first line of content (stripped of HTML), truncated if needed.
// Falls back to channelID for empty/whitespace-only content.
func ExtractTitle(content string, channelID string) string {
	if content == "" {
		return channelID
	}

	if addr := findTrailingAddress(content); addr != "" {
		return channelID + " [" + addr + "]"
	}

	firstLine := strings.SplitN(content, "\n", 2)[0]
	firstLine = stripHTMLTags(firstLine)
	firstLine = html.UnescapeString(firstLine)
	firstLine = strings.TrimSpace(firstLine)

	if firstLine == "" {
		return channelID
	}

	runes := []rune(firstLine)
	if len(runes) <= maxTitleLength {
		return firstLine
	}

	truncated := string(runes[:maxTitleLength])
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > 0 {
		truncated = truncated[:lastSpace]
	}
	return truncated + "..."
}

// RemoveAddressPattern removes the trailing address pattern from content
// and trims any trailing whitespace and <br> tags left behind.
func RemoveAddressPattern(content string) string {
	content = removeTrailingAddress(content)

	for {
		trimmed := strings.TrimRight(content, " \t\n")
		trimmed = strings.TrimSuffix(trimmed, "<br>")
		if trimmed == content {
			break
		}
		content = trimmed
	}
	return content
}

// findTrailingAddress returns the hex address (e.g. "0x001A") at the end of
// content, or an empty string if none is found.
func findTrailingAddress(s string) string {
	end := len(s)
	if end > 0 && s[end-1] == '\n' {
		end--
	}

	hexEnd := end
	for hexEnd > 0 && isHexDigit(s[hexEnd-1]) {
		hexEnd--
	}

	if hexEnd >= end || hexEnd < 2 || s[hexEnd-2:hexEnd] != "0x" {
		return ""
	}

	return s[hexEnd-2 : end]
}

// removeTrailingAddress removes the trailing hex address pattern from content,
// including an optional two-whitespace + newline prefix before the address.
func removeTrailingAddress(s string) string {
	end := len(s)
	if end > 0 && s[end-1] == '\n' {
		end--
	}

	hexEnd := end
	for hexEnd > 0 && isHexDigit(s[hexEnd-1]) {
		hexEnd--
	}

	if hexEnd >= end || hexEnd < 2 || s[hexEnd-2:hexEnd] != "0x" {
		return s
	}

	cut := hexEnd - 2

	if cut >= 3 && s[cut-1] == '\n' && isWhitespace(s[cut-3]) && isWhitespace(s[cut-2]) {
		cut -= 3
	}

	return s[:cut]
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\f' || b == '\r'
}

// stripHTMLTags removes HTML tags from a string, returning plain text.
func stripHTMLTags(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	depth := 0
	for _, r := range s {
		switch {
		case r == '<':
			depth++
		case r == '>':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// Post represents a Zola blog post
type Post struct {
	ID         int64
	Title      string
	Content    string
	Date       time.Time
	ImageNames []string
}

// BuildFrontMatter generates TOML front matter for a Zola post.
func BuildFrontMatter(post Post) string {
	var sb strings.Builder
	sb.WriteString("+++\n")
	sb.WriteString("title = \"")
	sb.WriteString(strings.ReplaceAll(post.Title, "\"", "\\\""))
	sb.WriteString("\"\n")
	sb.WriteString("date = ")
	sb.WriteString(post.Date.Format(time.RFC3339))
	sb.WriteString("\n")

	if len(post.ImageNames) > 0 {
		sb.WriteString("\n")
		sb.WriteString("[extra]\n")
		sb.WriteString("images = [")
		for i, imgName := range post.ImageNames {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("\"")
			sb.WriteString(imgName)
			sb.WriteString("\"")
		}
		sb.WriteString("]\n")
	}

	sb.WriteString("+++\n\n")
	return sb.String()
}
