package core

import (
	"reflect"
	"strings"
	"testing"
	"wispy-core/models"
	"wispy-core/tests"
)

func TestParseHTMLCommentMetadata(t *testing.T) {
	testCases := []struct {
		name     string
		metadata string
		expect   *models.Page
	}{
		{
			name: "basic metadata",
			metadata: `
@name home.html
@url /
@author Wispy Core Team
@layout default
@is_draft false
@require_auth false
@required_roles []
`,
			expect: &models.Page{
				Title:         "home",
				URL:           "/",
				Author:        "Wispy Core Team",
				Layout:        "default",
				IsDraft:       false,
				RequireAuth:   false,
				RequiredRoles: []string{},
				CustomData:    make(map[string]string),
			},
		},
		{
			name: "metadata with custom data",
			metadata: `
@name about.html
@url /about
@author Wispy Core Team
@layout default
@hero_title Welcome to Wispy Core
@hero_description A modern, multisite content management system built with Go.
`,
			expect: &models.Page{
				Title:  "about",
				URL:    "/about",
				Author: "Wispy Core Team",
				Layout: "default",
				CustomData: map[string]string{
					"hero_title":       "Welcome to Wispy Core",
					"hero_description": "A modern, multisite content management system built with Go.",
				},
			},
		},
		{
			name: "boolean flags without values",
			metadata: `
@name draft-page.html
@url /draft
@author Wispy Core Team
@layout default
@is_draft
@require_auth
`,
			expect: &models.Page{
				Title:       "draft-page",
				URL:         "/draft",
				Author:      "Wispy Core Team",
				Layout:      "default",
				IsDraft:     true,
				RequireAuth: true,
				CustomData:  make(map[string]string),
			},
		},
		{
			name: "with required roles",
			metadata: `
@name admin.html
@url /admin
@author Wispy Core Team
@layout admin
@require_auth true
@required_roles ["admin", "editor"]
`,
			expect: &models.Page{
				Title:         "admin",
				URL:           "/admin",
				Author:        "Wispy Core Team",
				Layout:        "admin",
				RequireAuth:   true,
				RequiredRoles: []string{"admin", "editor"},
				CustomData:    make(map[string]string),
			},
		},
		{
			name:     "empty metadata",
			metadata: "",
			expect: &models.Page{
				CustomData: make(map[string]string),
			},
		},
	}

	for _, tc := range testCases {
		page := &models.Page{
			CustomData: make(map[string]string),
		}
		err := ParseHTMLCommentMetadata(tc.metadata, page)

		if err != nil {
			t.Error(tests.LogFail(tc.name))
			t.Log(tests.LogWarn("unexpected error: %v", err))
			continue
		}

		// Compare relevant fields (ignoring Site, CreatedAt, UpdatedAt, Content)
		fieldsMatch := true

		if page.Title != tc.expect.Title {
			fieldsMatch = false
			t.Log(tests.LogWarn("Title: expected '%s', got '%s'", tc.expect.Title, page.Title))
		}

		if page.URL != tc.expect.URL {
			fieldsMatch = false
			t.Log(tests.LogWarn("URL: expected '%s', got '%s'", tc.expect.URL, page.URL))
		}

		if page.Author != tc.expect.Author {
			fieldsMatch = false
			t.Log(tests.LogWarn("Author: expected '%s', got '%s'", tc.expect.Author, page.Author))
		}

		if page.Layout != tc.expect.Layout {
			fieldsMatch = false
			t.Log(tests.LogWarn("Layout: expected '%s', got '%s'", tc.expect.Layout, page.Layout))
		}

		if page.IsDraft != tc.expect.IsDraft {
			fieldsMatch = false
			t.Log(tests.LogWarn("IsDraft: expected '%t', got '%t'", tc.expect.IsDraft, page.IsDraft))
		}

		if page.RequireAuth != tc.expect.RequireAuth {
			fieldsMatch = false
			t.Log(tests.LogWarn("RequireAuth: expected '%t', got '%t'", tc.expect.RequireAuth, page.RequireAuth))
		}

		if !reflect.DeepEqual(page.RequiredRoles, tc.expect.RequiredRoles) {
			fieldsMatch = false
			t.Log(tests.LogWarn("RequiredRoles: expected %v, got %v", tc.expect.RequiredRoles, page.RequiredRoles))
		}

		if !reflect.DeepEqual(page.CustomData, tc.expect.CustomData) {
			fieldsMatch = false
			t.Log(tests.LogWarn("CustomData: expected %v, got %v", tc.expect.CustomData, page.CustomData))
		}

		if fieldsMatch {
			t.Log(tests.LogPass(tc.name))
		} else {
			t.Error(tests.LogFail(tc.name))
		}
	}
}

