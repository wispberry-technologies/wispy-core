package core

import (
	"testing"
	"wispy-core/models"
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

func TestTemplateEngine_AssetCSS(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())

	tests := []struct {
		name      string
		template  string
		wantTags  int // Number of expected HTML document tags
		wantError bool
	}{
		{
			name:      "import external CSS",
			template:  `{% asset "css" "public/css/style.css" %}`,
			wantTags:  1,
			wantError: false,
		},
		{
			name:      "import inline CSS",
			template:  `{% asset "css-inline" "assets/css/test.css" %}`,
			wantTags:  1,
			wantError: false,
		},
		{
			name:      "import remote CSS",
			template:  `{% asset "css" "https://cdn.jsdelivr.net/npm/daisyui@5/dist/full.css" %}`,
			wantTags:  1,
			wantError: false,
		},
		{
			name:      "import CSS without type",
			template:  `{% asset %}`,
			wantTags:  0,
			wantError: true,
		},
		{
			name:      "invalid asset type",
			template:  `{% asset "invalid" "public/css/style.css" %}`,
			wantTags:  0,
			wantError: true,
		},
		{
			name:      "invalid path prefix",
			template:  `{% asset "css" "invalid/path/style.css" %}`,
			wantTags:  0,
			wantError: true,
		},
		{
			name:      "inline remote asset (should error)",
			template:  `{% asset "css-inline" "https://example.com/style.css" %}`,
			wantTags:  0,
			wantError: true,
		},
		{
			name:      "import same CSS twice (should only create one tag)",
			template:  `{% asset "css" "public/css/style.css" %}{% asset "css" "public/css/style.css" %}`,
			wantTags:  1,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestTemplateContext(map[string]interface{}{}, engine)

			_, errs := engine.Render(tt.template, ctx)

			if tt.wantError && len(errs) == 0 {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && len(errs) > 0 {
				t.Errorf("Unexpected error: %v", errs)
			}

			if len(ctx.InternalContext.HtmlDocumentTags) != tt.wantTags {
				t.Errorf("Expected %d HTML document tags, got %d", tt.wantTags, len(ctx.InternalContext.HtmlDocumentTags))
			}

			// Check that CSS tags have correct attributes
			if tt.wantTags > 0 && !tt.wantError && len(ctx.InternalContext.HtmlDocumentTags) > 0 {
				tag := ctx.InternalContext.HtmlDocumentTags[0]
				if tag.Location != "head" {
					t.Errorf("Expected CSS tag location to be 'head', got '%s'", tag.Location)
				}
			}
		})
	}
}

func TestTemplateEngine_AssetJS(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())

	tests := []struct {
		name      string
		template  string
		wantTags  int
		wantError bool
	}{
		{
			name:      "import external JS",
			template:  `{% asset "js" "public/js/script.js" %}`,
			wantTags:  1,
			wantError: false,
		},
		{
			name:      "import inline JS",
			template:  `{% asset "js-inline" "assets/js/test.js" %}`,
			wantTags:  1,
			wantError: false,
		},
		{
			name:      "import JS with location",
			template:  `{% asset "js" "public/js/script.js" location="pre-footer" %}`,
			wantTags:  1,
			wantError: false,
		},
		{
			name:      "import remote JS",
			template:  `{% asset "js" "https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js" %}`,
			wantTags:  1,
			wantError: false,
		},
		{
			name:      "import JS without path",
			template:  `{% asset "js" %}`,
			wantTags:  0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestTemplateContext(map[string]interface{}{}, engine)

			_, errs := engine.Render(tt.template, ctx)

			if tt.wantError && len(errs) == 0 {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantError && len(errs) > 0 {
				t.Errorf("Unexpected error: %v", errs)
			}

			if len(ctx.InternalContext.HtmlDocumentTags) != tt.wantTags {
				t.Errorf("Expected %d HTML document tags, got %d", tt.wantTags, len(ctx.InternalContext.HtmlDocumentTags))
			}

			// Check JS tag attributes
			if tt.wantTags > 0 && !tt.wantError && len(ctx.InternalContext.HtmlDocumentTags) > 0 {
				tag := ctx.InternalContext.HtmlDocumentTags[0]
				if tag.TagName != "script" {
					t.Errorf("Expected JS tag name to be 'script', got '%s'", tag.TagName)
				}
			}
		})
	}
}

func TestTemplateEngine_AssetResourceDeduplication(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := newTestTemplateContext(map[string]interface{}{}, engine)

	// Import same CSS file twice
	template := `{% asset "css" "public/css/style.css" %}{% asset "css" "public/css/style.css" %}`
	_, errs := engine.Render(template, ctx)

	if len(errs) > 0 {
		t.Errorf("Unexpected error: %v", errs)
	}

	// Should only have one tag
	if len(ctx.InternalContext.HtmlDocumentTags) != 1 {
		t.Errorf("Expected 1 HTML document tag after deduplication, got %d", len(ctx.InternalContext.HtmlDocumentTags))
	}

	// Check ImportedResources map
	if len(ctx.InternalContext.ImportedResources) != 1 {
		t.Errorf("Expected 1 imported resource, got %d", len(ctx.InternalContext.ImportedResources))
	}

	expectedKey := "public/css/style.css|css"
	if ctx.InternalContext.ImportedResources[expectedKey] != "css" {
		t.Errorf("Expected imported resource type to be 'css', got '%s'", ctx.InternalContext.ImportedResources[expectedKey])
	}
}

