package stringutils_test

import (
	"testing"

	"github.com/habiliai/agentruntime/internal/stringutils"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeUnicodeString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "string with null byte",
			input:    "title\u0000with null",
			expected: "titlewith null",
		},
		{
			name:     "string with multiple control characters",
			input:    "test\u0000\u0001\u001f\u007fstring",
			expected: "teststring",
		},
		{
			name:     "string with valid whitespace",
			input:    "normal\tstring\nwith\rwhitespace",
			expected: "normal\tstring\nwith\rwhitespace",
		},
		{
			name:     "clean string",
			input:    "completely normal string",
			expected: "completely normal string",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "string with C1 control characters",
			input:    "test\u0080\u009fstring",
			expected: "teststring",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stringutils.SanitizeUnicodeString(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
