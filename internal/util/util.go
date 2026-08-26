package util

import (
	"fmt"
	"time"

	"golang.org/x/text/language"
)

func Pointer[T any](v T) *T { return &v }

// TruncateString cuts s down to at most maxLen runes, leaving it untouched if it's already short enough.
func TruncateString(s string, maxLen int) string {
	r := []rune(s)
	if len(r) <= maxLen {
		return s
	}
	return string(r[:maxLen])
}

func ParseLanguageTag(lang string) language.Tag {
	languageTag, err := language.Parse(lang)
	if err != nil {
		languageTag = language.English
	}
	return languageTag
}

func ParseDate(date string) (*time.Time, error) {
	formats := []string{time.RFC3339, time.DateOnly, time.DateTime}

	for _, format := range formats {
		t, err := time.Parse(format, date)
		if err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("invalid date format %q", date)
}
