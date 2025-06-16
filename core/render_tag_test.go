package core

import (
	"net/http"
	"strings"
	"testing"
	"wispy-core/models"
)

func TestRenderTagWithAppAndMarketingTemplates(t *testing.T) {
	// Create a test template engine
	engine := NewTemplateEngine(DefaultFunctionMap())

	// Create a mock request
	req, _ := http.NewRequest("GET", "http://localhost/test", nil)

	// Test @app/* template rendering
	t.Run("Render @app template", func(t *testing.T) {
		template := `{% render "@app/login-form" %}`

		ctx := &models.TemplateContext{
			Engine:  engine,
			Request: req,
			Data:    make(map[string]interface{}),
			InternalContext: &models.InternalContext{
				TemplatesCache: make(map[string]string),
			},
		}

		result, errs := engine.Render(template, ctx)

		if len(errs) > 0 {
			t.Errorf("Expected no errors, got: %v", errs)
		}

		if !strings.Contains(result, "Welcome Back") {
			t.Errorf("Expected template content to contain 'Welcome Back', got: %s", result)
		}
	})

	// Test @marketing/* template rendering
	t.Run("Render @marketing template", func(t *testing.T) {
		template := `{% render "@marketing/hero-banner" %}`

		ctx := &models.TemplateContext{
			Engine:  engine,
			Request: req,
			Data:    make(map[string]interface{}),
			InternalContext: &models.InternalContext{
				TemplatesCache: make(map[string]string),
			},
		}

		result, errs := engine.Render(template, ctx)

		if len(errs) > 0 {
			t.Errorf("Expected no errors, got: %v", errs)
		}

		if !strings.Contains(result, "hero-banner") {
			t.Errorf("Expected template content to contain 'hero-banner', got: %s", result)
		}
	})

	// Test template with context data
	t.Run("Render @app template with context", func(t *testing.T) {
		template := `{% render "@app/user-profile" %}`

		ctx := &models.TemplateContext{
			Engine:  engine,
			Request: req,
			Data: map[string]interface{}{
				"user": map[string]interface{}{
					"name":   "John Doe",
					"email":  "john@example.com",
					"avatar": "/images/john.jpg",
					"role":   "Admin",
				},
			},
			InternalContext: &models.InternalContext{
				TemplatesCache: make(map[string]string),
			},
		}

		result, errs := engine.Render(template, ctx)

		if len(errs) > 0 {
			t.Errorf("Expected no errors, got: %v", errs)
		}

		if !strings.Contains(result, "John Doe") {
			t.Errorf("Expected template to render user name 'John Doe', got: %s", result)
		}

		if !strings.Contains(result, "john@example.com") {
			t.Errorf("Expected template to render user email 'john@example.com', got: %s", result)
		}
	})

	// Test caching behavior
	t.Run("Test template caching", func(t *testing.T) {
		template := `{% render "@app/login-form" %}`

		ctx := &models.TemplateContext{
			Engine:  engine,
			Request: req,
			Data:    make(map[string]interface{}),
			InternalContext: &models.InternalContext{
				TemplatesCache: make(map[string]string),
			},
		}

		// First render - should load from file
		result1, errs1 := engine.Render(template, ctx)
		if len(errs1) > 0 {
			t.Errorf("Expected no errors on first render, got: %v", errs1)
		}

		// Check if template was cached
		if _, exists := ctx.InternalContext.TemplatesCache["@app/login-form"]; !exists {
			t.Errorf("Expected template to be cached after first render")
		}

		// Second render - should use cache
		result2, errs2 := engine.Render(template, ctx)
		if len(errs2) > 0 {
			t.Errorf("Expected no errors on second render, got: %v", errs2)
		}

		if result1 != result2 {
			t.Errorf("Expected same result from cached template")
		}
	})

	// Test error handling for non-existent template
	t.Run("Handle non-existent template", func(t *testing.T) {
		template := `{% render "@app/non-existent" %}`

		ctx := &models.TemplateContext{
			Engine:  engine,
			Request: req,
			Data:    make(map[string]interface{}),
			InternalContext: &models.InternalContext{
				TemplatesCache: make(map[string]string),
			},
		}

		_, errs := engine.Render(template, ctx)

		if len(errs) == 0 {
			t.Errorf("Expected error for non-existent template")
		}

		if !strings.Contains(errs[0].Error(), "template not found") {
			t.Errorf("Expected 'template not found' error, got: %v", errs[0])
		}
	})
}
