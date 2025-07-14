package tpl

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"wispy-core/common"
	"wispy-core/wispytail"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TemplateEngine handles template loading and rendering
type templateEngine struct {
	mu                  sync.RWMutex
	layoutsDir          string
	layouts             map[string][]byte
	templatesDir        string
	templates           map[string][]byte
	supportingTemplates *template.Template
	wispyTailTrie       *common.Trie
	funcMap             template.FuncMap
}

type TemplateEngine interface {
	LoadSupportingTemplates(supportingTemplatesDirs []string) (*template.Template, []error)
	// TODO: Add support for walking directories and loading templates
	// ScanAndLoadAllTemplates(rootDir string) error
	LoadTemplate(templatePathName string) ([]byte, error)
	LoadLayout(layoutPathName string) ([]byte, error)
	//
	GetTemplatesMap() map[string][]byte
	GetSupportingTemplates() *template.Template
	GetWispyTailTrie() *common.Trie
	GetFuncMap() template.FuncMap
	UpdateFuncMap(rs RenderState) // Update function map with render state for stateful functions
	// With layout support
	RenderWithLayout(templatePathName, layoutPathName string, data TemplateData) (RenderState, error)
	// Basic rendering
	RenderTemplate(templatePathName string, data TemplateData) (RenderState, error)
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(layoutsDir, templatesDir string) TemplateEngine {
	return &templateEngine{
		layoutsDir:          layoutsDir,
		layouts:             make(map[string][]byte),
		templatesDir:        templatesDir,
		templates:           make(map[string][]byte),
		supportingTemplates: template.New("supporting"),
		wispyTailTrie:       wispytail.GetBaseTrie(),
		funcMap:             getDefaultFuncMap(nil), // Initialize with default functions
	}
}

// getDefaultFuncMap returns a map of default template functions
func getDefaultFuncMap(rs RenderState) template.FuncMap {
	return template.FuncMap{
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"title": func(s string) string {
			return cases.Title(language.English).String(s)
		},
		"join": func(sep string, items []string) string {
			return strings.Join(items, sep)
		},
		"contains": func(substr, s string) bool {
			return strings.Contains(s, substr)
		},
		"has": func(items []string, item string) bool {
			return slices.Contains(items, item)
		},
		"dict": func(keysAndValues ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(keysAndValues); i += 2 {
				if i+1 < len(keysAndValues) {
					key := fmt.Sprintf("%v", keysAndValues[i])
					dict[key] = keysAndValues[i+1]
				}
			}
			return dict
		},
		"slice": func(items ...interface{}) []interface{} {
			return items
		},
		"substr": func(start, length int, s string) string {
			if start < 0 || start >= len(s) {
				return ""
			}
			end := start + length
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"default": func(defaultValue interface{}, value interface{}) interface{} {
			if value == nil {
				return defaultValue
			}
			// Check for empty string
			if str, ok := value.(string); ok && str == "" {
				return defaultValue
			}
			// Check for zero values
			if fmt.Sprintf("%v", value) == "" {
				return defaultValue
			}
			return value
		},
		"eq": func(a, b interface{}) bool {
			return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}
}

func (te *templateEngine) GetTemplatesMap() map[string][]byte {
	te.mu.RLock()
	defer te.mu.RUnlock()
	return te.templates
}

func (te *templateEngine) GetSupportingTemplates() *template.Template {
	te.mu.RLock()
	defer te.mu.RUnlock()
	return te.supportingTemplates
}

func (te *templateEngine) GetWispyTailTrie() *common.Trie {
	return te.wispyTailTrie
}

func (te *templateEngine) GetFuncMap() template.FuncMap {
	te.mu.RLock()
	defer te.mu.RUnlock()
	return te.funcMap
}

// UpdateFuncMap updates the function map with render state for stateful functions
func (te *templateEngine) UpdateFuncMap(rs RenderState) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.funcMap = getDefaultFuncMap(rs)
}

// LoadTemplate loads and caches a template from the given path.
func (te *templateEngine) LoadTemplate(templatePath string) ([]byte, error) {
	fullPath := filepath.Join(te.templatesDir, templatePath)
	if te.templatesDir == "" {
		fullPath = templatePath
	}

	te.mu.Lock()
	defer te.mu.Unlock()

	var templateData []byte
	var err error

	if cachedData, ok := te.templates[templatePath]; ok {
		return cachedData, nil
	} else {
		templateData, err = os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", fullPath, err)
		}
		te.templates[templatePath] = templateData
	}

	return templateData, nil
}

// LoadLayout loads and caches a layout from the given path.
func (te *templateEngine) LoadLayout(layoutPath string) ([]byte, error) {
	fullPath := filepath.Join(te.layoutsDir, layoutPath)
	if te.templatesDir == "" {
		fullPath = layoutPath
	}

	te.mu.Lock()
	defer te.mu.Unlock()

	var data []byte
	var err error

	if cachedData, ok := te.layouts[layoutPath]; ok {
		return cachedData, nil
	} else {
		data, err = os.ReadFile(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read layout %s: %w", fullPath, err)
		}
		te.layouts[layoutPath] = data
	}

	return data, nil
}

// LoadSupportingTemplates parses all templates in the provided directories into the supporting template set.
func (te *templateEngine) LoadSupportingTemplates(dirs []string) (*template.Template, []error) {
	te.mu.Lock()
	defer te.mu.Unlock()
	var errs []error

	// Reset supporting templates with function map
	te.supportingTemplates = template.New("supporting").Funcs(te.funcMap)

	common.Debug("Loading supporting templates")
	for _, dir := range dirs {
		// Walk each directory for template files
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				errs = append(errs, fmt.Errorf("error accessing %s: %w", path, err))
				return nil
			}
			if info.IsDir() {
				return nil
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				errs = append(errs, fmt.Errorf("failed to read supporting template %s: %w", path, readErr))
				return nil
			}
			name := common.NormalizeTemplateName(dir, path)
			name = filepath.Base(dir) + "/" + name // Ensure unique names based on directory
			common.Debug("-- [%s](%s)", name, path)
			if _, parseErr := te.supportingTemplates.New(name).Parse(string(data)); parseErr != nil {
				errs = append(errs, fmt.Errorf("failed to parse supporting template %s: %w", path, parseErr))
			}
			return nil
		})
		if err != nil {
			errs = append(errs, err)
		}
	}
	return te.supportingTemplates, errs
}

