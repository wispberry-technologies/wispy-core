package tpl

// Export main types for external use
type Engine = TemplateEngine
type Data = TemplateData

// NewEngine creates a new template engine
func NewEngine(layoutDir, pagesDir string) *Engine {
	return NewTemplateEngine(layoutDir, pagesDir)
}
