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
)

// ScaffoldConfig defines new site parameters for site creation
type ScaffoldConfig struct {
	ID               string            // Unique identifier for the site (required)
	Name             string            // Display name of the site (required)
	Domain           string            // Domain name for the site, e.g., "example.com" (required)
	BaseURL          string            // Base URL for the site, e.g., "https://example.com" (required)
	ThemeName        string            // Name of the theme, e.g., "pale-wisp" (required)
	ThemeMode        string            // Theme mode: "light" or "dark" (required)
	ContentTypes     []string          // List of content types to support, e.g., ["page", "post"] (required)
	WithExample      bool              // Whether to generate example content (optional)
	ThemeOptions     map[string]string // Custom theme options to override defaults (optional)
	TypographyPreset string            // Typography preset name: "default", "modern", "classic", "minimal" (optional)
	ColorPreset      string            // Color preset name: "default", "earthy", "vibrant", "monochrome" (optional)
}

// Scaffold creates a new site with complete theme support
// It handles directory creation, theme generation, and config file creation
func Scaffold(rootDir string, cfg ScaffoldConfig) (Site, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	sitePath := filepath.Join(rootDir, cfg.ID)
	dirs := []string{
		"content",
		"data",
		"themes",
		"assets/css",
		"assets/js",
		"design/partials",
		"design/layouts",
	}

	// Create directory structure
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(sitePath, dir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Generate theme
	theme := createTheme(cfg.ThemeName, cfg.ThemeMode, cfg)
	// if err := generateThemeFiles(sitePath, theme); err != nil {
	// 	return nil, fmt.Errorf("failed to generate theme files: %w", err)
	// }

	// Generate config file
	configPath := filepath.Join(sitePath, "config.toml")
	if err := generateConfigFile(configPath, cfg, theme); err != nil {
		return nil, fmt.Errorf("failed to generate config file: %w", err)
	}

	// Create example content if requested
	if cfg.WithExample {
		if err := generateExamples(sitePath, cfg); err != nil {
			return nil, fmt.Errorf("failed to generate example content: %w", err)
		}
	}

	// Initialize the site instance
	siteInstance := &site{
		mu:         sync.RWMutex{},
		ID:         cfg.ID,
		Name:       cfg.Name,
		Domain:     cfg.Domain,
		BaseURL:    cfg.BaseURL,
		Theme:      theme,
		ContentDir: "content",
		Data:       make(map[string]interface{}),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return siteInstance, nil
}

// createTheme generates a complete theme configuration based on the provided configuration options
func createTheme(name, mode string, cfg ScaffoldConfig) *Theme {
	// Validate and normalize theme mode
	if mode != "dark" {
		mode = "light" // default to light
	}

	// Get typography preset, defaulting if not specified
	typographyPreset := cfg.TypographyPreset
	if typographyPreset == "" {
		typographyPreset = "default"
	}

	// Get color preset, defaulting if not specified
	colorPreset := cfg.ColorPreset
	if colorPreset == "" {
		colorPreset = "default"
	}

	// Create base theme structure
	theme := &Theme{
		Name: name,
		Base: mode,
		Tokens: ThemeTokens{
			Spacing:    createSpacingTokens(),
			Typography: createTypographyTokens(typographyPreset),
			Borders:    createBorderTokens(mode),
			Shadows:    createShadowTokens(),
			Animations: createAnimationTokens(),
		},
		Variables: createThemeVariables(mode),
	}

	// Set color tokens based on the theme mode and preset
	if mode == "dark" {
		theme.Tokens.Colors = createDarkColorTokens(colorPreset)
	} else {
		theme.Tokens.Colors = createLightColorTokens(colorPreset)
	}

	// Apply any custom theme options provided in the config
	if cfg.ThemeOptions != nil {
		applyThemeOptions(theme, cfg.ThemeOptions)
	}

	return theme
}

// createLightColorTokens generates the color tokens for light mode
// Different presets can be specified to generate different palettes
func createLightColorTokens(preset string) ColorTokens {
	// Default light mode colors
	tokens := ColorTokens{
		Primary:          "oklch(69% 0.17 162.48)",
		PrimaryContent:   "oklch(97% 0.021 166.113)",
		Secondary:        "oklch(39% 0.095 152.535)",
		SecondaryContent: "oklch(94% 0.028 342.258)",
		Accent:           "oklch(62% 0.214 259.815)",
		AccentContent:    "oklch(97% 0.014 254.604)",
		Neutral:          "oklch(27% 0.006 286.033)",
		NeutralContent:   "oklch(98% 0.001 106.423)",
		Base100:          "oklch(100% 0 0)",
		Base200:          "oklch(95% 0.01 150)",
		Base300:          "oklch(90% 0.02 150)",
		BaseContent:      "oklch(15% 0.03 150)",
		Info:             "oklch(70% 0.165 254.624)",
		InfoContent:      "oklch(28% 0.091 267.935)",
		Success:          "oklch(84% 0.238 128.85)",
		SuccessContent:   "oklch(27% 0.072 132.109)",
		Warning:          "oklch(75% 0.183 55.934)",
		WarningContent:   "oklch(26% 0.079 36.259)",
		Error:            "oklch(71% 0.194 13.428)",
		ErrorContent:     "oklch(27% 0.105 12.094)",
	}

	// Apply preset-specific overrides
	switch preset {
	case "earthy":
		tokens.Primary = "oklch(65% 0.2 120)"   // Green-yellow
		tokens.Secondary = "oklch(55% 0.18 85)" // Earthy orange
		tokens.Accent = "oklch(60% 0.15 40)"    // Warm brown
	case "vibrant":
		tokens.Primary = "oklch(70% 0.25 250)"  // Vibrant blue
		tokens.Secondary = "oklch(65% 0.3 320)" // Vibrant purple
		tokens.Accent = "oklch(75% 0.25 140)"   // Vibrant green
	case "monochrome":
		tokens.Primary = "oklch(50% 0.01 260)"
		tokens.Secondary = "oklch(60% 0.01 260)"
		tokens.Accent = "oklch(40% 0.01 260)"
		tokens.Base200 = "oklch(96% 0 0)"
		tokens.Base300 = "oklch(92% 0 0)"
	}

	return tokens
}

// createDarkColorTokens generates the color tokens for dark mode
// Different presets can be specified to generate different palettes
func createDarkColorTokens(preset string) ColorTokens {
	// Default dark mode colors
	tokens := ColorTokens{
		Primary:          "oklch(84% 0.238 128.85)",
		PrimaryContent:   "oklch(26% 0.065 152.934)",
		Secondary:        "oklch(39% 0.095 152.535)",
		SecondaryContent: "oklch(94% 0.028 342.258)",
		Accent:           "oklch(26% 0 0)",
		AccentContent:    "oklch(98% 0.031 120.757)",
		Neutral:          "oklch(14% 0.005 285.823)",
		NeutralContent:   "oklch(92% 0.004 286.32)",
		Base100:          "oklch(25.33% 0.016 252.42)",
		Base200:          "oklch(23.26% 0.014 253.1)",
		Base300:          "oklch(21.15% 0.012 254.09)",
		BaseContent:      "oklch(97.807% 0.029 256.847)",
		Info:             "oklch(74% 0.16 232.661)",
		InfoContent:      "oklch(29% 0.066 243.157)",
		Success:          "oklch(79% 0.209 151.711)",
		SuccessContent:   "oklch(38% 0.063 188.416)",
		Warning:          "oklch(82% 0.189 84.429)",
		WarningContent:   "oklch(41% 0.112 45.904)",
		Error:            "oklch(71% 0.194 13.428)",
		ErrorContent:     "oklch(27% 0.105 12.094)",
	}

	// Apply preset-specific overrides
	switch preset {
	case "earthy":
		tokens.Primary = "oklch(70% 0.15 140)"  // Muted green
		tokens.Secondary = "oklch(60% 0.12 90)" // Muted orange
		tokens.Base100 = "oklch(18% 0.02 60)"   // Dark brown base
	case "vibrant":
		tokens.Primary = "oklch(75% 0.2 245)"    // Vibrant blue
		tokens.Secondary = "oklch(70% 0.25 320)" // Vibrant purple
		tokens.Base100 = "oklch(15% 0.04 280)"   // Deep purple-black
	case "monochrome":
		tokens.Primary = "oklch(75% 0.01 260)"
		tokens.Secondary = "oklch(65% 0.01 260)"
		tokens.Base100 = "oklch(18% 0 0)"
		tokens.Base200 = "oklch(15% 0 0)"
		tokens.Base300 = "oklch(13% 0 0)"
	}

	return tokens
}

// createSpacingTokens generates the spacing tokens
func createSpacingTokens() SpacingTokens {
	return SpacingTokens{
		Selector: "0.5rem",
		Field:    "0.25rem",
		Base:     "1rem",
		Sm:       "0.5rem",
		Md:       "1.5rem",
		Lg:       "2rem",
		Xl:       "3rem",
	}
}

// createTypographyTokens generates the typography tokens
// This can be customized based on the typography preset
func createTypographyTokens(preset string) TypographyTokens {
	// Default typography settings
	tokens := TypographyTokens{
		FontSans:       "Inter, system-ui, sans-serif",
		FontMono:       "JetBrains Mono, Roboto Mono, monospace",
		FontSerif:      "Georgia, serif",
		FontSize:       "1rem",
		FontSizeSm:     "0.875rem",
		FontSizeMd:     "1.125rem",
		FontSizeLg:     "1.25rem",
		FontSizeXl:     "1.5rem",
		LineHeight:     "1.5",
		LineHeightSm:   "1.25",
		LineHeightMd:   "1.625",
		FontWeight:     "400",
		FontWeightMd:   "500",
		FontWeightBold: "700",
	}

	// Apply preset-specific overrides
	switch preset {
	case "modern":
		tokens.FontSans = "Poppins, system-ui, sans-serif"
		tokens.LineHeight = "1.6"
		tokens.FontSizeMd = "1.15rem"
	case "classic":
		tokens.FontSerif = "Merriweather, Georgia, serif"
		tokens.FontSans = "Source Sans Pro, system-ui, sans-serif"
		tokens.LineHeight = "1.7"
	case "minimal":
		tokens.FontSans = "IBM Plex Sans, system-ui, sans-serif"
		tokens.FontMono = "IBM Plex Mono, monospace"
		tokens.LineHeight = "1.4"
		tokens.FontWeight = "300"
		tokens.FontWeightMd = "400"
		tokens.FontWeightBold = "600"
	}

	return tokens
}

// createBorderTokens generates the border tokens
func createBorderTokens(mode string) BorderTokens {
	tokens := BorderTokens{
		RadiusSelector: "0.5rem",
		RadiusField:    "0.25rem",
		RadiusBox:      "0.25rem",
		RadiusRound:    "9999px",
		Width:          "1px",
		Style:          "solid",
		StyleDashed:    "dashed",
	}

	// Dark mode has thicker borders for better visibility
	if mode == "dark" {
		tokens.Width = "1.5px"
	}

	return tokens
}

// createShadowTokens generates the shadow tokens
func createShadowTokens() ShadowTokens {
	return ShadowTokens{
		Base:  "0 1px 3px rgba(0,0,0,0.1)",
		Sm:    "0 1px 2px rgba(0,0,0,0.05)",
		Md:    "0 4px 6px rgba(0,0,0,0.1)",
		Lg:    "0 10px 15px rgba(0,0,0,0.1)",
		Xl:    "0 20px 25px rgba(0,0,0,0.1)",
		Inner: "inset 0 2px 4px 0 rgba(0,0,0,0.05)",
		None:  "none",
	}
}

// createAnimationTokens generates the animation tokens
func createAnimationTokens() AnimationTokens {
	return AnimationTokens{
		DurationFast:   "100ms",
		DurationNormal: "200ms",
		DurationSlow:   "300ms",
		FunctionEase:   "ease",
		FunctionLinear: "linear",
		FunctionBounce: "cubic-bezier(0.5, -0.5, 0.2, 1.5)",
	}
}

// applyThemeOptions applies custom theme options to override default theme values
func applyThemeOptions(theme *Theme, options map[string]string) {
	// Apply custom color overrides if specified
	if primary, ok := options["primary"]; ok && primary != "" {
		theme.Tokens.Colors.Primary = primary
	}
	if secondary, ok := options["secondary"]; ok && secondary != "" {
		theme.Tokens.Colors.Secondary = secondary
	}
	if accent, ok := options["accent"]; ok && accent != "" {
		theme.Tokens.Colors.Accent = accent
	}

	// Apply custom typography overrides if specified
	if fontSans, ok := options["font_sans"]; ok && fontSans != "" {
		theme.Tokens.Typography.FontSans = fontSans
	}
	if fontMono, ok := options["font_mono"]; ok && fontMono != "" {
		theme.Tokens.Typography.FontMono = fontMono
	}

	// Apply custom border radius overrides if specified
	if radius, ok := options["radius"]; ok && radius != "" {
		theme.Tokens.Borders.RadiusField = radius
		theme.Tokens.Borders.RadiusBox = radius
	}

	// Apply custom variable overrides
	for key, value := range options {
		if _, exists := theme.Variables[key]; exists || strings.HasPrefix(key, "custom_") {
			theme.Variables[key] = value
		}
	}
}

// createThemeVariables generates the theme variables
func createThemeVariables(mode string) map[string]string {
	variables := map[string]string{
		"noise":              "0",
		"depth":              "0",
		"transition-default": "all 200ms ease",
		"transition-slow":    "all 300ms ease",
		"transition-fast":    "all 100ms ease",
		"container-padding":  "1rem",
	}

	// Different focus ring styling based on theme mode
	if mode == "dark" {
		variables["focus-ring"] = "0 0 0 2px var(--color-primary-content)"
	} else {
		variables["focus-ring"] = "0 0 0 2px var(--color-primary)"
	}

	return variables
}

// // generateThemeFiles creates theme CSS and related files
// func generateThemeFiles(sitePath string, theme *Theme) error {
// 	// Generate theme CSS file
// 	themeCSSPath := filepath.Join(sitePath, "assets", "css", "theme.css")
// 	if err := generateThemeCSS(themeCSSPath, theme); err != nil {
// 		return fmt.Errorf("failed to generate theme CSS: %w", err)
// 	}

// 	// Create theme.toml configuration
// 	themeConfigPath := filepath.Join(sitePath, "themes", theme.Name, "theme.toml")
// 	if err := os.MkdirAll(filepath.Dir(themeConfigPath), 0755); err != nil {
// 		return fmt.Errorf("failed to create theme directory: %w", err)
// 	}

// 	themeConfig := fmt.Sprintf(`# Theme Configuration
// name = "%s"
// base = "%s"
// version = "1.0.0"

// [tokens.colors]
// primary = "%s"
// primary_content = "%s"
// secondary = "%s"
// secondary_content = "%s"
// accent = "%s"
// accent_content = "%s"
// neutral = "%s"
// neutral_content = "%s"
// base_100 = "%s"
// base_200 = "%s"
// base_300 = "%s"
// base_content = "%s"
// info = "%s"
// info_content = "%s"
// success = "%s"
// success_content = "%s"
// warning = "%s"
// warning_content = "%s"
// error = "%s"
// error_content = "%s"

// [tokens.spacing]
// selector = "%s"
// field = "%s"
// base = "%s"
// sm = "%s"
// md = "%s"
// lg = "%s"
// xl = "%s"

// [tokens.typography]
// font_sans = "%s"
// font_mono = "%s"
// font_serif = "%s"

// [tokens.borders]
// width = "%s"
// radius_selector = "%s"
// radius_field = "%s"
// radius_box = "%s"
// `,
// 		theme.Name,
// 		theme.Base,
// 		theme.Tokens.Colors.Primary,
// 		theme.Tokens.Colors.PrimaryContent,
// 		theme.Tokens.Colors.Secondary,
// 		theme.Tokens.Colors.SecondaryContent,
// 		theme.Tokens.Colors.Accent,
// 		theme.Tokens.Colors.AccentContent,
// 		theme.Tokens.Colors.Neutral,
// 		theme.Tokens.Colors.NeutralContent,
// 		theme.Tokens.Colors.Base100,
// 		theme.Tokens.Colors.Base200,
// 		theme.Tokens.Colors.Base300,
// 		theme.Tokens.Colors.BaseContent,
// 		theme.Tokens.Colors.Info,
// 		theme.Tokens.Colors.InfoContent,
// 		theme.Tokens.Colors.Success,
// 		theme.Tokens.Colors.SuccessContent,
// 		theme.Tokens.Colors.Warning,
// 		theme.Tokens.Colors.WarningContent,
// 		theme.Tokens.Colors.Error,
// 		theme.Tokens.Colors.ErrorContent,
// 		theme.Tokens.Spacing.Selector,
// 		theme.Tokens.Spacing.Field,
// 		theme.Tokens.Spacing.Base,
// 		theme.Tokens.Spacing.Sm,
// 		theme.Tokens.Spacing.Md,
// 		theme.Tokens.Spacing.Lg,
// 		theme.Tokens.Spacing.Xl,
// 		theme.Tokens.Typography.FontSans,
// 		theme.Tokens.Typography.FontMono,
// 		theme.Tokens.Typography.FontSerif,
// 		theme.Tokens.Borders.Width,
// 		theme.Tokens.Borders.RadiusSelector,
// 		theme.Tokens.Borders.RadiusField,
// 		theme.Tokens.Borders.RadiusBox,
// 	)

// 	if err := os.WriteFile(themeConfigPath, []byte(themeConfig), 0644); err != nil {
// 		return fmt.Errorf("failed to write theme config: %w", err)
// 	}

// 	return nil
// }

// generateThemeCSS creates the CSS file with theme variables
// It provides a comprehensive set of CSS variables for theme customization
func generateThemeCSS(filePath string, theme *Theme) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create theme CSS directory: %w", err)
	}

	css := fmt.Sprintf(`/* 
 * Generated Theme CSS - %s (%s)
 * Generated by Wispy Core on %s
 */

:root {
  /* Colors */
  --color-primary: %s;
  --color-primary-content: %s;
  --color-secondary: %s;
  --color-secondary-content: %s;
  --color-accent: %s;
  --color-accent-content: %s;
  --color-neutral: %s;
  --color-neutral-content: %s;
  --color-base-100: %s;
  --color-base-200: %s;
  --color-base-300: %s;
  --color-base-content: %s;
  --color-info: %s;
  --color-info-content: %s;
  --color-success: %s;
  --color-success-content: %s;
  --color-warning: %s;
  --color-warning-content: %s;
  --color-error: %s;
  --color-error-content: %s;
  
  /* Spacing */
  --spacing-selector: %s;
  --spacing-field: %s;
  --spacing-base: %s;
  --spacing-sm: %s;
  --spacing-md: %s;
  --spacing-lg: %s;
  --spacing-xl: %s;
  
  /* Typography */
  --font-sans: %s;
  --font-mono: %s;
  --font-serif: %s;
  --font-size: %s;
  --font-size-sm: %s;
  --font-size-md: %s;
  --font-size-lg: %s;
  --font-size-xl: %s;
  --line-height: %s;
  --line-height-sm: %s;
  --line-height-md: %s;
  --font-weight: %s;
  --font-weight-md: %s;
  --font-weight-bold: %s;
  
  /* Borders */
  --border-width: %s;
  --radius-selector: %s;
  --radius-field: %s;
  --radius-box: %s;
  --radius-round: %s;
  --border-style: %s;
  --border-style-dashed: %s;
  
  /* Shadows */
  --shadow-base: %s;
  --shadow-sm: %s;
  --shadow-md: %s;
  --shadow-lg: %s;
  --shadow-xl: %s;
  --shadow-inner: %s;
  --shadow-focus: %s;
  
  /* Animations */
  --duration-fast: %s;
  --duration-normal: %s;
  --duration-slow: %s;
  --function-ease: %s;
  --function-linear: %s;
  --function-bounce: %s;

  /* Theme variables */
  --focus-ring: %s;
  --transition-default: %s;
  --transition-slow: %s;
  --transition-fast: %s;
  --container-padding: %s;
  --noise: %s;
  --depth: %s;
}

/* CSS Theme Reset & Base Styles */
*, *::before, *::after {
  box-sizing: border-box;
}

html {
  font-family: var(--font-sans);
  font-size: var(--font-size);
  line-height: var(--line-height);
  -webkit-text-size-adjust: 100%%;
  -moz-tab-size: 4;
  tab-size: 4;
}

body {
  margin: 0;
  font-family: var(--font-sans);
  color: var(--color-base-content);
  background-color: var(--color-base-100);
  line-height: var(--line-height);
  font-weight: var(--font-weight);
}

/* Typography basics */
h1, h2, h3, h4, h5, h6 {
  margin-top: 0;
  font-weight: var(--font-weight-bold);
  line-height: var(--line-height-sm);
}

h1 { font-size: 2rem; }
h2 { font-size: 1.5rem; }
h3 { font-size: 1.25rem; }
h4 { font-size: 1.125rem; }
h5 { font-size: 1rem; }
h6 { font-size: 0.875rem; }

p {
  margin-top: 0;
  margin-bottom: 1rem;
}

code, pre, kbd {
  font-family: var(--font-mono);
}

/* Focus styling */
:focus-visible {
  outline: var(--focus-ring);
  outline-offset: 2px;
}

/* Transition defaults */
button, a, input, select, textarea {
  transition: var(--transition-default);
}

/* Custom theme components & utilities can be added below */
`,
		theme.Name,
		theme.Base,
		time.Now().Format(time.RFC3339),
		theme.Tokens.Colors.Primary,
		theme.Tokens.Colors.PrimaryContent,
		theme.Tokens.Colors.Secondary,
		theme.Tokens.Colors.SecondaryContent,
		theme.Tokens.Colors.Accent,
		theme.Tokens.Colors.AccentContent,
		theme.Tokens.Colors.Neutral,
		theme.Tokens.Colors.NeutralContent,
		theme.Tokens.Colors.Base100,
		theme.Tokens.Colors.Base200,
		theme.Tokens.Colors.Base300,
		theme.Tokens.Colors.BaseContent,
		theme.Tokens.Colors.Info,
		theme.Tokens.Colors.InfoContent,
		theme.Tokens.Colors.Success,
		theme.Tokens.Colors.SuccessContent,
		theme.Tokens.Colors.Warning,
		theme.Tokens.Colors.WarningContent,
		theme.Tokens.Colors.Error,
		theme.Tokens.Colors.ErrorContent,
		theme.Tokens.Spacing.Selector,
		theme.Tokens.Spacing.Field,
		theme.Tokens.Spacing.Base,
		theme.Tokens.Spacing.Sm,
		theme.Tokens.Spacing.Md,
		theme.Tokens.Spacing.Lg,
		theme.Tokens.Spacing.Xl,
		theme.Tokens.Typography.FontSans,
		theme.Tokens.Typography.FontMono,
		theme.Tokens.Typography.FontSerif,
		theme.Tokens.Typography.FontSize,
		theme.Tokens.Typography.FontSizeSm,
		theme.Tokens.Typography.FontSizeMd,
		theme.Tokens.Typography.FontSizeLg,
		theme.Tokens.Typography.FontSizeXl,
		theme.Tokens.Typography.LineHeight,
		theme.Tokens.Typography.LineHeightSm,
		theme.Tokens.Typography.LineHeightMd,
		theme.Tokens.Typography.FontWeight,
		theme.Tokens.Typography.FontWeightMd,
		theme.Tokens.Typography.FontWeightBold,
		theme.Tokens.Borders.Width,
		theme.Tokens.Borders.RadiusSelector,
		theme.Tokens.Borders.RadiusField,
		theme.Tokens.Borders.RadiusBox,
		theme.Tokens.Borders.RadiusRound,
		theme.Tokens.Borders.Style,
		theme.Tokens.Borders.StyleDashed,
		theme.Tokens.Shadows.Base,
		theme.Tokens.Shadows.Sm,
		theme.Tokens.Shadows.Md,
		theme.Tokens.Shadows.Lg,
		theme.Tokens.Shadows.Xl,
		theme.Tokens.Shadows.Inner,
		theme.Tokens.Shadows.Focus,
		theme.Tokens.Animations.DurationFast,
		theme.Tokens.Animations.DurationNormal,
		theme.Tokens.Animations.DurationSlow,
		theme.Tokens.Animations.FunctionEase,
		theme.Tokens.Animations.FunctionLinear,
		theme.Tokens.Animations.FunctionBounce,
		theme.Variables["focus-ring"],
		theme.Variables["transition-default"],
		theme.Variables["transition-slow"],
		theme.Variables["transition-fast"],
		theme.Variables["container-padding"],
		theme.Variables["noise"],
		theme.Variables["depth"],
	)

	return os.WriteFile(filePath, []byte(css), 0644)
}