// RenderWithLayoutTo renders a template with the given layout to a writer.
func (te *templateEngine) RenderWithLayout(templatePath, layoutPath string, data TemplateData) (RenderState, error) {
	layoutData, err := te.LoadLayout(layoutPath)
	if err != nil {
		common.Error("Failed to load layout %s: %v", layoutPath, err)
		return nil, err
	}

	contentData, err := te.LoadTemplate(templatePath)
	if err != nil {
		common.Error("Failed to load template %s: %v", templatePath, err)
		return nil, fmt.Errorf("failed to load template %s: %w", templatePath, err)
	}
	// Create a render state to store rendering information
	rs := NewRenderState()

	// Clone supporting templates
	tmpl, err := te.supportingTemplates.Clone()
	tmpl.Funcs(getDefaultFuncMap(rs)) // Pass render state to function map
	if err != nil {
		return nil, fmt.Errorf("failed to clone supporting templates: %w", err)
	}

	// Parse the layout template
	_, err = tmpl.Parse(string(layoutData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse layout template %s: %w", layoutPath, err)
	}

	// Parse the content template and associate it with the layout template
	// This allows the content template to define blocks that the layout references
	_, err = tmpl.Parse(string(contentData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse content template %s: %w", templatePath, err)
	}

	// Populate render state with data from template
	PopulateRenderStateFromTemplateData(rs, data)

	// Execute the combined template (layout + content blocks)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data.Data); err != nil {
		return nil, fmt.Errorf("failed to render template with layout: %w", err)
	}

	// Set the rendered content in the render state
	rs.SetBody(buf.String())

	return rs, nil
}

// RenderWithLayoutTo renders a template with the given layout to a writer.
func (te *templateEngine) RenderTemplate(templatePath string, data TemplateData) (RenderState, error) {
	result, err := te.LoadTemplate(templatePath)
	if err != nil {
		common.Error("Failed to load template %s: %v", templatePath, err)
		return nil, fmt.Errorf("failed to load template %s: %w", templatePath, err)
	}

	// Create a render state to store rendering information
	rs := NewRenderState()

	// Clone supporting templates
	tmpl, err := te.supportingTemplates.Clone()
	tmpl.Funcs(getDefaultFuncMap(rs)) // Pass render state to function map
	if err != nil {
		return nil, fmt.Errorf("failed to clone supporting templates: %w", err)
	}

	// Parse the template
	_, err = tmpl.Parse(string(result))

	// Populate render state with data from template
	PopulateRenderStateFromTemplateData(rs, data)

	var contentBuf bytes.Buffer
	if err := tmpl.Execute(&contentBuf, data.Data); err != nil {
		return nil, fmt.Errorf("failed to render body template %s: %w", templatePath, err)
	}

	// Set the body content in the render state
	rs.SetBody(contentBuf.String())

	// Convert our RenderState to core/render.RenderState and render the complete HTML page
	return rs, err
}
