package core

import (
	"testing"
	"wispy-core/models"
	"wispy-core/tests"
)

// TestTemplateEngine runs a suite of tests for the TemplateEngine with various templates and data structures.
func TestTemplateEngine(t *testing.T) {
	tmpl := NewTemplateEngine(DefaultFunctionMap())

	template_engine_tests := []struct {
		name       string
		tmpl       string
		ctx        *models.TemplateContext
		expect     string
		shouldFail bool
	}{
		{
			name:   "simple variable",
			tmpl:   "Hello, {%name%}!",
			ctx:    NewTemplateContext(map[string]interface{}{"name": "Alice"}, tmpl),
			expect: "Hello, Alice!",
		},
		{
			name:   "dot notation",
			tmpl:   "User: {%user.name%}",
			ctx:    NewTemplateContext(map[string]interface{}{"user": map[string]interface{}{"name": "Bob"}}, tmpl),
			expect: "User: Bob",
		},
		{
			name:   "if tag true",
			tmpl:   "{%if show%}Visible{%endif%}",
			ctx:    NewTemplateContext(map[string]interface{}{"show": true}, tmpl),
			expect: "Visible",
		},
		{
			name:   "if tag false",
			tmpl:   "{%if show%}Visible{%endif%}",
			ctx:    NewTemplateContext(map[string]interface{}{"show": false}, tmpl),
			expect: "",
		},
		{
			name:   "unless tag",
			tmpl:   "{%unless show%}Hidden{%endunless%}",
			ctx:    NewTemplateContext(map[string]interface{}{"show": false}, tmpl),
			expect: "Hidden",
		},
		{
			name:   "for tag",
			tmpl:   "{%for item in items%}-{%item%}-{%endfor%}",
			ctx:    NewTemplateContext(map[string]interface{}{"items": []string{"a", "b"}}, tmpl),
			expect: "-a--b-",
		},
		{
			name:   "assign tag",
			tmpl:   "{%assign foo bar%}{%foo%}",
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "bar",
		},
		{
			name:   "case tag",
			tmpl:   `{%case x%}{%when a%}A{%when b%}B{%when c%}C{%endcase%}`,
			ctx:    NewTemplateContext(map[string]interface{}{"x": "b"}, tmpl),
			expect: "B",
		},
		{
			name:   "define/render",
			tmpl:   `{%define "block"%}Block: {%msg%}{%enddefine%}{%render "block"%}`,
			ctx:    NewTemplateContext(map[string]interface{}{"msg": "Hello"}, tmpl),
			expect: "Block: Hello",
		},
		{
			name:   "comment tag",
			tmpl:   "A{%comment%}hidden{%endcomment%}B",
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "AB",
		},
	}

	for _, tc := range template_engine_tests {
		out, errs := tmpl.Render(tc.tmpl, tc.ctx)
		if tc.shouldFail {
			if len(errs) == 0 {
				t.Error(tests.LogFail(tc.name))
			} else {
				t.Log(tests.LogPass(tc.name))
			}
			continue
		}
		if len(errs) > 0 {
			t.Error(tests.LogFail(tc.name))
			for _, err := range errs {
				t.Log(tests.LogWarn(err.Error()))
			}
			t.Error(tests.LogWarn("unexpected error for %s: %v", tc.name, errs))
			continue
		}
		if out != tc.expect {
			t.Error(tests.LogFail(tc.name))
			t.Log(tests.LogWarn("expected '%s', got '%s'", tc.expect, out))
		} else {
			t.Log(tests.LogPass(tc.name))
		}
	}
}