// validateConfig checks for required fields and valid values
func validateConfig(cfg ScaffoldConfig) error {
	if cfg.ID == "" {
		return errors.New("site ID cannot be empty")
	}
	if !isValidDomain(cfg.Domain) {
		return errors.New("invalid domain format, expected format: example.com")
	}
	if cfg.ThemeName == "" {
		return errors.New("theme name cannot be empty")
	}
	if cfg.ThemeMode != "light" && cfg.ThemeMode != "dark" {
		return errors.New("theme mode must be either 'light' or 'dark'")
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
	Theme     *Theme
	CreatedAt time.Time
}

// generateConfigFile creates the site configuration TOML file with proper formatting
// It includes site information, design system configuration, and content types
func generateConfigFile(path string, cfg ScaffoldConfig, theme *Theme) error {
	const configTemplate = `# Site Configuration
[site]
id = "{{.ID}}"
name = "{{.Name}}"
domain = "{{.Domain}}"
base_url = "{{.BaseURL}}"
content_dir = "content"
created_at = {{.CreatedAt | formatTime}}
updated_at = {{.CreatedAt | formatTime}}

# Design System Configuration
[design]
theme = "{{.ThemeName}}"
mode = "{{.ThemeMode}}"

# Theme Tokens
[design.tokens.colors]
primary = "{{.Theme.Tokens.Colors.Primary}}"
primary_content = "{{.Theme.Tokens.Colors.PrimaryContent}}"
secondary = "{{.Theme.Tokens.Colors.Secondary}}"
secondary_content = "{{.Theme.Tokens.Colors.SecondaryContent}}"
accent = "{{.Theme.Tokens.Colors.Accent}}"
accent_content = "{{.Theme.Tokens.Colors.AccentContent}}"
neutral = "{{.Theme.Tokens.Colors.Neutral}}"
neutral_content = "{{.Theme.Tokens.Colors.NeutralContent}}"
base_100 = "{{.Theme.Tokens.Colors.Base100}}"
base_200 = "{{.Theme.Tokens.Colors.Base200}}"
base_300 = "{{.Theme.Tokens.Colors.Base300}}"
base_content = "{{.Theme.Tokens.Colors.BaseContent}}"
info = "{{.Theme.Tokens.Colors.Info}}"
info_content = "{{.Theme.Tokens.Colors.InfoContent}}"
success = "{{.Theme.Tokens.Colors.Success}}"
success_content = "{{.Theme.Tokens.Colors.SuccessContent}}"
warning = "{{.Theme.Tokens.Colors.Warning}}"
warning_content = "{{.Theme.Tokens.Colors.WarningContent}}"
error = "{{.Theme.Tokens.Colors.Error}}"
error_content = "{{.Theme.Tokens.Colors.ErrorContent}}"

[design.tokens.spacing]
selector = "{{.Theme.Tokens.Spacing.Selector}}"
field = "{{.Theme.Tokens.Spacing.Field}}"
base = "{{.Theme.Tokens.Spacing.Base}}"
sm = "{{.Theme.Tokens.Spacing.Sm}}"
md = "{{.Theme.Tokens.Spacing.Md}}"
lg = "{{.Theme.Tokens.Spacing.Lg}}"
xl = "{{.Theme.Tokens.Spacing.Xl}}"

[design.tokens.typography]
font_sans = "{{.Theme.Tokens.Typography.FontSans}}"
font_mono = "{{.Theme.Tokens.Typography.FontMono}}"
font_serif = "{{.Theme.Tokens.Typography.FontSerif}}"

[design.tokens.borders]
width = "{{.Theme.Tokens.Borders.Width}}"
radius_selector = "{{.Theme.Tokens.Borders.RadiusSelector}}"
radius_field = "{{.Theme.Tokens.Borders.RadiusField}}"
radius_box = "{{.Theme.Tokens.Borders.RadiusBox}}"

[design.tokens.shadows]
base = "{{.Theme.Tokens.Shadows.Base}}"
sm = "{{.Theme.Tokens.Shadows.Sm}}"
md = "{{.Theme.Tokens.Shadows.Md}}"
lg = "{{.Theme.Tokens.Shadows.Lg}}"
xl = "{{.Theme.Tokens.Shadows.Xl}}"

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
		Theme:          theme,
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
		"content/pages",
		"content/blog",
		"design/components/ui",
		"design/components/blocks",
		"design/layouts",
		"assets/images",
	}

	for _, dir := range exampleDirs {
		if err := os.MkdirAll(filepath.Join(sitePath, dir), 0755); err != nil {
			return fmt.Errorf("failed to create example directory %s: %w", dir, err)
		}
	}

	// 1. Home page
	homeContent := `+++
title = "Welcome to {{.Name}}"
description = "Get started with your new site"
date = "{{.Date}}"
draft = false
+++

# Welcome to {{.Name}}

This is your new site powered by our platform.

## Getting Started

1. Edit this page in the content/_index.md file
2. Create new pages in the content/ directory
3. Customize the design in the design/ directory
`

	homeContent = strings.ReplaceAll(homeContent, "{{.Name}}", cfg.Name)
	homeContent = strings.ReplaceAll(homeContent, "{{.Date}}", time.Now().Format("2006-01-02"))
	homePath := filepath.Join(sitePath, "content", "_index.md")
	if err := os.WriteFile(homePath, []byte(homeContent), 0644); err != nil {
		return fmt.Errorf("failed to create home page: %w", err)
	}

	// 2. About page
	aboutContent := `+++
title = "About {{.Name}}"
description = "Learn more about our site"
date = "{{.Date}}"
draft = false
+++

# About {{.Name}}

This is the about page for your new site. Edit this content to provide information about your site or organization.
`

	aboutContent = strings.ReplaceAll(aboutContent, "{{.Name}}", cfg.Name)
	aboutContent = strings.ReplaceAll(aboutContent, "{{.Date}}", time.Now().Format("2006-01-02"))
	aboutPath := filepath.Join(sitePath, "content", "pages", "about.md")
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
	{{ with .attributes }}{{ . | safeAttr }}{{ end }}>
	{{ with .icon }}<span class="icon">{{ . | safeHTML }}</span>{{ end }}
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
	{{ with .attributes }}{{ . | safeAttr }}{{ end }}>
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
			{{ . | safeHTML }}
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
	{{ with .attributes }}{{ . | safeAttr }}{{ end }}>
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
	layoutContent := `<!DOCTYPE html>
<html lang="en" data-theme="{{.ThemeMode}}">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>{{.Title}} | {{.Site.Name}}</title>
	<meta name="description" content="{{.Description}}">
	<link rel="stylesheet" href="/assets/css/theme.css">
	<!-- Favicon -->
	<link rel="icon" href="/assets/images/favicon.ico">
	<!-- Additional meta tags -->
	<meta name="generator" content="Wispy Core">
	<meta property="og:title" content="{{.Title}}">
	<meta property="og:description" content="{{.Description}}">
	<meta property="og:type" content="website">
	<meta property="og:url" content="{{.Site.BaseURL}}">
	<!-- Enable responsive design -->
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body class="bg-base-100 text-base-content min-h-screen flex flex-col">
	<!-- Header -->
	<header class="bg-base-200 shadow-sm">
		<div class="container mx-auto p-4">
			<nav class="flex items-center justify-between">
				<div class="flex-none">
					<a href="/" class="text-xl font-bold">{{.Site.Name}}</a>
				</div>
				<div class="flex-1 px-2 mx-2">
					<ul class="flex space-x-4">
						<li><a href="/" class="hover:text-primary">Home</a></li>
						<li><a href="/pages/about" class="hover:text-primary">About</a></li>
					</ul>
				</div>
			</nav>
		</div>
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

	<!-- Scripts -->
	<script src="/assets/js/main.js"></script>
</body>
</html>`

	layoutContent = strings.ReplaceAll(layoutContent, "{{.ThemeMode}}", cfg.ThemeMode)
	layoutContent = strings.ReplaceAll(layoutContent, "{{.CurrentYear}}", time.Now().Format("2006"))
	layoutPath := filepath.Join(sitePath, "design", "layouts", "default.html")
	if err := os.WriteFile(layoutPath, []byte(layoutContent), 0644); err != nil {
		return fmt.Errorf("failed to create default layout: %w", err)
	}

	// 5. Basic JavaScript file
	jsContent := `// Main JavaScript file

// Wait for DOM to be ready
document.addEventListener('DOMContentLoaded', function() {
  console.log('Site initialized');
  
  // Example of theme toggling functionality
  const themeToggle = document.getElementById('theme-toggle');
  if (themeToggle) {
    themeToggle.addEventListener('click', function() {
      const html = document.querySelector('html');
      const currentTheme = html.getAttribute('data-theme');
      html.setAttribute('data-theme', currentTheme === 'light' ? 'dark' : 'light');
    });
  }
});
`

	jsPath := filepath.Join(sitePath, "assets", "js", "main.js")
	if err := os.WriteFile(jsPath, []byte(jsContent), 0644); err != nil {
		return fmt.Errorf("failed to create main.js file: %w", err)
	}

	return nil
}
