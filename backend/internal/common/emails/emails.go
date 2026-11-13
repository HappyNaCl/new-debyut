// Package emails holds the single rule for how email addresses are stored and
// compared: trimmed and lowercased. Normalize every email at the service
// boundary so the database only ever sees the canonical form and lookups can
// use plain equality.
package emails

import "strings"

func Normalize(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
