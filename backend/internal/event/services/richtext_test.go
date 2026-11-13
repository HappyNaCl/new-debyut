package services

import (
	"strings"
	"testing"
)

func TestSanitizeRichTextRemovesXss(t *testing.T) {
	payloads := []string{
		`<script>alert(1)</script><p>Hi</p>`,
		`<p onclick="alert(1)">Hi</p>`,
		`<img src=x onerror=alert(1)><p>Hi</p>`,
		`<a href="javascript:alert(1)">Hi</a>`,
		`<a href="JaVaScRiPt:alert(1)">Hi</a>`,
		`<a href="data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==">Hi</a>`,
		`<iframe src="https://evil.example"></iframe><p>Hi</p>`,
		`<svg onload=alert(1)><p>Hi</p></svg>`,
		`<p style="background:url(javascript:alert(1))">Hi</p>`,
		`<math><mtext><table><mglyph><style><img src=x onerror=alert(1)></style></mglyph></table></mtext></math>`,
		`<p>Hi</p><object data="evil.swf"></object><embed src="evil.swf">`,
		`<form action="https://evil.example"><input name=x></form><p>Hi</p>`,
	}
	forbidden := []string{"<script", "onclick", "onerror", "onload", "javascript:", "data:", "<iframe", "<svg", "<img", "style=", "<object", "<embed", "<form", "<input", "<math"}

	for _, payload := range payloads {
		clean := strings.ToLower(sanitizeRichText(payload))
		for _, bad := range forbidden {
			if strings.Contains(clean, bad) {
				t.Errorf("sanitize(%q) = %q still contains %q", payload, clean, bad)
			}
		}
	}
}

func TestSanitizeRichTextKeepsEditorFormatting(t *testing.T) {
	input := `<h2>Lineup</h2><p><strong>Bold</strong> <em>it</em> <u>u</u> <s>s</s></p><ul><li>One</li></ul><ol><li>Two</li></ol><blockquote><p>Q</p></blockquote>`
	if got := sanitizeRichText(input); got != input {
		t.Fatalf("formatting changed:\n got  %s\n want %s", got, input)
	}
}

func TestSanitizeRichTextHardensLinks(t *testing.T) {
	got := sanitizeRichText(`<p><a href="https://example.com">site</a></p>`)
	for _, want := range []string{`href="https://example.com"`, `rel="nofollow noreferrer noopener"`, `target="_blank"`} {
		if !strings.Contains(got, want) {
			t.Errorf("link %q missing %s", got, want)
		}
	}
}

func TestSanitizeRichTextTreatsBlankAsEmpty(t *testing.T) {
	for _, blank := range []string{"", "<p></p>", "<p>   </p><p><br></p>", "<script>alert(1)</script>"} {
		if got := sanitizeRichText(blank); got != "" {
			t.Errorf("sanitize(%q) = %q, want empty", blank, got)
		}
	}
}

func TestRichTextLengthCountsVisibleText(t *testing.T) {
	if got := richTextLength(`<p><strong>Hello</strong></p><p>world &amp; you</p>`); got != len("Hello world & you") {
		t.Fatalf("length = %d", got)
	}
}
