package zola

import (
	"strings"
	"testing"
	"time"
)

func TestProcessContent_Empty(t *testing.T) {
	result := ProcessContent("")
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestProcessContent_SimpleText(t *testing.T) {
	input := "Hello world"
	expected := "Hello world"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessContent_LineBreaks(t *testing.T) {
	input := "Line 1\nLine 2\nLine 3"
	expected := "Line 1  \nLine 2  \nLine 3"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessContent_DoubleLineBreaks(t *testing.T) {
	input := "Paragraph 1\n\nParagraph 2"
	expected := "Paragraph 1  \n  \nParagraph 2"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessContent_MultipleLineBreaks(t *testing.T) {
	input := "Paragraph 1\n\n\nParagraph 2"
	expected := "Paragraph 1<br>  \n  \nParagraph 2"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessContent_ManyLineBreaks(t *testing.T) {
	input := "Paragraph 1\n\n\n\n\nParagraph 2"
	expected := "Paragraph 1<br><br><br>  \n  \nParagraph 2"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// Note: Spoiler tag tests removed - Telegram sends MessageEntitySpoiler entities,
// not HTML tags. EntitiesToHTML converts these to <span class="spoiler"> directly.

func TestProcessContent_InlineCode(t *testing.T) {
	input := "Use <code>fmt.Println()</code> to print"
	expected := "Use <code>fmt.Println()</code> to print"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessContent_InlineCodeWithAngleBrackets(t *testing.T) {
	// EntitiesToHTML already escapes < and > via html.EscapeString
	// So ProcessContent receives pre-escaped content
	input := "Check <code>if x &lt; 10 &amp;&amp; y &gt; 5</code> condition"
	expected := "Check <code>if x &lt; 10 &amp;&amp; y &gt; 5</code> condition"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessContent_CodeBlock(t *testing.T) {
	input := `<pre><code class="language-go">package main

func main() {
    println("Hello")
}
</code></pre>`
	// Code blocks should preserve their internal formatting (no double spaces)
	// But the content inside <pre> tags doesn't get line break processing
	expected := "```go\npackage main\n\nfunc main() {\n    println(\"Hello\")\n}\n```"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessContent_CodeBlockWithLanguage(t *testing.T) {
	input := `<pre><code class="language-javascript">console.log("test");</code></pre>`
	expected := "```javascript\nconsole.log(\"test\");\n```"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessContent_Blockquote(t *testing.T) {
	// Blockquote newlines are now handled by EntitiesToHTML, not ProcessContent
	// ProcessContent receives blockquotes with <br> already inserted
	input := "<blockquote>This is a quote<br>with multiple lines</blockquote>"
	expected := "<blockquote>This is a quote<br>with multiple lines</blockquote>"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessContent_ComplexContent(t *testing.T) {
	// Note: Blockquote newlines are handled by EntitiesToHTML, so input has <br> already
	input := `Here's some text
with line breaks.

<blockquote>This is quoted<br>text</blockquote>

More text with <span class="spoiler">hidden content</span>.

<pre><code class="language-go">func test() {
    return true
}
</code></pre>

Final text.`
	// Code blocks should preserve their internal formatting (no double spaces)
	expected := "Here's some text  \nwith line breaks.  \n  \n<blockquote>This is quoted<br>text</blockquote>  \n  \nMore text with <span class=\"spoiler\">hidden content</span>.  \n  \n```go\nfunc test() {\n    return true\n}\n```  \n  \nFinal text."
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestProcessContent_HTMLEntities(t *testing.T) {
	// TypeScript implementation does NOT decode HTML entities
	// They remain as-is in the output
	input := "Text with &lt;entities&gt; and &amp; symbols"
	expected := "Text with &lt;entities&gt; and &amp; symbols"
	result := ProcessContent(input)
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestExtractTitle_EmptyContent(t *testing.T) {
	result := ExtractTitle("", "@channel")
	if result != "@channel" {
		t.Errorf("Expected @channel, got %q", result)
	}
}

func TestExtractTitle_NoAddress(t *testing.T) {
	content := "Some content without address"
	result := ExtractTitle(content, "@channel")
	if result != "@channel" {
		t.Errorf("Expected @channel, got %q", result)
	}
}

func TestExtractTitle_WithAddress(t *testing.T) {
	content := "Some content\n0x1234abcd"
	result := ExtractTitle(content, "@channel")
	expected := "@channel [0x1234abcd]"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestExtractTitle_WithAddressAndNewline(t *testing.T) {
	content := "Some content\n0x1234abcd\n"
	result := ExtractTitle(content, "@channel")
	expected := "@channel [0x1234abcd]"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestExtractTitle_WithAddressAndWhitespace(t *testing.T) {
	content := "Some content  \n0x1234abcd"
	result := ExtractTitle(content, "@channel")
	expected := "@channel [0x1234abcd]"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestExtractTitle_WithAddressUppercase(t *testing.T) {
	content := "Some content\n0xABCD1234"
	result := ExtractTitle(content, "@channel")
	expected := "@channel [0xABCD1234]"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRemoveAddressPattern(t *testing.T) {
	content := "Some content\n0x1234abcd"
	result := RemoveAddressPattern(content)
	expected := "Some content\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRemoveAddressPattern_WithNewline(t *testing.T) {
	content := "Some content\n0x1234abcd\n"
	result := RemoveAddressPattern(content)
	expected := "Some content\n"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRemoveAddressPattern_NoAddress(t *testing.T) {
	content := "Some content without address"
	result := RemoveAddressPattern(content)
	if result != content {
		t.Errorf("Expected unchanged content, got %q", result)
	}
}

func TestBuildFrontMatter_Simple(t *testing.T) {
	post := Post{
		ID:      123,
		Title:   "Test Post",
		Content: "Content here",
		Date:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	result := BuildFrontMatter(post)
	expected := `+++
title = "Test Post"
date = 2024-01-15T10:30:00Z
+++

`
	if result != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestBuildFrontMatter_WithImages(t *testing.T) {
	post := Post{
		ID:         456,
		Title:      "Post with Images",
		Content:    "Content",
		Date:       time.Date(2024, 2, 20, 15, 45, 0, 0, time.UTC),
		ImageNames: []string{"image_0.jpg", "image_1.png"},
	}
	result := BuildFrontMatter(post)
	expected := `+++
title = "Post with Images"
date = 2024-02-20T15:45:00Z

[extra]
images = ["image_0.jpg", "image_1.png"]
+++

`
	if result != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestBuildFrontMatter_WithQuotesInTitle(t *testing.T) {
	post := Post{
		ID:      789,
		Title:   `Title with "quotes"`,
		Content: "Content",
		Date:    time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
	}
	result := BuildFrontMatter(post)
	expected := `+++
title = "Title with \"quotes\""
date = 2024-03-01T12:00:00Z
+++

`
	if result != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result)
	}
}

func TestBuildFrontMatter_EmptyImages(t *testing.T) {
	post := Post{
		ID:         999,
		Title:      "No Images",
		Content:    "Content",
		Date:       time.Date(2024, 4, 10, 8, 0, 0, 0, time.UTC),
		ImageNames: []string{},
	}
	result := BuildFrontMatter(post)
	// Should not include [extra] section if no images
	if strings.Contains(result, "[extra]") {
		t.Errorf("Expected no [extra] section, got:\n%q", result)
	}
}
