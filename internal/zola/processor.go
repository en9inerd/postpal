package zola

import (
	"cmp"
	"html"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/tg"
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

type entityInfo struct {
	offset       int
	length       int
	startTag     string
	endTag       string
	id           int  // unique ID for tracking in stack
	isBlockquote bool // true for blockquote entities (need special newline handling)
	isPre        bool // true for pre blocks (content not escaped, becomes markdown)
}

type tagEvent struct {
	pos      int
	isStart  bool
	entity   *entityInfo
	priority int // for sorting: starts before ends at same position
}

// EntitiesToHTML converts Telegram message entities to HTML.
// Properly handles overlapping/nested entities.
func EntitiesToHTML(text string, entities []tg.MessageEntityClass) string {
	if len(entities) == 0 {
		return html.EscapeString(text)
	}

	runes := []rune(text)

	var infos []*entityInfo
	for i, entity := range entities {
		var offset, length int
		var startTag, endTag string

		switch e := entity.(type) {
		case *tg.MessageEntityBold:
			offset, length = e.Offset, e.Length
			startTag, endTag = "<strong>", "</strong>"
		case *tg.MessageEntityItalic:
			offset, length = e.Offset, e.Length
			startTag, endTag = "<em>", "</em>"
		case *tg.MessageEntityCode:
			offset, length = e.Offset, e.Length
			startTag, endTag = "<code>", "</code>"
		case *tg.MessageEntityPre:
			offset, length = e.Offset, e.Length
			if offset < 0 || offset+length > len(runes) {
				continue
			}
			lang := e.Language
			if lang != "" {
				startTag = "<pre><code class=\"language-" + lang + "\">"
			} else {
				startTag = "<pre><code>"
			}
			endTag = "</code></pre>"
			infos = append(infos, &entityInfo{
				offset:   offset,
				length:   length,
				startTag: startTag,
				endTag:   endTag,
				id:       i,
				isPre:    true,
			})
			continue
		case *tg.MessageEntityStrike:
			offset, length = e.Offset, e.Length
			startTag, endTag = "<s>", "</s>"
		case *tg.MessageEntityUnderline:
			offset, length = e.Offset, e.Length
			startTag, endTag = "<u>", "</u>"
		case *tg.MessageEntityBlockquote:
			offset, length = e.Offset, e.Length
			if offset < 0 || offset+length > len(runes) {
				continue
			}
			if e.Collapsed {
				startTag, endTag = "<blockquote class=\"expandable\">", "</blockquote>"
			} else {
				startTag, endTag = "<blockquote>", "</blockquote>"
			}
			infos = append(infos, &entityInfo{
				offset:       offset,
				length:       length,
				startTag:     startTag,
				endTag:       endTag,
				id:           i,
				isBlockquote: true,
			})
			continue
		case *tg.MessageEntitySpoiler:
			offset, length = e.Offset, e.Length
			startTag, endTag = "<span class=\"spoiler\">", "</span>"
		case *tg.MessageEntityTextURL:
			offset, length = e.Offset, e.Length
			startTag = "<a href=\"" + html.EscapeString(e.URL) + "\">"
			endTag = "</a>"
		case *tg.MessageEntityMentionName:
			offset, length = e.Offset, e.Length
			startTag = "<a href=\"tg://user?id=" + strconv.FormatInt(e.UserID, 10) + "\">"
			endTag = "</a>"
		case *tg.MessageEntityMention:
			offset, length = e.Offset, e.Length
			if offset < 0 || offset+length > len(runes) {
				continue
			}
			username := strings.TrimPrefix(string(runes[offset:offset+length]), "@")
			startTag = "<a href=\"https://t.me/" + username + "\">"
			endTag = "</a>"
		case *tg.MessageEntityURL:
			offset, length = e.Offset, e.Length
			if offset < 0 || offset+length > len(runes) {
				continue
			}
			url := string(runes[offset : offset+length])
			startTag = "<a href=\"" + html.EscapeString(url) + "\">"
			endTag = "</a>"
		case *tg.MessageEntityCustomEmoji:
			continue
		default:
			continue
		}

		if offset < 0 || offset+length > len(runes) {
			continue
		}

		// Trim trailing whitespace from entity boundaries
		// Telegram's markdown parser often includes trailing spaces in entities
		for length > 0 && (runes[offset+length-1] == ' ' || runes[offset+length-1] == '\t') {
			length--
		}
		if length <= 0 {
			continue
		}

		infos = append(infos, &entityInfo{
			offset:   offset,
			length:   length,
			startTag: startTag,
			endTag:   endTag,
			id:       i,
		})
	}

	if len(infos) == 0 {
		return html.EscapeString(text)
	}

	var events []tagEvent
	for _, info := range infos {
		events = append(events, tagEvent{
			pos:      info.offset,
			isStart:  true,
			entity:   info,
			priority: 0, // starts first
		})
		events = append(events, tagEvent{
			pos:      info.offset + info.length,
			isStart:  false,
			entity:   info,
			priority: 1, // ends after starts at same position
		})
	}

	// Sort events: by position, then starts before ends, then by entity length (longer entities wrap shorter)
	slices.SortFunc(events, func(a, b tagEvent) int {
		if c := cmp.Compare(a.pos, b.pos); c != 0 {
			return c
		}
		if c := cmp.Compare(a.priority, b.priority); c != 0 {
			return c
		}
		// For starts at same position: longer entities should open first (wrap shorter)
		// For ends at same position: shorter entities should close first
		if a.isStart {
			return cmp.Compare(b.entity.length, a.entity.length)
		}
		return cmp.Compare(a.entity.length, b.entity.length)
	})

	var result strings.Builder
	var openStack []*entityInfo // stack of currently open entities
	lastPos := 0
	alreadyClosed := make(map[int]bool) // track entities we've already written close tags for

	// Track which entities are closing at each position (for same-position optimization)
	closingAt := make(map[int]map[int]bool) // pos -> entity id -> true
	for _, event := range events {
		if !event.isStart {
			if closingAt[event.pos] == nil {
				closingAt[event.pos] = make(map[int]bool)
			}
			closingAt[event.pos][event.entity.id] = true
		}
	}

	// Helper to check if currently inside a blockquote
	insideBlockquote := func() bool {
		for _, e := range openStack {
			if e.isBlockquote {
				return true
			}
		}
		return false
	}

	insidePre := func() bool {
		for _, e := range openStack {
			if e.isPre {
				return true
			}
		}
		return false
	}

	for _, event := range events {
		if event.pos > lastPos {
			rawText := string(runes[lastPos:event.pos])
			var text string
			if insidePre() {
				// Don't escape content inside pre blocks (will become markdown code fence)
				text = rawText
			} else {
				text = html.EscapeString(rawText)
				if insideBlockquote() {
					text = strings.ReplaceAll(text, "\n", "<br>")
				}
			}
			result.WriteString(text)
			lastPos = event.pos
		}

		if event.isStart {
			result.WriteString(event.entity.startTag)
			openStack = append(openStack, event.entity)
		} else {
			if alreadyClosed[event.entity.id] {
				continue
			}

			idx := -1
			for i, e := range openStack {
				if e.id == event.entity.id {
					idx = i
					break
				}
			}

			if idx >= 0 {
				var toClose []*entityInfo
				for i := len(openStack) - 1; i >= idx; i-- {
					toClose = append(toClose, openStack[i])
				}

				for _, e := range toClose {
					result.WriteString(e.endTag)
					alreadyClosed[e.id] = true
				}

				var newStack []*entityInfo
				for _, e := range openStack {
					if !alreadyClosed[e.id] {
						newStack = append(newStack, e)
					}
				}
				openStack = newStack

				// Reopen tags that were closed but are NOT ending at this position
				for _, e := range toClose {
					if e.id == event.entity.id {
						continue // This is the one we're actually closing
					}
					if closingAt[event.pos] != nil && closingAt[event.pos][e.id] {
						continue // Also closing at this position
					}
					// Reopen and add back to stack
					result.WriteString(e.startTag)
					openStack = append(openStack, e)
					delete(alreadyClosed, e.id)
				}
			}
		}
	}

	if lastPos < len(runes) {
		text := html.EscapeString(string(runes[lastPos:]))
		if insideBlockquote() {
			text = strings.ReplaceAll(text, "\n", "<br>")
		}
		result.WriteString(text)
	}

	return result.String()
}
