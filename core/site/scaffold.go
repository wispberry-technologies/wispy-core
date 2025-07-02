package site

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"
	"wispy-core/common"
	"wispy-core/core/tenant/databases"
	"wispy-core/wispytail"
)

// ScaffoldConfig defines new site parameters for site creation
type ScaffoldConfig struct {
	// ID           string   // Unique identifier for the site (required)
	Name         string   // Display name of the site (required)
	Domain       string   // Domain name for the site, e.g., "example.com" (required)
	BaseURL      string   // Base URL for the site, e.g., "https://example.com" (required)
	ContentTypes []string // List of content types to support, e.g., ["page", "post"] (required)
	WithExample  bool     // Whether to generate example content (optional)
}

// Scaffold creates a new site with complete theme support
// It handles directory creation, theme generation, and config file creation
func Scaffold(rootDir string, cfg ScaffoldConfig) (Site, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	sitePath := filepath.Join(rootDir, cfg.Domain)
	dirs := []string{
		"content",
		"data",
		"databases/sqlite",
		"assets/css",
		"assets/js",
		"design/partials",
		"layouts",
		"design/atoms",
		"design/components",
		"pages/index",
		"pages/about",
		"themes",
	}

	// Create directory structure
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(sitePath, dir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate config file
	configPath := filepath.Join(sitePath, "config.toml")
	if err := generateConfigFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to generate config file: %w", err)
	}

	// Create theme css file
	themeCSSPath := filepath.Join(sitePath, "themes", "default.css")
	themeCSSContent := wispytail.CssReset + "\n" + wispytail.DefaultCssTheme
	if err := os.WriteFile(themeCSSPath, []byte(themeCSSContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to create theme CSS file: %w", err)
	}

	// Create example content if requested
	if cfg.WithExample {
		if err := generateExamples(sitePath, cfg); err != nil {
			return nil, fmt.Errorf("failed to generate example content: %w", err)
		}
	}

	// Initialize the site instance
	siteInstance := &site{
		mu: sync.RWMutex{},
		// ID:         cfg.ID,
		Name:      cfg.Name,
		Domain:    cfg.Domain,
		BaseURL:   cfg.BaseURL,
		Data:      make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return siteInstance, nil
}

// validateConfig checks for required fields and valid values
func validateConfig(cfg ScaffoldConfig) error {
	if !isValidDomain(cfg.Domain) {
		return errors.New("invalid domain format, expected format: example.com")
	}
	if len(cfg.ContentTypes) == 0 {
		return errors.New("at least one content type is required")
	}
	return nil
}

// isValidDomain performs domain validation according to RFC standards
// It checks length, format, and valid characters
func isValidDomain(domain string) bool {
	// Basic length check per RFC
	if len(domain) < 4 || len(domain) > 253 {
		return false
	}

	// No spaces allowed
	if strings.Contains(domain, " ") {
		return false
	}

	// Must have at least one dot
	if !strings.Contains(domain, ".") {
		return false
	}

	// Check if localhost (allowed for development)
	if domain == "localhost" || strings.HasSuffix(domain, ".localhost") {
		return true
	}

	// Check for invalid characters (simplified)
	for _, c := range domain {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '.' && c != '-' {
			return false
		}
	}

	// Domain should not start or end with a hyphen
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return false
	}

	// Domain labels should not start or end with a dot
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	return true
}

// configTemplateData represents data for the site configuration template
type configTemplateData struct {
	ScaffoldConfig
	CreatedAt time.Time
}

// generateConfigFile creates the site configuration TOML file with proper formatting
// It includes site information, design system configuration, and content types
func generateConfigFile(path string, cfg ScaffoldConfig) error {
	const configTemplate = `# Site Configuration
[site]
name = "{{.Name}}"
domain = "{{.Domain}}"
base_url = "{{.BaseURL}}"
created_at = {{.CreatedAt | formatTime}}
updated_at = {{.CreatedAt | formatTime}}

# Content Types
{{- range $index, $type := .ContentTypes}}
[[content_types]]
name = "{{$type}}"
{{- end}}
`

	// Create template function map for formatting
	funcMap := template.FuncMap{
		"formatTime": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
	}

	// Parse the template with the function map
	tmpl, err := template.New("config").Funcs(funcMap).Parse(configTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create the file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Prepare the template data
	data := configTemplateData{
		ScaffoldConfig: cfg,
		CreatedAt:      time.Now(),
	}

	// Execute the template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// generateExamples creates starter content, components, and layouts
// for the new site based on the specified configuration
func generateExamples(sitePath string, cfg ScaffoldConfig) error {
	// Create directory structure for examples
	exampleDirs := []string{
		"pages/about",
		"pages/index",
		"blog",
		"design/components/ui",
		"design/components/blocks",
		"layouts",
		"assets/images",
	}

	for _, dir := range exampleDirs {
		if err := os.MkdirAll(filepath.Join(sitePath, dir), 0755); err != nil {
			return fmt.Errorf("failed to create example directory %s: %w", dir, err)
		}
	}

	// 1. Home page
	homeContent := `
<h1>Welcome to {{.Name}}</h1>
<p>This is your new site powered by our platform.</p>

<h2>Getting Started</h2>
<ol>
	<li>Edit this page in the content/index/index.html file</li>
	<li>Create new pages in the content/ directory</li>
	<li>Customize the design in the design/ directory</li>
</ol>
`

	homeContent = strings.ReplaceAll(homeContent, "{{.Name}}", cfg.Name)
	homeContent = strings.ReplaceAll(homeContent, "{{.Date}}", time.Now().Format("2006-01-02"))
	homePath := filepath.Join(sitePath, "pages", "index", "index.html")
	if err := os.WriteFile(homePath, []byte(homeContent), 0644); err != nil {
		return fmt.Errorf("failed to create home page: %w", err)
	}

	// 2. About page
	aboutContent := `
<p>This is your new site powered by our platform.</p>
`

	aboutContent = strings.ReplaceAll(aboutContent, "{{.Name}}", cfg.Name)
	aboutContent = strings.ReplaceAll(aboutContent, "{{.Date}}", time.Now().Format("2006-01-02"))
	aboutPath := filepath.Join(sitePath, "pages", "about", "index.html")
	if err := os.WriteFile(aboutPath, []byte(aboutContent), 0644); err != nil {
		return fmt.Errorf("failed to create about page: %w", err)
	}

	// 3. Example UI components
	// Button component
	buttonContent := `<!-- A reusable button component with variants -->
<button class="btn btn-{{ .variant | default "primary" }} 
	{{ with .size }}btn-{{ . }}{{ end }}
	{{ with .class }}{{ . }}{{ end }}"
	{{ with .disabled }}disabled{{ end }}
	{{ with .id }}id="{{ . }}"{{ end }}
	{{ with .attributes }}{{ . }}{{ end }}>
	{{ with .icon }}<span class="icon">{{ . }}</span>{{ end }}
	{{ .children }}
</button>`

	buttonPath := filepath.Join(sitePath, "design", "components", "ui", "button.html")
	if err := os.WriteFile(buttonPath, []byte(buttonContent), 0644); err != nil {
		return fmt.Errorf("failed to create button component: %w", err)
	}

	// Card component
	cardContent := `<!-- A reusable card component -->
<div class="card {{ with .class }}{{ . }}{{ end }}"
	{{ with .id }}id="{{ . }}"{{ end }}
	{{ with .attributes }}{{ . }}{{ end }}>
	{{ with .image }}
	<figure>
		<img src="{{ . }}" alt="{{ $.imageAlt | default "Card image" }}" />
	</figure>
	{{ end }}
	<div class="card-body">
		{{ with .title }}<h2 class="card-title">{{ . }}</h2>{{ end }}
		<div class="card-content">
			{{ .children }}
		</div>
		{{ with .actions }}
		<div class="card-actions">
			{{ . }}
		</div>
		{{ end }}
	</div>
</div>`

	cardPath := filepath.Join(sitePath, "design", "components", "ui", "card.html")
	if err := os.WriteFile(cardPath, []byte(cardContent), 0644); err != nil {
		return fmt.Errorf("failed to create card component: %w", err)
	}

	// Hero block component
	heroContent := `<!-- A hero section block -->
<section class="hero min-h-screen {{ with .class }}{{ . }}{{ end }}"
	{{ with .id }}id="{{ . }}"{{ end }}
	{{ with .attributes }}{{ . }}{{ end }}>
	<div class="hero-content text-center">
		<div class="max-w-md">
			{{ with .title }}<h1 class="text-5xl font-bold">{{ . }}</h1>{{ end }}
			{{ with .subtitle }}<p class="py-6">{{ . }}</p>{{ end }}
			<div class="hero-actions">
				{{ .children }}
			</div>
		</div>
	</div>
</section>`

	heroPath := filepath.Join(sitePath, "design", "components", "blocks", "hero.html")
	if err := os.WriteFile(heroPath, []byte(heroContent), 0644); err != nil {
		return fmt.Errorf("failed to create hero component: %w", err)
	}

	// 4. Example layouts
	// Default layout
	layoutContent := `
<body class="bg-base-100 text-base-content min-h-screen flex flex-col">
	<!-- Header -->
	<header class="">
	</header>

	<!-- Main content -->
	<main class="container mx-auto p-4 flex-grow">
		{{ template "content" . }}
	</main>

	<!-- Footer -->
	<footer class="bg-base-200 p-4">
		<div class="container mx-auto text-center">
			<p>&copy; {{.CurrentYear}} {{.Site.Name}}. All rights reserved.</p>
		</div>
	</footer>
</body>
</html>`
	layoutContent = strings.ReplaceAll(layoutContent, "{{.CurrentYear}}", time.Now().Format("2006"))
	layoutPath := filepath.Join(sitePath, "layouts", "default.html")
	if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
		return fmt.Errorf("failed to create default layout: %w", err)
	}

	return nil
}

func ScaffoldSqliteDatabases(sitePath string) error {
	dbm := NewDatabaseManager(sitePath)
	databases := databases.DatabaseScaffolds

	for name := range databases {
		dbPath := filepath.Join(sitePath, "databases", "sqlite", name+".db")
		if err := dbm.CreateDatabase(dbPath); err != nil {
			return fmt.Errorf("failed to create database %s: %w", name, err)
		}

		common.Debug("Created database %s at %s", name, dbPath)
	}

	return nil
}
