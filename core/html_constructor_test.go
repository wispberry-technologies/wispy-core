package core

import (
	"net/http/httptest"
	"strings"
	"testing"
	"wispy-core/models"
)

func TestWriteHtmlDocumentTags(t *testing.T) {
	tests := []struct {
		name     string
		tags     []models.HtmlDocumentTags
		expected string
	}{
		{
			name: "link tag with attributes",
			tags: []models.HtmlDocumentTags{
				{
					TagType:  "link",
					TagName:  "link",
					Location: "head",
					Contents: "",
					Priority: 15,
					Attributes: map[string]string{
						"rel":  "stylesheet",
						"href": "/assets/css/test.css",
						"type": "text/css",
					},
					SelfClosing: true,
				},
			},
			expected: `<link rel="stylesheet" href="/assets/css/test.css" type="text/css" />`,
		},
		{
			name: "script tag with attributes and content",
			tags: []models.HtmlDocumentTags{
				{
					TagType:  "script",
					TagName:  "script",
					Location: "head",
					Contents: "console.log('test');",
					Priority: 20,
					Attributes: map[string]string{
						"type": "text/javascript",
					},
					SelfClosing: false,
				},
			},
			expected: `<script type="text/javascript">
console.log('test');
</script>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			writeHtmlDocumentTags(recorder, tt.tags)

			result := strings.TrimSpace(recorder.Body.String())
			expected := strings.TrimSpace(tt.expected)

			if result != expected {
				t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, result)
			}
		})
	}
}
