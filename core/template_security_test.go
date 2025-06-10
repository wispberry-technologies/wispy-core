package core

import (
	"testing"
	"wispy-core/tests"
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

	template_engine_sanitization_tests := []struct {
		key      string
		wantSafe string
	}{
		{"bio", "<b>hello</b>"},
		{"script", ""}, // UGCPolicy removes script tags and content
		{"onload", "<img src=\"x\">"},
		{"style", ""},           // UGCPolicy removes style tags and content
		{"nested.desc", "link"}, // e s
		{"plain", "safe text"},
	}

	for _, tc := range template_engine_sanitization_tests {
		got := ResolveDotNotation(ctx, tc.key)
		if got != tc.wantSafe {
			t.Error(tests.LogFail(tc.key))
			t.Log(tests.LogWarn("expected '%s', got '%s'", tc.wantSafe, got))
		} else {
			t.Log(tests.LogPass(tc.key))
		}
	}

	// Non-string values should not be sanitized
	if v := ResolveDotNotation(ctx, "num"); v != 42 {
		t.Error(tests.LogFail("num"))
		t.Log(tests.LogWarn("expected 42 for 'num', got %v", v))
	} else {
		t.Log(tests.LogPass("num"))
	}

	// Empty key returns nil
	if v := ResolveDotNotation(ctx, ""); v != nil {
		t.Error(tests.LogFail("empty key"))
		t.Log(tests.LogWarn("expected nil for empty key, got %v", v))
	} else {
		t.Log(tests.LogPass("empty key"))
	}

	// Nonexistent key returns nil
	if v := ResolveDotNotation(ctx, "doesNotExist"); v != nil {
		t.Error(tests.LogFail("doesNotExist"))
		t.Log(tests.LogWarn("expected nil for missing key, got %v", v))
	} else {
		t.Log(tests.LogPass("doesNotExist"))
	}
}
