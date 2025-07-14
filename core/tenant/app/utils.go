package app

import (
	"os"
	"path/filepath"
	"wispy-core/common"
	"wispy-core/tpl"
	"wispy-core/wispytail"
)

// renderCMSTemplate is a helper function to render CMS templates with consistent styling
func renderCMSTemplate(engine tpl.TemplateEngine, pagePath, layoutPath string, data tpl.TemplateData, theme string) (tpl.RenderState, error) {
	var state tpl.RenderState

	// Render the template with layout
	state, err := engine.RenderWithLayout(pagePath, layoutPath, data)
	if err != nil {
		common.Error("Failed to render template %s with layout %s: %v", pagePath, layoutPath, err)
		return state, err
	}

	// Generate theme CSS
	themeConfig := wispytail.DefaultThemeConfig()
	trie := engine.GetWispyTailTrie()

	var themeCss = wispytail.DefaultCssTheme
	themePath := filepath.Join("_data", "design", "systems", "themes", theme+".css")
	themeBytes, err := os.ReadFile(themePath)
	if err != nil {
		common.Error("Failed to read theme file %s: %v", themePath, err)
		// Continue without theme CSS rather than failing
	} else {
		themeCss += string(themeBytes)
	}

	baseTwCss := wispytail.GenerateThemeLayer(themeConfig)
	css := wispytail.Generate(state.GetBody(), themeConfig, trie)

	state.AddHeadInlineCSS(themeCss + "\n" + baseTwCss + "\n" + css)

	return state, nil
}
