package tests

import (
	"testing"
	"wispy-core/core"
	"wispy-core/models"
)

// TestTemplateEngine runs a suite of tests for the TemplateEngine with various templates and data structures.
func TestTemplateEngine(t *testing.T) {
	tmpl := core.NewTemplateEngine(core.DefaultFunctionMap)

	tests := []struct {
		name       string
		tmpl       string
		ctx        *models.TemplateContext
		expect     string
		shouldFail bool
	}{
		{
			name:   "simple variable",
			tmpl:   "Hello, {{name}}!",
			ctx:    &models.TemplateContext{Data: map[string]interface{}{"name": "Alice"}, Engine: tmpl},
			expect: "Hello, Alice!",
		},
		{
			name:   "dot notation",
			tmpl:   "User: {{user.name}}",
			ctx:    &models.TemplateContext{Data: map[string]interface{}{"user": map[string]interface{}{"name": "Bob"}}, Engine: tmpl},
			expect: "User: Bob",
		},
		{
			name:   "if tag true",
			tmpl:   "{{if show}}Visible{{end-if}}",
			ctx:    &models.TemplateContext{Data: map[string]interface{}{"show": true}, Engine: tmpl},
			expect: "Visible",
		},
		{
			name:   "if tag false",
			tmpl:   "{{if show}}Visible{{end-if}}",
			ctx:    &models.TemplateContext{Data: map[string]interface{}{"show": false}, Engine: tmpl},
			expect: "",
		},
		{
			name:   "unless tag",
			tmpl:   "{{unless show}}Hidden{{end-unless}}",
			ctx:    &models.TemplateContext{Data: map[string]interface{}{"show": false}, Engine: tmpl},
			expect: "Hidden",
		},
		{
			name:   "for tag",
			tmpl:   "{{for item in items}}-{{item}}-{{end-for}}",
			ctx:    &models.TemplateContext{Data: map[string]interface{}{"items": []string{"a", "b"}}, Engine: tmpl},
			expect: "-a--b-",
		},
		{
			name:   "assign tag",
			tmpl:   "{{assign foo bar}}{{foo}}",
			ctx:    &models.TemplateContext{Data: map[string]interface{}{}, Engine: tmpl},
			expect: "bar",
		},
		{
			name:   "case tag",
			tmpl:   `{{case x}}{{when a}}A{{when b}}B{{when c}}C{{end-case}}`,
			ctx:    &models.TemplateContext{Data: map[string]interface{}{"x": "b"}, Engine: tmpl},
			expect: "B",
		},
		{
			name:   "define/render",
			tmpl:   `{{define "block"}}Block: {{msg}}{{end-define}}{{render "block"}}`,
			ctx:    &models.TemplateContext{Data: map[string]interface{}{"msg": "Hello"}, Engine: tmpl},
			expect: "Block: Hello",
		},
		{
			name:   "comment tag",
			tmpl:   "A{{comment}}hidden{{end-comment}}B",
			ctx:    &models.TemplateContext{Data: map[string]interface{}{}, Engine: tmpl},
			expect: "AB",
		},
	}

	for _, tc := range tests {
		out, errs := tmpl.Render(tc.tmpl, tc.ctx)
		if tc.shouldFail {
			if len(errs) == 0 {
				t.Error(logFail(tc.name))
			} else {
				t.Log(logPass(tc.name))
			}
			continue
		}
		if len(errs) > 0 {
			t.Error(logFail(tc.name))
			for _, err := range errs {
				t.Log(logWarn(err.Error()))
			}
			t.Errorf("unexpected error for %s: %v", tc.name, errs)
			continue
		}
		if out != tc.expect {
			t.Error(logFail(tc.name))
			t.Log("expected '%s', got '%s'", tc.expect, out)
		} else {
			t.Log(logPass(tc.name))
		}
	}
}
