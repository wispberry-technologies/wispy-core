package html

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"wispy-core/pkg/common"
	"wispy-core/pkg/models"
	"wispy-core/pkg/template"
	"wispy-core/pkg/wispytail"
)

var TRI *wispytail.Trie

func init() {
	common.Info("[WISPY-TAIL]")
	trieStart := time.Now()
	TRI = wispytail.BuildFullTrie()
	common.Info(" - Trie Built In: %s", time.Since(trieStart))
	common.Info(" ----------- ")
}

// RenderPageWithLayout renders a page using a layout template
func RenderPageWithLayout(w http.ResponseWriter, r *http.Request, page *models.Page, siteInstance *models.SiteInstance, user *models.UserContext, data map[string]interface{}) {
	// Create template context
	engine, ctx := template.NewSiteTemplateEngine(data, r, siteInstance, page, user)

	// Merge data into context
	for k, v := range data {
		ctx.Data[k] = v
	}

	// Determine layout to use
	layoutName := page.LayoutName
	if layoutName == "" {
		layoutName = "default"
	}

	// Get file paths
	layoutPath := fmt.Sprintf("%s/layouts/%s.html", siteInstance.BasePath, layoutName)
	pagePath := fmt.Sprintf("%s/pages/%s", siteInstance.BasePath, page.FilePath)

	// Read file content
	layoutContent, err := os.ReadFile(layoutPath)
	if err != nil {
		common.Error("Error loading layout %s: %v", layoutName, err)
		http.Error(w, "Error loading layout", http.StatusInternalServerError)
		return
	}
	pageContent, err := os.ReadFile(pagePath)
	if err != nil {
		common.Error("Error loading page %s: %v", page.FilePath, err)
		http.Error(w, "Error loading page", http.StatusInternalServerError)
		return
	}

	// First, process the page content to extract defined blocks
	_, pageErrs := engine.Render(string(pageContent), ctx)
	if len(pageErrs) > 0 {
		for _, err := range pageErrs {
			common.Error("Processing page %s: %v", page.FilePath, err)
		}
		// Handle errors gracefully & silently
		// http.Error(w, "Error processing page", http.StatusInternalServerError)
		// return
	}

	// Now render the layout with the blocks from the page available
	result, layoutErrs := engine.Render(string(layoutContent), ctx)
	if len(layoutErrs) > 0 {
		for _, err := range layoutErrs {
			common.Error("Render: Slug(%s): %v", page.Slug, err)
		}
		// Handle errors gracefully & silently
		// http.Error(w, "Error rendering template", http.StatusInternalServerError)
		// return
	}

	// Process CSS if needed (wispy-tail)
	if siteInstance.CssProcessor == "wispy-tail" {
		// Compile Tailwind CSS if configured
		common.Info("[WISPY-TAIL]")

		extractTime := time.Now()
		// Extract unique class names from the HTML.
		classes := wispytail.ExtractClasses(result)
		common.Info(" - Extract In: %s", time.Since(extractTime))

		generationTime := time.Now()
		// Generate CSS rules for the extracted classes with theme and base layers.
		css := wispytail.GenerateFullCSS(classes, nil, TRI)
		common.Info(" - Generated In: %s", time.Since(generationTime))

		ctx.InternalContext.HtmlDocumentTags = append(ctx.InternalContext.HtmlDocumentTags, models.HtmlDocumentTags{
			TagType:    "style",
			TagName:    "style",
			Location:   "head",
			Contents:   css,
			Priority:   10,
			Attributes: map[string]string{},
		})
	}

	// Build HTML document
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	htmlDocument := &models.ConstructHTMLDocument{
		Body:         result,
		Lang:         page.Lang,
		Title:        page.Title,
		DocumentTags: ctx.InternalContext.HtmlDocumentTags,
		MetaTags:     buildMetaTags(ctx, page),
	}

	// Write HTML document
	WriteHTMLDocument(w, htmlDocument)
}

// buildMetaTags creates meta tags for the HTML document
// Similar to constructMetaTags but with a different name to avoid conflicts
func buildMetaTags(ctx *models.TemplateContext, page *models.Page) []models.HtmlMetaTag {
	var metaTags []models.HtmlMetaTag
	var hasSetViewport bool
	var hasSetCharset bool
	var hasSetTitle bool

	for _, tag := range append(page.MetaTags, ctx.InternalContext.MetaTags...) {
		if tag.Name == "viewport" {
			hasSetViewport = true
		}
		if tag.Name == "charset" {
			hasSetCharset = true
		}
		if tag.Name == "title" {
			hasSetTitle = true
		}

		// Check for custom attributes and add them to the meta tag
		customAttrs := make(map[string]string)
		for attr, value := range tag.CustomAttr {
			customAttrs[attr] = value
		}

		metaTags = append(metaTags, models.HtmlMetaTag{
			Name:       tag.Name,
			Content:    tag.Content,
			Property:   tag.Property,
			HttpEquiv:  tag.HttpEquiv,
			Charset:    tag.Charset,
			CustomAttr: customAttrs,
		})
	}

	if !hasSetViewport {
		metaTags = append(metaTags, models.HtmlMetaTag{
			Name:    "viewport",
			Content: "width=device-width, initial-scale=1",
		})
	}
	if !hasSetCharset {
		metaTags = append(metaTags, models.HtmlMetaTag{
			Name:    "charset",
			Content: "UTF-8",
		})
	}
	if !hasSetTitle {
		metaTags = append(metaTags, models.HtmlMetaTag{
			Name:    "title",
			Content: "Untitled Document",
		})
	}

	return metaTags
}
