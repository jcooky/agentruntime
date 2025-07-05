package stringutils

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// sanitizeUnicodeString removes problematic Unicode characters
func SanitizeUnicodeString(s string) string {
	// Quick check: if string is valid UTF-8 and has no null bytes, return as-is
	if utf8.ValidString(s) && !strings.Contains(s, "\u0000") && !hasControlChars(s) {
		return s
	}

	var builder strings.Builder
	builder.Grow(len(s)) // pre-allocate to avoid reallocations

	for _, r := range s {
		// Remove NULL bytes and control characters except common whitespace
		if r == 0 { // NULL byte
			continue
		}
		if r < 32 && r != '\t' && r != '\n' && r != '\r' { // control characters except tab, newline, carriage return
			continue
		}
		if r == 127 { // DEL character
			continue
		}
		if r >= 128 && r <= 159 { // C1 control characters
			continue
		}

		// Keep printable characters and common whitespace
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// hasControlChars checks if string contains problematic control characters
func hasControlChars(s string) bool {
	for _, r := range s {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return true
		}
		if r == 127 || (r >= 128 && r <= 159) {
			return true
		}
	}
	return false
}
