package core

import (
	"os"
	"testing"
)

func TestTemplateEngine_RealHTMLFiles(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := NewTemplateContext(map[string]interface{}{
		"Page": map[string]interface{}{
			"Meta": map[string]interface{}{
				"Title":       "Test Page",
				"Description": "A test page",
				"Author":      "Test Author",
				"CustomData": map[string]interface{}{
					"features_title":       "Amazing Features",
					"features_description": "Check out these great features",
				},
			},
		},
		"Site": map[string]interface{}{
			"Name": "Test Site",
		},
	}, engine)

	// Test rendering the about page - first process it to define blocks
	aboutPath := "/Users/theo/Desktop/wispy-core/sites/localhost/pages/about.html"
	if content, err := os.ReadFile(aboutPath); err == nil {
		htmlContent := string(content)

		// First render processes the {% define %} tags
		_, errs := engine.Render(htmlContent, ctx)
		if len(errs) > 0 {
			t.Errorf("errors processing about page: %v", errs)
		}

		// Now render the defined block
		blockTemplate := `{% block "page-body" %}`
		result, errs := engine.Render(blockTemplate, ctx)
		if len(errs) > 0 {
			t.Errorf("errors rendering page-body block: %v", errs)
		}

		if len(result) == 0 {
			t.Error("page-body block rendered to empty string")
		}

		// Check that it contains expected content
		if !contains(result, "About Wispy Core") {
			t.Errorf("page-body block should contain 'About Wispy Core', but got: %q", result)
		}
	} else {
		t.Errorf("could not read about page: %v", err)
	}
}

func TestTemplateEngine_HomePageWithCustomData(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := NewTemplateContext(map[string]interface{}{
		"Page": map[string]interface{}{
			"Meta": map[string]interface{}{
				"CustomData": map[string]interface{}{
					"features_title":       "Amazing Features",
					"features_description": "Check out these great features",
				},
			},
		},
		"Site": map[string]interface{}{
			"Name": "Test Site",
		},
	}, engine)

	// Test rendering the home page
	homePath := "/Users/theo/Desktop/wispy-core/sites/localhost/pages/home.html"
	if content, err := os.ReadFile(homePath); err == nil {
		htmlContent := string(content)

		// First render processes the {% define %} tags
		_, errs := engine.Render(htmlContent, ctx)
		if len(errs) > 0 {
			t.Errorf("errors processing home page: %v", errs)
		}

		// Now render the defined block
		blockTemplate := `{% block "page-body" %}`
		result, errs := engine.Render(blockTemplate, ctx)
		if len(errs) > 0 {
			t.Errorf("errors rendering page-body block: %v", errs)
		}

		if len(result) == 0 {
			t.Error("home page rendered to empty string")
		}

		// Check that variables were interpolated
		if !contains(result, "Amazing Features") {
			t.Errorf("home page should contain interpolated features_title, got: %q", result)
		}

		if !contains(result, "Check out these great features") {
			t.Errorf("home page should contain interpolated features_description, got: %q", result)
		}
	} else {
		t.Errorf("could not read home page: %v", err)
	}
}

func TestTemplateEngine_HeroSection(t *testing.T) {
	engine := NewTemplateEngine(DefaultFunctionMap())
	ctx := NewTemplateContext(map[string]interface{}{
		"Data": map[string]interface{}{
			"hero_title":       "Welcome to Wispy",
			"hero_description": "The fastest CMS in the world",
			"hero_button_text": "Get Started",
			"hero_button_link": "/start",
		},
	}, engine)

	// Test rendering the hero section
	heroPath := "/Users/theo/Desktop/wispy-core/sites/localhost/templates/sections/hero.html"
	if content, err := os.ReadFile(heroPath); err == nil {
		htmlContent := string(content)

		result, errs := engine.Render(htmlContent, ctx)
		if len(errs) > 0 {
			t.Errorf("errors rendering hero section: %v", errs)
		}

		if len(result) == 0 {
			t.Error("hero section rendered to empty string")
		}

		// Check that variables were interpolated
		if !contains(result, "Welcome to Wispy") {
			t.Error("hero section should contain interpolated hero_title")
		}

		if !contains(result, "The fastest CMS in the world") {
			t.Error("hero section should contain interpolated hero_description")
		}

		if !contains(result, "Get Started") {
			t.Error("hero section should contain interpolated hero_button_text")
		}

		if !contains(result, "/start") {
			t.Error("hero section should contain interpolated hero_button_link")
		}
	} else {
		t.Errorf("could not read hero section: %v", err)
	}
}
