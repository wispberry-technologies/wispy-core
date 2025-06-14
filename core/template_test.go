package core

import (
	"testing"
)

func TestTemplateEngine_BasicVariableInterpolation(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := NewTemplateContext(map[string]interface{}{
		"name": "World",
		"user": map[string]interface{}{
			"name": "John",
			"age":  30,
		},
	}, engine)

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "simple variable",
			template: "Hello {{ name }}!",
			expected: "Hello World!",
		},
		{
			name:     "nested variable",
			template: "Hello {{ user.name }}!",
			expected: "Hello John!",
		},
		{
			name:     "variable with filter",
			template: "Hello {{ name | upcase }}!",
			expected: "Hello WORLD!",
		},
		{
			name:     "multiple variables",
			template: "{{ user.name }} is {{ user.age }} years old",
			expected: "John is 30 years old",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := engine.Render(tt.template, ctx)
			if len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngine_IfTag(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := NewTemplateContext(map[string]interface{}{
		"showMessage": true,
		"user": map[string]interface{}{
			"isAdmin": false,
		},
	}, engine)

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "if true",
			template: "{% if showMessage %}Hello World!{% endif %}",
			expected: "Hello World!",
		},
		{
			name:     "if false",
			template: "{% if user.isAdmin %}Admin Content{% endif %}",
			expected: "",
		},
		{
			name:     "if with content around",
			template: "Before {% if showMessage %}Middle{% endif %} After",
			expected: "Before Middle After",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := engine.Render(tt.template, ctx)
			if len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngine_ForTag(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := NewTemplateContext(map[string]interface{}{
		"items": []interface{}{"apple", "banana", "cherry"},
		"users": []interface{}{
			map[string]interface{}{"name": "Alice"},
			map[string]interface{}{"name": "Bob"},
		},
	}, engine)

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "simple loop",
			template: "{% for item in items %}{{ item }},{% endfor %}",
			expected: "apple,banana,cherry,",
		},
		{
			name:     "nested data loop",
			template: "{% for user in users %}{{ user.name }} {% endfor %}",
			expected: "Alice Bob ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := engine.Render(tt.template, ctx)
			if len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngine_ComplexTemplate(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := NewTemplateContext(map[string]interface{}{
		"title": "My Blog",
		"posts": []interface{}{
			map[string]interface{}{
				"title":   "First Post",
				"content": "This is the first post",
				"author":  "John",
			},
			map[string]interface{}{
				"title":   "Second Post",
				"content": "This is the second post",
				"author":  "Jane",
			},
		},
		"showPosts": true,
	}, engine)

	template := `<h1>{{ title | upcase }}</h1>
{% if showPosts %}
<div class="posts">
{% for post in posts %}
  <article>
    <h2>{{ post.title }}</h2>
    <p>{{ post.content }}</p>
    <small>By {{ post.author }}</small>
  </article>
{% endfor %}
</div>
{% endif %}`

	result, errs := engine.Render(template, ctx)
	if len(errs) > 0 {
		t.Errorf("unexpected errors: %v", errs)
	}

	// Check that key elements are present
	expectedSubstrings := []string{
		"<h1>MY BLOG</h1>",
		"<div class=\"posts\">",
		"<h2>First Post</h2>",
		"<h2>Second Post</h2>",
		"By John",
		"By Jane",
	}

	for _, substring := range expectedSubstrings {
		if !contains(result, substring) {
			t.Errorf("expected result to contain %q, but it didn't. Result: %q", substring, result)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestTemplateEngine_ErrorResilience(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())

	tests := []struct {
		name      string
		template  string
		data      map[string]interface{}
		expected  string
		shouldErr bool
		errCount  int
	}{
		{
			name:      "unresolved variable renders empty",
			template:  "Before {{ missing.variable }} After",
			data:      map[string]interface{}{},
			expected:  "Before  After",
			shouldErr: false, // Unresolved variables don't generate errors, just render empty
			errCount:  0,
		},
		{
			name:     "mixed resolved and unresolved variables",
			template: "Hello {{ name }}, your {{ missing.value }} is {{ status }}!",
			data: map[string]interface{}{
				"name":   "John",
				"status": "active",
			},
			expected:  "Hello John, your  is active!",
			shouldErr: false, // Unresolved variables don't generate errors
			errCount:  0,
		},
		{
			name:      "unresolved if condition defaults to false",
			template:  "{% if missing.condition %}Hidden{% endif %}Visible",
			data:      map[string]interface{}{},
			expected:  "Visible",
			shouldErr: false, // Unresolved conditions don't generate errors, just evaluate as false
			errCount:  0,
		},
		{
			name:     "broken filter chain continues rendering",
			template: "Before {{ name | unknown_filter }} After",
			data: map[string]interface{}{
				"name": "John",
			},
			expected:  "Before  After",
			shouldErr: true,
			errCount:  1,
		},
		{
			name:     "multiple errors from filters continue rendering",
			template: "{{ name1 | badfilter }} - {{ name }} - {{ name2 | another_bad_filter }} - END",
			data: map[string]interface{}{
				"name":  "Test",
				"name1": "Value1",
				"name2": "Value2",
			},
			expected:  " - Test -  - END", // Bad filters render empty but known vars render
			shouldErr: true,
			errCount:  2, // badfilter, another_bad_filter
		},
		{
			name:      "for loop with unresolved collection skips gracefully",
			template:  "Before {% for item in missing.collection %}{{ item }}{% endfor %} After",
			data:      map[string]interface{}{},
			expected:  "Before  After",
			shouldErr: false, // Unresolved collections don't generate errors, just skip loop
			errCount:  0,
		},
		{
			name:      "unknown template tag is skipped",
			template:  "Before {% unknown_tag %} After",
			data:      map[string]interface{}{},
			expected:  "Before  After",
			shouldErr: true,
			errCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewTemplateContext(tt.data, engine)
			result, errs := engine.Render(tt.template, ctx)

			if tt.shouldErr && len(errs) == 0 {
				t.Errorf("expected errors but got none")
			}
			if !tt.shouldErr && len(errs) > 0 {
				t.Errorf("expected no errors but got: %v", errs)
			}
			if tt.errCount > 0 && len(errs) != tt.errCount {
				t.Errorf("expected %d errors, got %d: %v", tt.errCount, len(errs), errs)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