func TestParsePageHTML(t *testing.T) {
	testCases := []struct {
		name    string
		content string
		expect  *models.Page
	}{
		{
			name: "basic page",
			content: `<!--
@name home.html
@url /
@author Wispy Core Team
@layout default
@is_draft false
@require_auth false
@required_roles []
-->
<div class="hero min-h-screen bg-base-200">
    <div class="hero-content text-center">
        <div class="max-w-4xl">
            <h1 class="text-5xl font-bold mb-8">Welcome to Wispy Core</h1>
            <p class="text-xl mb-8">A modern, multisite content management system built with Go.</p>
        </div>
    </div>
</div>`,
			expect: &models.Page{
				Title:         "home",
				URL:           "/",
				Author:        "Wispy Core Team",
				Layout:        "default",
				IsDraft:       false,
				RequireAuth:   false,
				RequiredRoles: []string{},
				CustomData:    make(map[string]string),
				IsStatic:      true,
			},
		},
		{
			name: "page with custom data",
			content: `<!--
@name about.html
@url /about
@author Wispy Core Team
@layout default
@hero_title Welcome to Wispy Core
@hero_description A modern, multisite content management system built with Go.
-->
<div class="container mx-auto px-4 py-8">
    <h1>About Us</h1>
    <p>Content goes here</p>
</div>`,
			expect: &models.Page{
				Title:  "about",
				URL:    "/about",
				Author: "Wispy Core Team",
				Layout: "default",
				CustomData: map[string]string{
					"hero_title":       "Welcome to Wispy Core",
					"hero_description": "A modern, multisite content management system built with Go.",
				},
				IsStatic: true,
			},
		},
		{
			name: "page with multiple HTML comments",
			content: `<!--
@name contact.html
@url /contact
@author Wispy Core Team
@layout default
-->
<div class="container mx-auto px-4 py-8">
    <h1>Contact Us</h1>
    <!-- This is a regular comment that should be removed -->
    <p>Content goes here</p>
    <!-- 
    Another multi-line comment
    that should be removed
    -->
    <p>More content</p>
</div>`,
			expect: &models.Page{
				Title:      "contact",
				URL:        "/contact",
				Author:     "Wispy Core Team",
				Layout:     "default",
				CustomData: make(map[string]string),
				IsStatic:   true,
			},
		},
	}

	for _, tc := range testCases {
		// Create a site instance for testing
		siteInstance := &models.SiteInstance{
			Site: &models.Site{
				Domain:   "test.com",
				Name:     "Test Site",
				IsActive: true,
			},
		}

		page, err := ParsePageHTML(siteInstance, tc.content)

		if err != nil {
			t.Error(tests.LogFail(tc.name))
			t.Log(tests.LogWarn("unexpected error: %v", err))
			continue
		}

		// Compare relevant fields (ignoring CreatedAt, UpdatedAt, Content)
		fieldsMatch := true

		if page.Title != tc.expect.Title {
			fieldsMatch = false
			t.Log(tests.LogWarn("Title: expected '%s', got '%s'", tc.expect.Title, page.Title))
		}

		if page.URL != tc.expect.URL {
			fieldsMatch = false
			t.Log(tests.LogWarn("URL: expected '%s', got '%s'", tc.expect.URL, page.URL))
		}

		if page.Author != tc.expect.Author {
			fieldsMatch = false
			t.Log(tests.LogWarn("Author: expected '%s', got '%s'", tc.expect.Author, page.Author))
		}

		if page.Layout != tc.expect.Layout {
			fieldsMatch = false
			t.Log(tests.LogWarn("Layout: expected '%s', got '%s'", tc.expect.Layout, page.Layout))
		}

		if page.IsDraft != tc.expect.IsDraft {
			fieldsMatch = false
			t.Log(tests.LogWarn("IsDraft: expected '%t', got '%t'", tc.expect.IsDraft, page.IsDraft))
		}

		if page.RequireAuth != tc.expect.RequireAuth {
			fieldsMatch = false
			t.Log(tests.LogWarn("RequireAuth: expected '%t', got '%t'", tc.expect.RequireAuth, page.RequireAuth))
		}

		if !reflect.DeepEqual(page.RequiredRoles, tc.expect.RequiredRoles) {
			fieldsMatch = false
			t.Log(tests.LogWarn("RequiredRoles: expected %v, got %v", tc.expect.RequiredRoles, page.RequiredRoles))
		}

		if !reflect.DeepEqual(page.CustomData, tc.expect.CustomData) {
			fieldsMatch = false
			t.Log(tests.LogWarn("CustomData: expected %v, got %v", tc.expect.CustomData, page.CustomData))
		}

		if page.IsStatic != tc.expect.IsStatic {
			fieldsMatch = false
			t.Log(tests.LogWarn("IsStatic: expected '%t', got '%t'", tc.expect.IsStatic, page.IsStatic))
		}

		// Check that content was processed and contains define tags
		if !strings.Contains(page.Content, "define \"page-content\"") {
			fieldsMatch = false
			t.Log(tests.LogWarn("Content: missing 'define \"page-content\"' tag"))
		}

		if fieldsMatch {
			t.Log(tests.LogPass(tc.name))
		} else {
			t.Error(tests.LogFail(tc.name))
		}
	}
}

func TestParsePageHTML_ErrorCases(t *testing.T) {
	testCases := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name: "malformed metadata comment - no closing tag",
			content: `<!--
@name home.html
@url /
@author Wispy Core Team
<div class="hero min-h-screen bg-base-200">
    <div class="hero-content text-center">
        <div class="max-w-4xl">
            <h1>Welcome to Wispy Core</h1>
        </div>
    </div>
</div>`,
			expectError: true,
		},
		{
			name: "no metadata comment",
			content: `<div class="hero min-h-screen bg-base-200">
    <div class="hero-content text-center">
        <div class="max-w-4xl">
            <h1>Welcome to Wispy Core</h1>
        </div>
    </div>
</div>`,
			expectError: false, // Function doesn't return error, just logs a warning
		},
	}

	for _, tc := range testCases {
		siteInstance := &models.SiteInstance{
			Site: &models.Site{
				Domain:   "test.com",
				Name:     "Test Site",
				IsActive: true,
			},
		}

		_, err := ParsePageHTML(siteInstance, tc.content)

		if tc.expectError && err == nil {
			t.Error(tests.LogFail(tc.name))
			t.Log(tests.LogWarn("expected error but got none"))
		} else if !tc.expectError && err != nil {
			t.Error(tests.LogFail(tc.name))
			t.Log(tests.LogWarn("unexpected error: %v", err))
		} else {
			t.Log(tests.LogPass(tc.name))
		}
	}
}
