package tpl

import (
	"html/template"
	"os"
	"path/filepath"
)

// TemplateData represents the data passed to templates
type TemplateData struct {
	Title       string
	Description string
	Site        SiteData
	Content     template.HTML
	Data        map[string]interface{}
}

// SiteData represents site information for templates
type SiteData struct {
	Name    string
	Domain  string
	BaseURL string
}

func LoadSupportingTemplates(supportingTemplatesDirs []string) (supportingTemplates *template.Template, errs []error) {
	for _, dir := range supportingTemplatesDirs {
		if supportingTemplates == nil {
			supportingTemplates = template.New(filepath.Base(dir))
		}
		err := WalkAndLoadAllDirectories(dir, supportingTemplates)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return supportingTemplates, errs
}

func WalkAndLoadAllDirectories(dir string, tmpl *template.Template) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".html" {
			return nil
		}
		name := filepath.Base(path)
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		tmpl.New(name).Parse(string(content))
		return nil
	})
}
