package app

import (
	"path/filepath"
	"wispy-core/common"
	"wispy-core/tpl"
)

type wispyCms struct {
	tplEngine tpl.TemplateEngine
	theme     string
}

type WispyCms interface {
	GetTemplateEngine() tpl.TemplateEngine
	GetTheme() string
}

func NewWispyCms() WispyCms {
	supportingTemplatesDirs := []string{
		filepath.Join("_data", "design", "templates", "cms", "partials"),
		filepath.Join("_data", "design", "systems", "components"),
		filepath.Join("_data", "design", "systems", "atoms"),
	}

	// Create template engine for this site
	layoutsDir := filepath.Join("_data", "design", "templates", "cms", "layouts")
	pagesDir := filepath.Join("_data", "design", "templates", "cms", "pages")
	// Initialize the template engine with the pages and layouts directories
	templateEngine := tpl.NewTemplateEngine(layoutsDir, pagesDir)
	_, suppTempErrs := templateEngine.LoadSupportingTemplates(supportingTemplatesDirs)
	if len(suppTempErrs) > 0 {
		common.Error("Failed to load supporting templates!")
		for _, err := range suppTempErrs {
			common.Warning("-->: %v", err)
		}
	}

	return &wispyCms{
		tplEngine: templateEngine,
		theme:     "robot-green",
	}
}

func (wc *wispyCms) GetTemplateEngine() tpl.TemplateEngine {
	return wc.tplEngine
}

func (wc *wispyCms) GetTheme() string {
	return wc.theme
}
