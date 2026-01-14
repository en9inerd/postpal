package zola

import (
	"html"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gotd/td/tg"
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

	// Handle code blocks WITH language → convert to markdown
	codeBlockWithLangRegex := regexp.MustCompile(`<pre><code class="language-(.*?)">([\s\S]*?)</code></pre>`)
	codeBlockPlaceholders := make(map[string]string)
	codeIndex := 0
	content = codeBlockWithLangRegex.ReplaceAllStringFunc(content, func(match string) string {
		matches := codeBlockWithLangRegex.FindStringSubmatch(match)
		language := matches[1]
		codeContent := strings.TrimRight(matches[2], "\n")
		markdownCodeBlock := "```" + language + "\n" + codeContent + "\n```"
		placeholder := "___CODEBLOCK_" + string(rune('A'+codeIndex)) + "___"
		codeBlockPlaceholders[placeholder] = markdownCodeBlock
		codeIndex++
		return placeholder
	})

	// Handle code blocks WITHOUT language → preserve as placeholder to protect from newline conversion
	codeBlockNoLangRegex := regexp.MustCompile(`<pre><code>([\s\S]*?)</code></pre>`)
	content = codeBlockNoLangRegex.ReplaceAllStringFunc(content, func(match string) string {
		matches := codeBlockNoLangRegex.FindStringSubmatch(match)
		codeContent := strings.TrimRight(matches[1], "\n")
		// Convert to markdown code block without language
		markdownCodeBlock := "```\n" + codeContent + "\n```"
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

	// Convert legacy spoiler tags if present
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

// entityInfo holds entity position and tag info
type entityInfo struct {
	offset   int
	length   int
	startTag string
	endTag   string
	id       int // unique ID for tracking in stack
}

// tagEvent represents an opening or closing tag at a position
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
			lang := e.Language
			if lang != "" {
				startTag = "<pre><code class=\"language-" + lang + "\">"
			} else {
				startTag = "<pre><code>"
			}
			endTag = "</code></pre>"
		case *tg.MessageEntityStrike:
			offset, length = e.Offset, e.Length
			startTag, endTag = "<s>", "</s>"
		case *tg.MessageEntityUnderline:
			offset, length = e.Offset, e.Length
			startTag, endTag = "<u>", "</u>"
		case *tg.MessageEntityBlockquote:
			offset, length = e.Offset, e.Length
			startTag, endTag = "<blockquote>", "</blockquote>"
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
			username := strings.TrimPrefix(string(runes[offset:offset+length]), "@")
			startTag = "<a href=\"https://t.me/" + username + "\">"
			endTag = "</a>"
		case *tg.MessageEntityURL:
			offset, length = e.Offset, e.Length
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
	sort.Slice(events, func(i, j int) bool {
		if events[i].pos != events[j].pos {
			return events[i].pos < events[j].pos
		}
		if events[i].priority != events[j].priority {
			return events[i].priority < events[j].priority
		}
		// For starts at same position: longer entities should open first (wrap shorter)
		// For ends at same position: shorter entities should close first
		if events[i].isStart {
			return events[i].entity.length > events[j].entity.length
		}
		return events[i].entity.length < events[j].entity.length
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

	for _, event := range events {
		if event.pos > lastPos {
			result.WriteString(html.EscapeString(string(runes[lastPos:event.pos])))
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
		result.WriteString(html.EscapeString(string(runes[lastPos:])))
	}

	return result.String()
}
