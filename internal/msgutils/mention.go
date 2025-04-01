package msgutils

import "strings"

func ExtractMentions(msg string) []string {
	// This function extracts mentions from a message.
	// A mention is defined as a string that starts with '@' and is followed by alphanumeric characters or underscores.
	// For example, "@user1" and "@user_2" are valid mentions.

	var mentions []string
	words := strings.Fields(msg)
	for _, word := range words {
		if strings.HasPrefix(word, "@") {
			mentions = append(mentions, word[1:])
		}
	}
	return mentions
}
