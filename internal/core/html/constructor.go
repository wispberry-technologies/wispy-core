// Package html provides functions for HTML construction and rendering
package html

import (
	"net/http"
	"wispy-core/pkg/models"
)

// WriteString writes a string to the HTTP response
func WriteString(w http.ResponseWriter, s string) {
	if _, err := w.Write([]byte(s)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// WriteHTMLDocument writes a complete HTML document to the response writer
func WriteHTMLDocument(w http.ResponseWriter, doc *models.ConstructHTMLDocument) {
	var lang = "en"
	if doc.Lang != "" {
		lang = doc.Lang
	}

	WriteString(w, "<!DOCTYPE html>\n<html lang=\"")
	WriteString(w, lang)
	WriteString(w, "\">\n<head>\n")
	WriteString(w, "<title>")
	WriteString(w, doc.Title)
	WriteString(w, "</title>\n")
	//  Write meta tags
	for _, meta := range doc.MetaTags {
		WriteString(w, "<meta name=\"")
		WriteString(w, meta.Name)
		WriteString(w, "\" content=\"")
		WriteString(w, meta.Content)
		WriteString(w, "\">\n")
	}
	// Write additional meta tags
	WriteString(w, "<!-- Document Tags -->")
	writeHtmlDocumentTags(w, doc.DocumentTags)
	//
	WriteString(w, "</head>\n<body>\n")
	WriteString(w, doc.Body)
	WriteString(w, "\n</body>\n")
	// Close HTML
	WriteString(w, "</html>")
}

// ConstructMetaTags generates HTML meta tags from page metadata
// This is a pure function that transforms input to output without side effects
func ConstructMetaTags(ctx *models.TemplateContext, page *models.Page) []models.HtmlMetaTag {
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

// writeHtmlDocumentTags writes HTML document tags to the response writer
func writeHtmlDocumentTags(w http.ResponseWriter, tags []models.HtmlDocumentTags) {
	for _, tag := range tags {
		WriteString(w, "<")
		WriteString(w, tag.TagName)

		// Add attributes
		for attrName, attrValue := range tag.Attributes {
			WriteString(w, " ")
			WriteString(w, attrName)
			WriteString(w, "=\"")
			WriteString(w, attrValue)
			WriteString(w, "\"")
		}

		if tag.SelfClosing {
			WriteString(w, " />\n")
		} else {
			WriteString(w, ">\n")
			if tag.Contents != "" {
				WriteString(w, tag.Contents)
			}
			WriteString(w, "</")
			WriteString(w, tag.TagName)
			WriteString(w, ">\n")
		}
	}
}
