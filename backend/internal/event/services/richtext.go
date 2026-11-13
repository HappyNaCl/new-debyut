package services

import (
	"html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// richTextPolicy is the allowlist for event descriptions and terms. It keeps
// only the formatting the editor can produce; everything else (scripts, event
// handler attributes, styles, iframes, images, javascript: and data: links) is
// removed. Stored rich text is always the output of this policy, so it is safe
// to render as HTML.
var richTextPolicy = func() *bluemonday.Policy {
	policy := bluemonday.NewPolicy()
	policy.AllowElements("p", "br", "strong", "b", "em", "i", "u", "s", "h2", "h3", "ul", "ol", "li", "blockquote", "hr")
	policy.AllowAttrs("href").OnElements("a")
	policy.AllowURLSchemes("http", "https", "mailto")
	policy.RequireParseableURLs(true)
	policy.RequireNoFollowOnLinks(true)
	policy.RequireNoReferrerOnLinks(true)
	policy.AddTargetBlankToFullyQualifiedLinks(true)
	return policy
}()

var textPolicy = bluemonday.StrictPolicy()

var whitespace = regexp.MustCompile(`\s+`)

// sanitizeRichText returns the safe HTML for user-supplied rich text, or ""
// when it has no visible text (e.g. an empty "<p></p>").
func sanitizeRichText(raw string) string {
	clean := strings.TrimSpace(richTextPolicy.Sanitize(raw))
	if richTextLength(clean) == 0 {
		return ""
	}
	return clean
}

// richTextLength counts the visible characters in sanitized rich text, which
// is what the length limits apply to.
func richTextLength(clean string) int {
	text := html.UnescapeString(textPolicy.Sanitize(strings.ReplaceAll(clean, "<", " <")))
	return len([]rune(strings.TrimSpace(whitespace.ReplaceAllString(text, " "))))
}
