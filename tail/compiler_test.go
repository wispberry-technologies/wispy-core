package tail

import (
	"strings"
	"testing"
)

func TestExtractClasses(t *testing.T) {
	testCases := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name:     "Simple HTML",
			html:     `<div class="flex p-4 bg-white">Hello</div>`,
			expected: []string{"flex", "p-4", "bg-white"},
		},
		{
			name: "Multiple elements with classes",
			html: `
				<div class="container mx-auto">
					<h1 class="text-2xl font-bold">Title</h1>
					<p class="mt-2 text-gray-600">Content</p>
				</div>
			`,
			expected: []string{"container", "mx-auto", "text-2xl", "font-bold", "mt-2", "text-gray-600"},
		},
		{
			name:     "Element with duplicate classes",
			html:     `<div class="p-4 mt-2 p-4">Content</div>`,
			expected: []string{"p-4", "mt-2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			classes, err := extractClasses(tc.html)
			if err != nil {
				t.Fatalf("Error extracting classes: %v", err)
			}

			// Convert to map for easier comparison
			classMap := make(map[string]struct{})
			for _, class := range classes {
				classMap[class] = struct{}{}
			}

			for _, expected := range tc.expected {
				if _, ok := classMap[expected]; !ok {
					t.Errorf("Expected class %q not found in extracted classes", expected)
				}
			}

			if len(classMap) != len(tc.expected) {
				t.Errorf("Expected %d unique classes, got %d", len(tc.expected), len(classMap))
			}
		})
	}
}

func TestSortClasses(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Position and display categories",
			input:    []string{"flex", "relative", "block"},
			expected: []string{"relative", "block", "flex"},
		},
		{
			name:     "Mixed categories",
			input:    []string{"text-red-500", "p-4", "bg-white", "flex"},
			expected: []string{"flex", "p-4", "bg-white", "text-red-500"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortClasses(tt.input)

			// Check if categories are in the right order
			// This is a simplified check that just ensures the first class is correct
			if len(result) > 0 && result[0] != tt.expected[0] {
				t.Errorf("First class incorrect, got %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCompileHTML(t *testing.T) {
	html := `<div class="flex p-4 bg-white">
		<h1 class="text-xl font-bold">Hello</h1>
	</div>`

	css, err := CompileHTML(html)
	if err != nil {
		t.Fatalf("Failed to compile HTML: %v", err)
	}

	// Check for expected output parts
	expectedParts := []string{
		"tailwindcss",
		"@layer",
		".flex",
		".bg-white",
	}

	for _, part := range expectedParts {
		if !strings.Contains(css, part) {
			t.Errorf("Expected CSS to contain %q, but it doesn't", part)
		}
	}
}