func TestTemplateEngine_AssetResourceConflictDetection(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := newTestTemplateContext(map[string]interface{}{}, engine)

	// Try to import same file as external then inline (should error)
	template := `{% asset "css" "public/css/style.css" %}{% asset "css-inline" "public/css/style.css" %}`
	_, errs := engine.Render(template, ctx)

	if len(errs) == 0 {
		t.Errorf("Expected error when importing same file with different types")
	}

	// Should only have one tag from the first import
	if len(ctx.InternalContext.HtmlDocumentTags) != 1 {
		t.Errorf("Expected 1 HTML document tag after conflict, got %d", len(ctx.InternalContext.HtmlDocumentTags))
	}
}

func TestTemplateEngine_AssetErrorHandling(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())

	// Create mock page context with localhost domain
	mockPage := &models.Page{
		SiteDetails: models.SiteDetails{
			Domain: "localhost",
		},
	}

	tests := []struct {
		name             string
		template         string
		expectedContains string
		shouldHaveErrors bool
	}{
		{
			name:             "missing inline CSS file continues rendering",
			template:         `Before asset {% asset "css-inline" "assets/css/nonexistent.css" %} After asset`,
			expectedContains: "Before asset  After asset",
			shouldHaveErrors: true,
		},
		{
			name:             "missing inline JS file continues rendering",
			template:         `Before asset {% asset "js-inline" "assets/js/nonexistent.js" %} After asset`,
			expectedContains: "Before asset  After asset",
			shouldHaveErrors: true,
		},
		{
			name:             "invalid asset type continues rendering",
			template:         `Before asset {% asset "invalid" "assets/css/test.css" %} After asset`,
			expectedContains: "Before asset  After asset",
			shouldHaveErrors: true,
		},
		{
			name:             "invalid path continues rendering",
			template:         `Before asset {% asset "css" "../../../etc/passwd" %} After asset`,
			expectedContains: "Before asset  After asset",
			shouldHaveErrors: true,
		},
		{
			name:             "missing external file continues rendering",
			template:         `Before asset {% asset "css" "public/css/nonexistent.css" %} After asset`,
			expectedContains: "Before asset  After asset",
			shouldHaveErrors: false, // External files don't check existence during render
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize context with fresh ImportedResources map
			ctx := NewTemplateContext(map[string]interface{}{
				"Page": mockPage,
			}, engine)
			ctx.InternalContext.ImportedResources = make(map[string]string)

			result, errs := engine.Render(tt.template, ctx)

			// Check that the template continues rendering
			if result != tt.expectedContains {
				t.Errorf("expected result to contain '%s', got '%s'", tt.expectedContains, result)
			}

			// Check error expectation
			if tt.shouldHaveErrors && len(errs) == 0 {
				t.Error("expected errors but got none")
			}
			if !tt.shouldHaveErrors && len(errs) > 0 {
				t.Errorf("expected no errors but got: %v", errs)
			}

			// Verify that errors don't prevent rendering from completing
			if result == "" && tt.expectedContains != "" {
				t.Error("template rendering was completely blocked by errors")
			}
		})
	}
}

func TestTemplateEngine_VerbatimTag(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := NewTemplateContext(map[string]interface{}{
		"name": "World",
	}, engine)

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "verbatim preserves template syntax literally",
			template: `Before {% verbatim %}{{ name }} and {% if true %}content{% endif %}{% endverbatim %} After`,
			expected: `Before {{ name }} and {% if true %}content{% endif %} After`,
		},
		{
			name:     "verbatim with asset tag example",
			template: `Example: {% verbatim %}{% asset "css" "public/css/style.css" %}{% endverbatim %}`,
			expected: `Example: {% asset "css" "public/css/style.css" %}`,
		},
		{
			name:     "empty verbatim",
			template: `Before {% verbatim %}{% endverbatim %} After`,
			expected: `Before  After`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, errs := engine.Render(tt.template, ctx)
			if len(errs) > 0 {
				t.Errorf("unexpected errors: %v", errs)
			}
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// newTestTemplateContext creates a template context with a mock page for testing
func newTestTemplateContext(data map[string]interface{}, engine *models.TemplateEngine) *models.TemplateContext {
	// Create a mock page with site details
	page := &models.Page{
		SiteDetails: models.SiteDetails{
			Domain: "localhost",
			Name:   "Test Site",
		},
	}

	// Merge the page into the data
	if data == nil {
		data = make(map[string]interface{})
	}
	data["Page"] = page

	return NewTemplateContext(data, engine)
}
