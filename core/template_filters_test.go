package core

import (
	"testing"
	"wispy-core/common"
	"wispy-core/models"
	"wispy-core/tests"
)

func TestTemplateFilters(t *testing.T) {
	tmpl := NewTemplateEngine(DefaultFunctionMap())

	filter_tests := []struct {
		name   string
		tmpl   string
		ctx    *models.TemplateContext
		expect string
	}{
		{
			name:   "upcase filter",
			tmpl:   common.WrapBraces(`"Parker Moore" | upcase`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "PARKER MOORE",
		},
		{
			name:   "upcase filter on already uppercase",
			tmpl:   common.WrapBraces(`"APPLE" | upcase`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "APPLE",
		},
		{
			name:   "downcase filter",
			tmpl:   common.WrapBraces(`"Parker Moore" | downcase`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "parker moore",
		},
		{
			name:   "remove filter",
			tmpl:   common.WrapBraces(`"I strained to see the train through the rain" | remove: "rain"`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "I sted to see the t through the ",
		},
		{
			name:   "replace filter",
			tmpl:   common.WrapBraces(`"Hello, world" | replace: "world", "universe"`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "Hello, universe",
		},
		{
			name:   "strip filter",
			tmpl:   common.WrapBraces(`"<p>Hello, world</p>" | strip`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "Hello, world",
		},
		{
			name:   "trim filter",
			tmpl:   common.WrapBraces(`"  Hello, world  " | trim`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "Hello, world",
		},
		{
			name:   "append filter",
			tmpl:   common.WrapBraces(`"Hello" | append: ", world"`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "Hello, world",
		},
		{
			name:   "prepend filter",
			tmpl:   common.WrapBraces(`"world" | prepend: "Hello, "`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "Hello, world",
		},
		{
			name:   "multiple filters",
			tmpl:   common.WrapBraces(`"hello" | upcase | append: " WORLD"`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "HELLO WORLD",
		},
		{
			name:   "variable with filter",
			tmpl:   common.WrapBraces(`name | upcase`),
			ctx:    NewTemplateContext(map[string]interface{}{"name": "John"}, tmpl),
			expect: "JOHN",
		},
		{
			name:   "split filter debug for the failing case",
			tmpl:   common.WrapBraces(`"John, Paul, George, Ringo" | split: ", " | join: "-"`),
			ctx:    NewTemplateContext(map[string]interface{}{}, tmpl),
			expect: "John-Paul-George-Ringo",
		},
	}

	for _, tc := range filter_tests {
		out, errs := tmpl.Render(tc.tmpl, tc.ctx)
		if len(errs) > 0 {
			t.Error(tests.LogFail(tc.name))
			for _, err := range errs {
				t.Log(tests.LogWarn(err.Error()))
			}
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
