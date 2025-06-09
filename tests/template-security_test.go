package tests

import (
	"testing"
	"wispy-core/core"
)

func TestResolveDotNotation_SanitizesUGC(t *testing.T) {
	ctx := map[string]interface{}{
		"bio":    `<b>hello</b>`,
		"script": `<script>alert('x')</script>`,
		"onload": `<img src=x onerror=alert(1)>`,
		"style":  `<style>body{background:red;}</style>`,
		"nested": map[string]interface{}{
			"desc": `<a href='javascript:alert(1)'>link</a>`},
		"plain": "safe text",
		"num":   42,
	}

	tests := []struct {
		key      string
		wantSafe string
	}{
		{"bio", "<b>hello</b>"},
		{"script", "alert('x')"},
		{"onload", "<img src=\"x\">"},
		{"style", "body{background:red;}"},
		{"nested.desc", "<a>link</a>"},
		{"plain", "safe text"},
	}

	for _, tc := range tests {
		got := core.ResolveDotNotation(ctx, tc.key)
		if got != tc.wantSafe {
			t.Errorf("key %q: want %q, got %q", tc.key, tc.wantSafe, got)
		}
	}

	// Non-string values should not be sanitized
	if v := core.ResolveDotNotation(ctx, "num"); v != 42 {
		t.Errorf("expected 42 for 'num', got %v", v)
	}

	// Empty key returns nil
	if v := core.ResolveDotNotation(ctx, ""); v != nil {
		t.Errorf("expected nil for empty key, got %v", v)
	}

	// Nonexistent key returns nil
	if v := core.ResolveDotNotation(ctx, "doesnotexist"); v != nil {
		t.Errorf("expected nil for missing key, got %v", v)
	}
}
