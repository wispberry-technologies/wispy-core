package tpl

// PopulateRenderStateFromTemplateData extracts styling and scripting information from template data
// and populates it into a RenderState object.
func PopulateRenderStateFromTemplateData(rs RenderState, data TemplateData) {
	// Set the title from the template data
	rs.SetHeadTitle(data.Title)

	// Set the body content if it exists in TemplateData.Content
	if data.Content != "" {
		rs.SetBody(string(data.Content))
	}

	// Extract any styles/scripts from the template data
	if data.Data != nil {
		// Handle both with and without prefix for backward compatibility
		styleKeys := []string{"styles", "__styles"}
		scriptKeys := []string{"scripts", "__scripts"}
		inlineCSSKeys := []string{"inlineCSS", "__inlineCSS"}
		inlineJSKeys := []string{"inlineJS", "__inlineJS"}

		// Check for styles
		for _, key := range styleKeys {
			if styles, ok := data.Data[key].([]string); ok {
				for _, src := range styles {
					rs.AddStyles(StyleAsset{Src: src})
				}
			} else if styleAssets, ok := data.Data[key].([]StyleAsset); ok {
				for _, style := range styleAssets {
					rs.AddStyles(style)
				}
			}
		}

		// Check for scripts
		for _, key := range scriptKeys {
			if scripts, ok := data.Data[key].([]string); ok {
				for _, src := range scripts {
					rs.AddScripts(ScriptAsset{Src: src, Defer: true})
				}
			} else if scriptAssets, ok := data.Data[key].([]ScriptAsset); ok {
				for _, script := range scriptAssets {
					rs.AddScripts(script)
				}
			}
		}

		// Check for inline CSS
		for _, key := range inlineCSSKeys {
			if css, ok := data.Data[key].(string); ok && css != "" {
				// Append to existing CSS if any
				if existing := rs.GetHeadInlineCSS(); existing != "" {
					rs.AddHeadInlineCSS(existing + "\n" + css)
				} else {
					rs.AddHeadInlineCSS(css)
				}
			}
		}

		// Check for inline JS
		for _, key := range inlineJSKeys {
			if js, ok := data.Data[key].(string); ok && js != "" {
				// Append to existing JS if any
				if existing := rs.GetHeadInlineJS(); existing != "" {
					rs.AddHeadInlineJS(existing + "\n" + js)
				} else {
					rs.AddHeadInlineJS(js)
				}
			}
		}
	}
}
