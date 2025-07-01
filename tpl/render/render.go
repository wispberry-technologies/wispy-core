package render

import (
	"bytes"
	"fmt"
	"html/template"
	"wispy-core/core"
	"wispy-core/tpl"
	"wispy-core/wispytail"
)

// RenderTemplateWithCSS renders a template and generates CSS from extracted classes
func RenderTemplateWithCSS(eng tpl.TemplateEngine, tmpl *template.Template, themeConfig wispytail.ThemeConf, data map[string]interface{}) (*core.RenderResult, error) {

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	htmlContent := buf.String()

	// Generate CSS for the extracted classes
	css := wispytail.GenerateWithBaseTheme(htmlContent, themeConfig, eng.GetWispyTailTrie())

	return &core.RenderResult{
		HTML: htmlContent,
		CSS:  css,
	}, nil
}
