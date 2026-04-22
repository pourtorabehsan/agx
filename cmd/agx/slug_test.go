package main

import "testing"

func TestSlug(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"refactor auth module", "refactor-auth-module"},
		{"  Fix THE Bug  ", "fix-the-bug"},
		{"hello!!! world???", "hello-world"},
		{"", ""},
		{"!!!", ""},
		{"a", "a"},
		{"café au lait", "caf-au-lait"},
		{"add feature/authN to login.go", "add-feature-authn-to-login-go"},
		// 60-char input → slug truncated to <= 40, no trailing hyphen.
		{
			"this prompt is rather long indeed and should be truncated here",
			"this-prompt-is-rather-long-indeed-and",
		},
	}
	for _, c := range cases {
		if got := Slug(c.in); got != c.want {
			t.Errorf("Slug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
