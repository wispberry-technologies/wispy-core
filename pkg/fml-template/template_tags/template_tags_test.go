package template_tags_test

import (
	"strings"
	"testing"
	"wispy-core/pkg/template"
	wispyTemplate "wispy-core/pkg/template"
	"wispy-core/pkg/template/template_tags"
)

func TestIfTag(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "if true condition",
			template: `{{ if .Value }}Show this{{ end }}`,
			data:     map[string]any{"Value": true},
			expected: `Show this`,
		},
		{
			name:     "if false condition",
			template: `{{ if .Value }}Show this{{ end }}`,
			data:     map[string]any{"Value": false},
			expected: ``,
		},
		{
			name:     "if with else - true condition",
			template: `{{ if .Value }}True{{ else }}False{{ end }}`,
			data:     map[string]any{"Value": true},
			expected: `True`,
		},
		{
			name:     "if with else - false condition",
			template: `{{ if .Value }}True{{ else }}False{{ end }}`,
			data:     map[string]any{"Value": false},
			expected: `False`,
		},
		{
			name:     "nested if statements",
			template: `{{ if .A }}A{{ if .B }}B{{ end }}{{ end }}`,
			data:     map[string]any{"A": true, "B": true},
			expected: `AB`,
		},
	}

	engine := wispyTemplate.NewEngine(wispyTemplate.EngineConfig{
		TemplateTags: []*wispyTemplate.TemplateTag{
			template_tags.TemplateIfTag,
			template_tags.TemplateElseTag,
		},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := engine.Render(tt.template, tt.data)
			if len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRangeTag(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "range over slice",
			template: `{{ range .Items }}{{ . }}{{ end }}`,
			data:     map[string]any{"Items": []any{"a", "b", "c"}},
			expected: `abc`,
		},
		{
			name:     "range with index",
			template: `{{ range .Items }}{{ .index }}:{{ . }} {{ end }}`,
			data:     map[string]any{"Items": []any{"a", "b", "c"}},
			expected: `0:a 1:b 2:c `,
		},
		{
			name:     "range over map",
			template: `{{ range .Items }}{{ .key }}={{ . }} {{ end }}`,
			data:     map[string]any{"Items": map[string]any{"a": "1", "b": "2"}},
			expected: `a=1 b=2 `,
		},
		{
			name:     "range over empty slice",
			template: `{{ range .Items }}{{ . }}{{ end }}`,
			data:     map[string]any{"Items": []any{}},
			expected: ``,
		},
	}

	engine := wispyTemplate.NewEngine(wispyTemplate.EngineConfig{
		TemplateTags: []*wispyTemplate.TemplateTag{
			template_tags.TemplateRangeTag,
		},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := engine.Render(tt.template, tt.data)
			if len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestWithTag(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "with context",
			template: `{{ with .Person }}Name: {{ .Name }}, Age: {{ .Age }}{{ end }}`,
			data:     map[string]any{"Person": map[string]any{"Name": "John", "Age": 30}},
			expected: `Name: John, Age: 30`,
		},
		{
			name:     "with nil context",
			template: `{{ with .Person }}Show this{{ end }}`,
			data:     map[string]any{"Person": nil},
			expected: ``,
		},
		{
			name:     "with nested context",
			template: `{{ with .Person }}{{ with .Address }}{{ .City }}, {{ .Country }}{{ end }}{{ end }}`,
			data: map[string]any{
				"Person": map[string]any{
					"Address": map[string]any{
						"City":    "New York",
						"Country": "USA",
					},
				},
			},
			expected: `New York, USA`,
		},
	}

	engine := wispyTemplate.NewEngine(wispyTemplate.EngineConfig{
		TemplateTags: []*wispyTemplate.TemplateTag{
			template_tags.TemplateWithTag,
		},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := engine.Render(tt.template, tt.data)
			if len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestDefineAndTemplateTag(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name: "define and template",
			template: `
{{ define "greeting" }}Hello, {{ . }}!{{ end }}
{{ template "greeting" .Name }}`,
			data:     map[string]any{"Name": "John"},
			expected: "\n\nHello, John!",
		},
		{
			name: "template with default context",
			template: `
{{ define "userInfo" }}User: {{ .Name }}, Age: {{ .Age }}{{ end }}
{{ template "userInfo" . }}`,
			data:     map[string]any{"Name": "John", "Age": 30},
			expected: "\n\nUser: John, Age: 30",
		},
	}

	engine := template.NewEngine(template.EngineConfig{
		TemplateTags: []*template.TemplateTag{
			template_tags.TemplateDefineTag,
			template_tags.TemplateTemplateTag,
		},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := engine.Render(tt.template, tt.data)
			if len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestBlockTag(t *testing.T) {
	tests := []struct {
		name      string
		templates []string
		data      map[string]any
		expected  string
	}{
		{
			name: "block with default content",
			templates: []string{
				`{{ block "content" . }}Default content{{ end }}`,
			},
			data:     map[string]any{},
			expected: `Default content`,
		},
		{
			name: "block with override",
			templates: []string{
				`{{ define "content" }}Overridden content{{ end }}{{ block "content" . }}Default content{{ end }}`,
			},
			data:     map[string]any{},
			expected: `Overridden content`,
		},
		{
			name: "block with context",
			templates: []string{
				`{{ block "user" .Person }}Name: {{ .Name }}, Age: {{ .Age }}{{ end }}`,
			},
			data:     map[string]any{"Person": map[string]any{"Name": "John", "Age": 30}},
			expected: `Name: John, Age: 30`,
		},
	}

	engine := template.NewEngine(template.EngineConfig{
		TemplateTags: []*template.TemplateTag{
			template_tags.TemplateDefineTag,
			template_tags.TemplateBlockTag,
		},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &strings.Builder{}

			for _, tmpl := range tt.templates {
				result, errs := engine.Render(tmpl, tt.data)
				if len(errs) > 0 {
					t.Errorf("unexpected errors: %v", errs)
				}
				sb.WriteString(result)
			}

			if sb.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, sb.String())
			}
		})
	}
}

func TestAllTagsTogether(t *testing.T) {
	template := `
{{ define "user" }}User: {{ .Name }}, Age: {{ .Age }}{{ end }}

{{ if .ShowUsers }}
  <h1>User List</h1>
  {{ with .Users }}
    {{ range . }}
      {{ template "user" . }}
    {{ end }}
  {{ end }}
{{ else }}
  <p>No users to display</p>
{{ end }}

{{ block "footer" . }}Default Footer{{ end }}
`

	data := map[string]any{
		"ShowUsers": true,
		"Users": []any{
			map[string]any{"Name": "Alice", "Age": 28},
			map[string]any{"Name": "Bob", "Age": 35},
			map[string]any{"Name": "Charlie", "Age": 42},
		},
	}

	expected := `

  <h1>User List</h1>
    
      User: Alice, Age: 28
    
      User: Bob, Age: 35
    
      User: Charlie, Age: 42
    
  

Default Footer
`

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

	result, errs := engine.Render(template, data)
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}

	if result != expected {
		t.Errorf("expected:\n%s\n\ngot:\n%s", expected, result)
	}
}
