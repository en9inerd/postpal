package zola

import (
	"regexp"
	"strings"
	"time"
)

// ProcessContent converts Telegram HTML content to Markdown format.
// HTML entities are NOT decoded (matches TypeScript implementation).
func ProcessContent(content string) string {
	if content == "" {
		return ""
	}

	codeRegex := regexp.MustCompile(`<code>([\s\S]*?)</code>`)
	content = codeRegex.ReplaceAllStringFunc(content, func(match string) string {
		codeContent := codeRegex.FindStringSubmatch(match)[1]
		escapedCodeContent := strings.ReplaceAll(codeContent, "<", "&lt;")
		escapedCodeContent = strings.ReplaceAll(escapedCodeContent, ">", "&gt;")
		return "<code>" + escapedCodeContent + "</code>"
	})

	codeBlockRegex := regexp.MustCompile(`<pre><code class="language-(.*?)">([\s\S]*?)</code></pre>`)
	codeBlockPlaceholders := make(map[string]string)
	codeIndex := 0
	content = codeBlockRegex.ReplaceAllStringFunc(content, func(match string) string {
		matches := codeBlockRegex.FindStringSubmatch(match)
		language := matches[1]
		codeContent := strings.TrimRight(matches[2], "\n")
		markdownCodeBlock := "```" + language + "\n" + codeContent + "\n```"
		placeholder := "___CODEBLOCK_" + string(rune('A'+codeIndex)) + "___"
		codeBlockPlaceholders[placeholder] = markdownCodeBlock
		codeIndex++
		return placeholder
	})

	blockquoteRegex := regexp.MustCompile(`<blockquote>([\s\S]*?)</blockquote>`)
	content = blockquoteRegex.ReplaceAllStringFunc(content, func(match string) string {
		blockquoteContent := blockquoteRegex.FindStringSubmatch(match)[1]
		blockquoteContent = strings.ReplaceAll(blockquoteContent, "\n", "<br>")
		return "<blockquote>" + blockquoteContent + "</blockquote>"
	})

	content = strings.ReplaceAll(content, "\n", "  \n")
	content = strings.ReplaceAll(content, "  \n  \n", "  \n\n")

	for placeholder, codeBlock := range codeBlockPlaceholders {
		content = strings.ReplaceAll(content, placeholder, codeBlock)
	}

	spoilerRegex := regexp.MustCompile(`<spoiler>([\s\S]*?)</spoiler>`)
	content = spoilerRegex.ReplaceAllString(content, `<span class="spoiler">$1</span>`)

	return content
}

// ExtractTitle looks for an address regex pattern (0x...) in content.
// Returns "channelID [address]" if found, otherwise returns channelID.
func ExtractTitle(content string, channelID string) string {
	if content == "" {
		return channelID
	}

	addressRegex := regexp.MustCompile(`(?m)(\s\s\n)?0x[0-9a-fA-F]+\n?$`)
	match := addressRegex.FindString(content)
	if match != "" {
		address := strings.TrimSpace(match)
		return channelID + " [" + address + "]"
	}

	return channelID
}

// RemoveAddressPattern removes the address regex pattern from content.
func RemoveAddressPattern(content string) string {
	addressRegex := regexp.MustCompile(`(?m)(\s\s\n)?0x[0-9a-fA-F]+\n?$`)
	return addressRegex.ReplaceAllString(content, "")
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
	sb.WriteString("\n\n")

	if len(post.ImageNames) > 0 {
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
