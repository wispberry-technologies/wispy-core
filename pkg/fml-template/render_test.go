package template_test

import (
	"html/template"
	"strings"
	"testing"
	wispyTemplate "wispy-core/pkg/template"
	"wispy-core/pkg/template/template_tags"
)

var raw = `
Dear {{.Name}},
{{if .Attended}}
It was a pleasure to see you at the wedding.
{{else}}
It is a shame you couldn't make it to the wedding.
{{end}}
{{with .Gift}}
Thank you for the lovely {{.}}.
{{end}}
Best wishes,
Josie
`

func TestNewEngine(t *testing.T) {
	var data = map[string]any{"Name": "Aunt Mildred", "Gift": "bone china tea set", "Attended": true}
	engine := wispyTemplate.NewEngine(wispyTemplate.EngineConfig{})
	result, errs := engine.Render(raw, data)

	if len(errs) > 0 {
		t.Errorf("Expected no errors, got: %q", errs)
	}

	t.Logf("Parsed Result: %q", result)
}

// Bench mark new engine vs html/template engine
func BenchmarkNewEngine(b *testing.B) {
	var data = map[string]any{"Name": "Aunt Mildred", "Gift": "bone china tea set", "Attended": true}
	engine := wispyTemplate.NewEngine(wispyTemplate.EngineConfig{
		TemplateTags: []*wispyTemplate.TemplateTag{
			template_tags.TemplateIfTag,
			template_tags.TemplateElseTag,
			template_tags.TemplateRangeTag,
			template_tags.TemplateWithTag,
			template_tags.TemplateDefineTag,
			template_tags.TemplateBlockTag,
			template_tags.TemplateTemplateTag,
		},
	})
	for i := 0; i < b.N; i++ {
		_, errs := engine.Render(raw, data)
		if len(errs) > 0 {
			b.Errorf("Expected no errors, got: %q", errs)
		}
	}
}

func BenchmarkHTMLTemplateEngine(b *testing.B) {
	var data = map[string]any{"Name": "Aunt Mildred", "Gift": "bone china tea set", "Attended": true}
	for i := 0; i < b.N; i++ {
		tmpl, err := template.New("test").Parse(raw)
		if err != nil {
			b.Errorf("Expected no errors, got: %v", err)
		}

		// Create a strings.Builder to capture the output
		var buf strings.Builder
		err = tmpl.Execute(&buf, data)
		if err != nil {
			b.Errorf("Expected no errors, got: %v", err)
		}
	}
}
