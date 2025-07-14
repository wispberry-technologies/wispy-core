package app

import (
	"path/filepath"
	"wispy-core/auth"
	"wispy-core/common"
	"wispy-core/config"
	"wispy-core/core/site"
	"wispy-core/tpl"
)

type wispyCms struct {
	authMiddleware auth.Middleware
	authProvider   auth.AuthProvider
	tplEngine      tpl.TemplateEngine
	theme          string
	siteManager    site.SiteManager
}

type WispyCms interface {
	GetTemplateEngine() tpl.TemplateEngine
	GetTheme() string
	GetSiteManager() site.SiteManager
}

func NewWispyCms(siteManager site.SiteManager) WispyCms {
	gConfig := config.GetGlobalConfig()
	if gConfig == nil {
		common.Error("Global configuration is not initialized!")
		panic("Global configuration is required to initialize Wispy CMS")
	}

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
		tplEngine:   templateEngine,
		theme:       "robot-green",
		siteManager: siteManager,
	}
}

func (wc *wispyCms) GetTemplateEngine() tpl.TemplateEngine {
	return wc.tplEngine
}

func (wc *wispyCms) GetTheme() string {
	return wc.theme
}

func (wc *wispyCms) GetSiteManager() site.SiteManager {
	return wc.siteManager
}
