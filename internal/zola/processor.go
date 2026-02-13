package zola

import (
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const maxTitleLength = 50

var (
	codeBlockWithLangRegex = regexp.MustCompile(`<pre><code class="language-(.*?)">([\s\S]*?)</code></pre>`)
	codeBlockNoLangRegex   = regexp.MustCompile(`<pre><code>([\s\S]*?)</code></pre>`)
	newlineRunRegex        = regexp.MustCompile(`\n+`)
	addressRegex           = regexp.MustCompile(`(\s\s\n)?0x[0-9a-fA-F]+\n?$`)
	htmlTagRegex           = regexp.MustCompile(`<[^>]*>`)
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
		markdownCodeBlock := "```\n" + codeContent + "\n```"
		placeholder := "___CODEBLOCK_" + strconv.Itoa(codeIndex) + "___"
		codeBlockPlaceholders[placeholder] = markdownCodeBlock
		codeIndex++
		return placeholder
	})

	// Convert newlines to Zola-compatible format:
	// - Single \n → trailing two spaces + \n (markdown hard line break)
	// - Double \n\n → paragraph break
	// - Triple+ \n{3,} → paragraph break + <br> prepended to next paragraph
	//   (CommonMark collapses multiple blank lines into one paragraph break;
	//   <br> tags at the END of a <p> lose one visible line because the
	//   paragraph's bottom margin overlaps the last <br>, so we place them
	//   at the START of the next paragraph where they reliably render)
	content = newlineRunRegex.ReplaceAllStringFunc(content, func(match string) string {
		n := len(match)
		if n == 1 {
			return "  \n"
		}
		if n == 2 {
			return "  \n  \n"
		}
		return "  \n  \n" + strings.Repeat("<br>", n-1)
	})

	for placeholder, codeBlock := range codeBlockPlaceholders {
		content = strings.ReplaceAll(content, placeholder, codeBlock)
	}

	return content
}

// ExtractTitle derives a post title from content.
// If a hex address (0x...) is found at the end, returns "channelID [address]".
// Otherwise uses the first line of content (stripped of HTML), truncated if needed.
// Falls back to channelID for empty/whitespace-only content.
func ExtractTitle(content string, channelID string) string {
	if content == "" {
		return channelID
	}

	match := addressRegex.FindString(content)
	if match != "" {
		address := strings.TrimSpace(match)
		return channelID + " [" + address + "]"
	}

	firstLine := strings.SplitN(content, "\n", 2)[0]
	firstLine = htmlTagRegex.ReplaceAllString(firstLine, "")
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

// RemoveAddressPattern removes the address regex pattern from content
// and trims any trailing whitespace and <br> tags left behind.
func RemoveAddressPattern(content string) string {
	content = addressRegex.ReplaceAllString(content, "")

	// Trim trailing <br> tags, spaces, and newlines that are no longer
	// needed once the address (the next paragraph) has been removed.
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
