package main

import "strings"

const slugMaxLen = 40

func Slug(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteByte('-')
		}
	}
	s = b.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if len(s) > slugMaxLen {
		s = s[:slugMaxLen]
		if idx := strings.LastIndex(s, "-"); idx > 0 {
			s = s[:idx]
		}
	}
	return s
}
