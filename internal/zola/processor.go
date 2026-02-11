package zola

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Package-level compiled regexps — avoids recompilation on every function call.
var (
	codeBlockWithLangRegex = regexp.MustCompile(`<pre><code class="language-(.*?)">([\s\S]*?)</code></pre>`)
	codeBlockNoLangRegex   = regexp.MustCompile(`<pre><code>([\s\S]*?)</code></pre>`)
	addressRegex           = regexp.MustCompile(`(?m)(\s\s\n)?0x[0-9a-fA-F]+\n?$`)
)

// ProcessContent converts HTML content (from EntitiesToHTML) to Zola-compatible format.
// Converts code blocks to markdown and handles line breaks.
func ProcessContent(content string) string {
	if content == "" {
		return ""
	}

	// Handle code blocks WITH language → convert to markdown
	codeBlockPlaceholders := make(map[string]string)
	codeIndex := 0
	content = codeBlockWithLangRegex.ReplaceAllStringFunc(content, func(match string) string {
		matches := codeBlockWithLangRegex.FindStringSubmatch(match)
		language := matches[1]
		codeContent := strings.TrimRight(matches[2], "\n")
		markdownCodeBlock := "```" + language + "\n" + codeContent + "\n```"
		placeholder := "___CODEBLOCK_" + strconv.Itoa(codeIndex) + "___"
		codeBlockPlaceholders[placeholder] = markdownCodeBlock
		codeIndex++
		return placeholder
	})

	// Handle code blocks WITHOUT language → preserve as placeholder to protect from newline conversion
	content = codeBlockNoLangRegex.ReplaceAllStringFunc(content, func(match string) string {
		matches := codeBlockNoLangRegex.FindStringSubmatch(match)
		codeContent := strings.TrimRight(matches[1], "\n")
		// Convert to markdown code block without language
		markdownCodeBlock := "```\n" + codeContent + "\n```"
		placeholder := "___CODEBLOCK_" + strconv.Itoa(codeIndex) + "___"
		codeBlockPlaceholders[placeholder] = markdownCodeBlock
		codeIndex++
		return placeholder
	})

	content = strings.ReplaceAll(content, "\n", "  \n")
	content = strings.ReplaceAll(content, "  \n  \n", "  \n\n")

	for placeholder, codeBlock := range codeBlockPlaceholders {
		content = strings.ReplaceAll(content, placeholder, codeBlock)
	}

	return content
}

// ExtractTitle looks for an address regex pattern (0x...) in content.
// Returns "channelID [address]" if found, otherwise returns channelID.
func ExtractTitle(content string, channelID string) string {
	if content == "" {
		return channelID
	}

	match := addressRegex.FindString(content)
	if match != "" {
		address := strings.TrimSpace(match)
		return channelID + " [" + address + "]"
	}

	return channelID
}

// RemoveAddressPattern removes the address regex pattern from content.
func RemoveAddressPattern(content string) string {
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
