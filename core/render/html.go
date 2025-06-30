package render

import (
	"io"
	"net/http"
	"strings"
)

var (
	// Reusable template for HTML escaping
	htmlEscaper = strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&#34;",
		`'`, "&#39;",
	)
)

func HtmlBaseRenderResponse(w http.ResponseWriter, rs RenderState) error {
	return HtmlBaseRender(w, rs)
}

func HtmlBaseRender(w io.Writer, rs RenderState) (err error) {

	// 3. Build HTML document
	writeDocStart(w, rs)
	writeHead(w, rs)
	writeBody(w, rs)
	writeDocEnd(w)

	return nil
}

func writeDocStart(w io.Writer, rs RenderState) {
	w.Write([]byte(`<!DOCTYPE html><html lang="en">`))
}

func writeHead(w io.Writer, rs RenderState) {
	w.Write([]byte(`<head>`))
	w.Write([]byte(`<meta charset="UTF-8">`))
	w.Write([]byte(`<meta name="viewport" content="width=device-width,initial-scale=1">`))

	// Title
	w.Write([]byte(`<title>`))
	htmlEscaper.WriteString(w, rs.GetHeadTitle())
	w.Write([]byte(`</title>`))

	// Stylesheets
	for _, style := range rs.GetHeadStyles() {
		// TODO: Handle style priority if needed
		w.Write([]byte(`<link rel="stylesheet" href="`))
		w.Write([]byte(style.Src))
		if len(style.Attrs) > 0 {
			writeAttributes(w, style.Attrs)
		}
		w.Write([]byte(`>`))
	}
	// Inline CSS
	if inlineCSS := rs.GetHeadInlineCSS(); inlineCSS != "" {
		w.Write([]byte(`<style>`))
		w.Write([]byte(inlineCSS))
		w.Write([]byte(`</style>`))
	}

	// Scripts
	for _, script := range rs.GetHeadScripts() {
		w.Write([]byte(`<script src="`))
		w.Write([]byte(script.Src))
		if script.Async {
			w.Write([]byte(`" async`))
		}
		if script.Defer {
			w.Write([]byte(`" defer`))
		}
		w.Write([]byte(`></script>`))
	}

	// Inline JS
	if inlineJS := rs.GetHeadInlineJS(); inlineJS != "" {
		w.Write([]byte(`<script>`))
		w.Write([]byte(inlineJS))
		w.Write([]byte(`</script>`))
	}

	w.Write([]byte(`</head>`))
}

func writeBody(w io.Writer, rs RenderState) {
	w.Write([]byte(`<body>`))
	// Body content
	w.Write([]byte(rs.Body()))
	w.Write([]byte(`</body>`))
}

func writeDocEnd(w io.Writer) {
	w.Write([]byte(`</html>`))
}

func writeAttributes(w io.Writer, attrs map[string]string) {
	if len(attrs) == 0 {
		return
	}

	for key, value := range attrs {
		w.Write([]byte(` `))
		w.Write([]byte(key))
		w.Write([]byte(`="`))
		htmlEscaper.WriteString(w, value)
		w.Write([]byte(`"`))
	}
}
