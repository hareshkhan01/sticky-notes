package note

import (
	"strings"
	"unicode"
)

// ValidateTitle trims and checks a title against MaxTitleRunes.
func ValidateTitle(title string) (string, error) {
	t := strings.TrimSpace(title)
	if n := len([]rune(t)); n > MaxTitleRunes {
		return "", ErrTitleTooLong{Length: n}
	}
	return t, nil
}

// ValidateTopic trims and checks one topic against MaxTopicRunes.
func ValidateTopic(text string) (string, error) {
	t := strings.TrimSpace(text)
	if t == "" {
		return "", ErrEmpty
	}
	if n := len([]rune(t)); n > MaxTopicRunes {
		return "", ErrTopicTooLong{Length: n}
	}
	return t, nil
}

// Sanitize splits raw text into topic candidates: one per non-empty line,
// each trimmed and stripped of common bullet markers. Only runes are kept
// intact; counting is done elsewhere.
func Sanitize(raw string) []string {
	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "-*•→")
		line = strings.TrimSpace(line)
		if line != "" && !strings.ContainsFunc(line, func(r rune) bool { return !unicode.IsPrint(r) && r != '\t' }) {
			out = append(out, line)
		}
	}
	return out
}

// DeriveTitle builds a display title from the first topic when the user
// did not supply one.
func DeriveTitle(topics []Topic) string {
	if len(topics) == 0 {
		return ""
	}
	first := topics[0].Text
	runes := []rune(first)
	if len(runes) <= MaxTitleRunes {
		return first
	}
	return string(runes[:MaxTitleRunes-1]) + "…"
}
