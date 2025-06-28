Here's a refined implementation that uses a state-pushing approach with clean separation between content rendering and HTML scaffolding:

### 1. **Render State Management (`core/render/state.go`)**
```go
package render

type RenderState struct {
    Head struct {
        Title       string
        MetaTags    []MetaTag
        Styles      []string
        Scripts     []Script
        InlineCSS   string
    }
    Body struct {
        Classes    []string
        Attributes map[string]string
        Content    strings.Builder
    }
    Context *site.SiteContext
}

func NewRenderState(ctx *site.SiteContext) *RenderState {
    return &RenderState{
        Context: ctx,
    }
}

// Push methods for collecting render data
func (rs *RenderState) PushTitle(title string) {
    rs.Head.Title = title
}

func (rs *RenderState) PushScript(src string, opts ScriptOptions) {
    rs.Head.Scripts = append(rs.Head.Scripts, Script{
        Src:   src,
        Async: opts.Async,
        Defer: opts.Defer,
    })
}

func (rs *RenderState) PushStyle(content string) {
    rs.Head.Styles = append(rs.Head.Styles, content)
}

func (rs *RenderState) PushBodyClass(class string) {
    rs.Body.Classes = append(rs.Body.Classes, class)
}
```

### 2. **Page Renderer (`core/render/page.go`)**
```go
package render

func PageRender(ctx *site.SiteContext, page *page.Page) (string, error) {
    state := NewRenderState(ctx)
    
    // Push initial state
    state.PushTitle(page.Title)
    state.PushBodyClass("page-" + page.Slug)
    
    // Render content components (pushes more state)
    content, err := renderContent(state, page)
    if err != nil {
        return "", err
    }
    state.Body.Content.WriteString(content)
    
    // Final HTML assembly
    return HtmlBaseRender(state)
}

func renderContent(state *RenderState, page *page.Page) (string, error) {
    // Create isolated template context
    tmplCtx := map[string]interface{}{
        "State":  state,
        "Page":   page,
        "Site":   state.Context,
        "push":   state, // Allow templates to call push methods
    }
    
    // Execute the content template
    var buf strings.Builder
    err := state.Context.Templates.ExecuteTemplate(&buf, page.Template, tmplCtx)
    return buf.String(), err
}
```

### 3. **HTML Base Renderer (`core/render/html.go`)**
```go
package render

func HtmlBaseRender(state *RenderState) (string, error) {
    // Build the base template context
    data := map[string]interface{}{
        "Head":   state.Head,
        "Body":   state.Body,
        "Content": template.HTML(state.Body.Content.String()),
    }
    
    // Execute base template
    var buf strings.Builder
    err := state.Context.Templates.ExecuteTemplate(&buf, "base.html", data)
    return buf.String(), err
}
```

### 4. **Base Template (`design/systems/base/layouts/base.html`)**
```html
<!DOCTYPE html>
<html lang="{{ .Site.Language }}">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>{{ .Head.Title }}</title>
  
  {{ range .Head.MetaTags }}
  <meta name="{{ .Name }}" content="{{ .Content }}">
  {{ end }}
  
  {{ range .Head.Styles }}
  <style>{{ . }}</style>
  {{ end }}
  
  {{ if .Head.InlineCSS }}
  <style>{{ .Head.InlineCSS }}</style>
  {{ end }}
</head>
<body class="{{ join .Body.Classes " " }}">
  {{ .Content }}
  
  {{ range .Head.Scripts }}
  <script src="{{ .Src }}" {{ if .Async }}async{{ end }} {{ if .Defer }}defer{{ end }}></script>
  {{ end }}
</body>
</html>
```

### 5. **Component Usage Example (`design/systems/base/components/hero.html`)**
```html
{{ define "hero" }}
{{/* Push component requirements to render state */}}
{{ .State.PushStyle (include "css/hero.css") }}
{{ .State.PushScript "js/hero.js" (dict "async" true) }}
{{ .State.PushBodyClass "has-hero" }}

<section class="hero">
  <h1>{{ .Title }}</h1>
  {{ .Content }}
</section>
{{ end }}
```

### 6. **Template Functions (`core/template/funcs.go`)**
```go
// Add these to your FuncMap
funcMap["include"] = func(path string) (string, error) {
    // Implementation that reads and returns file content
    // with cache support
}

funcMap["join"] = func(elems []string, sep string) string {
    return strings.Join(elems, sep)
}
```

### Key Advantages:

1. **Explicit State Management**:
   - Components actively push their requirements
   - No magic variable collection
   - Clear data flow direction

2. **Flexible Composition**:
   - Easy to add new state types (e.g., structured data)
   - Components can modify any part of the render state
   - Base template remains simple and predictable

3. **Performance Optimizations**:
   - State collection during content render
   - Single pass final assembly
   - Built-in deduplication opportunities

4. **Debugging Friendly**:
   - Complete render state inspectable at any point
   - Clear separation between phases
   - No hidden context mutations

### Usage Example:

```go
// In your HTTP handler
func productHandler(w http.ResponseWriter, r *http.Request) {
    ctx := site.FromRequest(r)
    page := loadProductPage(r)
    
    html, err := render.PageRender(ctx, page)
    if err != nil {
        // handle error
    }
    
    w.Header().Set("Content-Type", "text/html")
    w.Write([]byte(html))
}
```

### Advanced Extensions:

1. **State Middleware**:
```go
type StateMiddleware func(*RenderState) error

func WithAnalytics(state *RenderState) error {
    state.PushScript("/js/analytics.js", ScriptOptions{Async: true})
    return nil
}

// Usage:
state.ApplyMiddleware(WithAnalytics, WithTracking)
```

2. **Render Hooks**:
```go
type RenderHook func(*RenderState, template.HTML) (template.HTML, error)

func WithSyntaxHighlighting(state *RenderState, content template.HTML) (template.HTML, error) {
    // Process content and return highlighted version
}
```

3. **Asset Bundling**:
```go
func (rs *RenderState) FinalizeAssets() {
    rs.Head.Styles = bundleCSS(rs.Head.Styles)
    rs.Head.Scripts = bundleJS(rs.Head.Scripts)
}
```

Would you like me to elaborate on any specific aspect of this architecture?