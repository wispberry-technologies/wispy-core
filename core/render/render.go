package render

import (
	"bytes"
	"fmt"
	"wispy-core/core"
	"wispy-core/tpl"
	"wispy-core/wispytail"
)

// RenderTemplateWithCSS renders a template and generates CSS from extracted classes
func RenderTemplateWithCSS(te core.SiteTplEngine, templatePath string, themeConfig wispytail.ThemeConf, data tpl.TemplateData) (*tpl.RenderResult, error) {
	tmpl, err := te.LoadTemplate(templatePath)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}
	htmlContent := buf.String()

	// Generate CSS for the extracted classes
	css := wispytail.GenerateWithBaseTheme(htmlContent, themeConfig, te.GetTrie())

	return &tpl.RenderResult{
		HTML: htmlContent,
		CSS:  css,
	}, nil
}
