package wispytail

import (
	"fmt"
	"strings"
)

// ThemeConfig represents the configuration for generating Tailwind CSS theme
type ThemeConfig struct {
	// Core design tokens
	FontFamilies    map[string]string            `json:"fontFamilies"`
	Colors          map[string]map[string]string `json:"colors"`
	DaisyUIColors   map[string]string            `json:"daisyUIColors"`
	Spacing         string                       `json:"spacing"`
	BorderRadius    map[string]string            `json:"borderRadius"`
	TextSizes       map[string]string            `json:"textSizes"`
	TextLineHeights map[string]string            `json:"textLineHeights"`
	FontWeights     map[string]string            `json:"fontWeights"`
	Breakpoints     map[string]string            `json:"breakpoints"`
	Shadows         map[string]string            `json:"shadows"`
	ZIndex          map[string]string            `json:"zIndex"`
	// Tailwind v4 specific variables
	BlurValues    map[string]string `json:"blurValues"`
	LineHeights   map[string]string `json:"lineHeights"`
	LetterSpacing map[string]string `json:"letterSpacing"`
	// Animation and transition
	Easings   map[string]string `json:"easings"`
	Durations map[string]string `json:"durations"`
	// Container sizes
	Containers map[string]string `json:"containers"`
}

// DefaultThemeConfig returns the default Tailwind v4 theme configuration
func DefaultThemeConfig() *ThemeConfig {
	return &ThemeConfig{
		FontFamilies: map[string]string{
			"sans":  "ui-sans-serif, system-ui, sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji'",
			"serif": "ui-serif, Georgia, Cambria, 'Times New Roman', Times, serif",
			"mono":  "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace",
		},
		Spacing: "0.25rem",
		DaisyUIColors: map[string]string{
			// DaisyUI semantic colors with OKLCH values
			"primary":           "oklch(0.649 0.177 259.7)",
			"primary-content":   "oklch(0.998 0.001 259.7)",
			"secondary":         "oklch(0.657 0.166 197.5)",
			"secondary-content": "oklch(0.998 0.001 197.5)",
			"accent":            "oklch(0.648 0.155 144.8)",
			"accent-content":    "oklch(0.998 0.001 144.8)",
			"neutral":           "oklch(0.321 0.02 255.5)",
			"neutral-content":   "oklch(0.998 0.001 255.5)",
			"base-100":          "oklch(1 0 0)",
			"base-200":          "oklch(0.961 0 0)",
			"base-300":          "oklch(0.922 0 0)",
			"base-content":      "oklch(0.278 0.029 256.8)",
			"info":              "oklch(0.7 0.2 220)",
			"info-content":      "oklch(0.998 0.001 220)",
			"success":           "oklch(0.65 0.25 140)",
			"success-content":   "oklch(0.998 0.001 140)",
			"warning":           "oklch(0.8 0.25 80)",
			"warning-content":   "oklch(0.2 0.05 80)",
			"error":             "oklch(0.65 0.3 30)",
			"error-content":     "oklch(0.998 0.001 30)",
		},
		Colors: map[string]map[string]string{
			"gray": {
				"50":  "oklch(0.985 0.002 247.8)",
				"100": "oklch(0.961 0.004 247.8)",
				"200": "oklch(0.898 0.009 247.8)",
				"300": "oklch(0.824 0.015 247.8)",
				"400": "oklch(0.686 0.026 247.8)",
				"500": "oklch(0.549 0.026 247.8)",
				"600": "oklch(0.457 0.026 247.8)",
				"700": "oklch(0.382 0.025 247.8)",
				"800": "oklch(0.271 0.023 247.8)",
				"900": "oklch(0.203 0.021 247.8)",
				"950": "oklch(0.128 0.015 247.8)",
			},
			"red": {
				"50":  "oklch(0.976 0.014 17.4)",
				"100": "oklch(0.942 0.032 17.4)",
				"200": "oklch(0.885 0.068 17.4)",
				"300": "oklch(0.808 0.115 17.4)",
				"400": "oklch(0.704 0.173 17.4)",
				"500": "oklch(0.637 0.237 25.3)",
				"600": "oklch(0.567 0.247 27.0)",
				"700": "oklch(0.478 0.204 27.3)",
				"800": "oklch(0.411 0.167 27.6)",
				"900": "oklch(0.367 0.134 28.2)",
				"950": "oklch(0.197 0.074 28.6)",
			},
			"blue": {
				"50":  "oklch(0.976 0.014 229.7)",
				"100": "oklch(0.942 0.032 229.7)",
				"200": "oklch(0.885 0.068 229.7)",
				"300": "oklch(0.808 0.115 229.7)",
				"400": "oklch(0.704 0.173 229.7)",
				"500": "oklch(0.637 0.237 258.3)",
				"600": "oklch(0.567 0.247 259.0)",
				"700": "oklch(0.478 0.204 259.3)",
				"800": "oklch(0.411 0.167 259.6)",
				"900": "oklch(0.367 0.134 260.2)",
				"950": "oklch(0.197 0.074 260.6)",
			},
			"sky": {
				"50":  "oklch(0.976 0.014 205.7)",
				"100": "oklch(0.942 0.032 205.7)",
				"200": "oklch(0.885 0.068 205.7)",
				"300": "oklch(0.808 0.115 205.7)",
				"400": "oklch(0.704 0.173 205.7)",
				"500": "oklch(0.637 0.237 213.3)",
				"600": "oklch(0.567 0.247 213.0)",
				"700": "oklch(0.478 0.204 213.3)",
				"800": "oklch(0.411 0.167 213.6)",
				"900": "oklch(0.367 0.134 214.2)",
				"950": "oklch(0.197 0.074 214.6)",
			},
			"green": {
				"50":  "oklch(0.976 0.014 142.0)",
				"100": "oklch(0.942 0.032 142.0)",
				"200": "oklch(0.885 0.068 142.0)",
				"300": "oklch(0.808 0.115 142.0)",
				"400": "oklch(0.704 0.173 142.0)",
				"500": "oklch(0.637 0.237 142.0)",
				"600": "oklch(0.567 0.247 142.0)",
				"700": "oklch(0.478 0.204 142.0)",
				"800": "oklch(0.411 0.167 142.0)",
				"900": "oklch(0.367 0.134 142.0)",
				"950": "oklch(0.197 0.074 142.0)",
			},
			"yellow": {
				"50":  "oklch(0.976 0.014 89.0)",
				"100": "oklch(0.942 0.032 89.0)",
				"200": "oklch(0.885 0.068 89.0)",
				"300": "oklch(0.808 0.115 89.0)",
				"400": "oklch(0.704 0.173 89.0)",
				"500": "oklch(0.637 0.237 89.0)",
				"600": "oklch(0.567 0.247 89.0)",
				"700": "oklch(0.478 0.204 89.0)",
				"800": "oklch(0.411 0.167 89.0)",
				"900": "oklch(0.367 0.134 89.0)",
				"950": "oklch(0.197 0.074 89.0)",
			},
		},
		BorderRadius: map[string]string{
			"none": "0",
			"xs":   "0.125rem",
			"sm":   "0.25rem",
			"md":   "0.375rem",
			"lg":   "0.5rem",
			"xl":   "0.75rem",
			"2xl":  "1rem",
			"3xl":  "1.5rem",
			"4xl":  "2rem",
			"full": "9999px",
		},
		TextSizes: map[string]string{
			"xs":   "0.75rem",
			"sm":   "0.875rem",
			"base": "1rem",
			"lg":   "1.125rem",
			"xl":   "1.25rem",
			"2xl":  "1.5rem",
			"3xl":  "1.875rem",
			"4xl":  "2.25rem",
			"5xl":  "3rem",
			"6xl":  "3.75rem",
			"7xl":  "4.5rem",
			"8xl":  "6rem",
			"9xl":  "8rem",
			"10xl": "10rem",
			"11xl": "12rem",
			"12xl": "14rem",
		},
		TextLineHeights: map[string]string{
			"xs":   "1rem",    // 16px
			"sm":   "1.25rem", // 20px
			"base": "1.5rem",  // 24px
			"lg":   "1.75rem", // 28px
			"xl":   "1.75rem", // 28px
			"2xl":  "2rem",    // 32px
			"3xl":  "2.25rem", // 36px
			"4xl":  "2.5rem",  // 40px
			"5xl":  "1",       // 1x
			"6xl":  "1",       // 1x
			"7xl":  "1",       // 1x
			"8xl":  "1",       // 1x
			"9xl":  "1",       // 1x
			"10xl": "1",       // 1x
			"11xl": "1",       // 1x
			"12xl": "1",       // 1x
		},
		FontWeights: map[string]string{
			"thin":       "100",
			"extralight": "200",
			"light":      "300",
			"normal":     "400",
			"medium":     "500",
			"semibold":   "600",
			"bold":       "700",
			"extrabold":  "800",
			"black":      "900",
		},
		Breakpoints: map[string]string{
			"sm":  "640px",
			"md":  "768px",
			"lg":  "1024px",
			"xl":  "1280px",
			"2xl": "1536px",
			"3xl": "1920px",
		},
		Shadows: map[string]string{
			"xs":    "0 1px 2px 0 rgba(0, 0, 0, 0.05)",
			"sm":    "0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1)",
			"md":    "0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1)",
			"lg":    "0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -4px rgba(0, 0, 0, 0.1)",
			"xl":    "0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 8px 10px -6px rgba(0, 0, 0, 0.1)",
			"2xl":   "0 25px 50px -12px rgba(0, 0, 0, 0.25)",
			"inner": "inset 0 2px 4px 0 rgba(0, 0, 0, 0.05)",
			"none":  "none",
		},
		ZIndex: map[string]string{
			"0":    "0",
			"10":   "10",
			"20":   "20",
			"30":   "30",
			"40":   "40",
			"50":   "50",
			"auto": "auto",
		},
		BlurValues: map[string]string{
			"none": "0",
			"sm":   "4px",
			"md":   "8px",
			"lg":   "16px",
			"xl":   "24px",
			"2xl":  "40px",
			"3xl":  "64px",
		},
		LineHeights: map[string]string{
			"none":    "1",
			"tight":   "1.25",
			"snug":    "1.375",
			"normal":  "1.5",
			"relaxed": "1.625",
			"loose":   "2",
		},
		LetterSpacing: map[string]string{
			"tighter": "-0.05em",
			"tight":   "-0.025em",
			"normal":  "0em",
			"wide":    "0.025em",
			"wider":   "0.05em",
			"widest":  "0.1em",
		},
		Easings: map[string]string{
			"linear": "linear",
			"in":     "cubic-bezier(0.4, 0, 1, 1)",
			"out":    "cubic-bezier(0, 0, 0.2, 1)",
			"in-out": "cubic-bezier(0.4, 0, 0.2, 1)",
		},
		Durations: map[string]string{
			"75":   "75ms",
			"100":  "100ms",
			"150":  "150ms",
			"200":  "200ms",
			"300":  "300ms",
			"500":  "500ms",
			"700":  "700ms",
			"1000": "1000ms",
		},
		Containers: map[string]string{
			"3xs": "256px",
			"2xs": "320px",
			"xs":  "480px",
			"sm":  "640px",
			"md":  "768px",
			"lg":  "1024px",
			"xl":  "1280px",
			"2xl": "1536px",
			"3xl": "1920px",
			"4xl": "2560px",
			"5xl": "3200px",
			"6xl": "3840px",
			"7xl": "5120px",
			"8xl": "6400px",
			"9xl": "7680px",
		},
	}
}

// CSS property builder utility
type CSSBuilder struct {
	lines []string
}

func NewCSSBuilder() *CSSBuilder {
	return &CSSBuilder{lines: make([]string, 0)}
}

func (b *CSSBuilder) AddSection(comment string) *CSSBuilder {
	if comment != "" {
		b.lines = append(b.lines, fmt.Sprintf("    /* %s */", comment))
	}
	return b
}

func (b *CSSBuilder) AddProperty(name, value string) *CSSBuilder {
	b.lines = append(b.lines, fmt.Sprintf("    %s: %s;", name, value))
	return b
}

func (b *CSSBuilder) AddPropertyFromMap(prefix string, props map[string]string) *CSSBuilder {
	for key, value := range props {
		b.AddProperty(fmt.Sprintf("--%s-%s", prefix, key), value)
	}
	return b
}

func (b *CSSBuilder) AddPropertyFromNestedMap(prefix string, props map[string]map[string]string) *CSSBuilder {
	for colorName, shades := range props {
		for shade, value := range shades {
			b.AddProperty(fmt.Sprintf("--%s-%s-%s", prefix, colorName, shade), value)
		}
	}
	return b
}

func (b *CSSBuilder) AddRule(selector, properties string) *CSSBuilder {
	b.lines = append(b.lines, fmt.Sprintf("  %s {", selector))
	// Split properties by semicolon and add proper indentation
	props := strings.Split(strings.TrimSpace(properties), ";")
	for _, prop := range props {
		prop = strings.TrimSpace(prop)
		if prop != "" {
			b.lines = append(b.lines, fmt.Sprintf("    %s;", prop))
		}
	}
	b.lines = append(b.lines, "  }")
	return b
}

func (b *CSSBuilder) AddLayer(layerName string, content func(*CSSBuilder)) *CSSBuilder {
	b.lines = append(b.lines, fmt.Sprintf("@layer %s {", layerName))
	content(b)
	b.lines = append(b.lines, "}")
	return b
}

func (b *CSSBuilder) AddBlock(content string) *CSSBuilder {
	lines := strings.Split(content, "\n")
	b.lines = append(b.lines, lines...)
	return b
}

func (b *CSSBuilder) String() string {
	return strings.Join(b.lines, "\n")
}

// GenerateTailwindTheme generates the complete Tailwind CSS theme layer
func GenerateThemeLayer(config *ThemeConfig) string {
	if config == nil {
		config = DefaultThemeConfig()
	}

	builder := NewCSSBuilder()

	// Define cascade layers
	builder.AddBlock("@layer theme, base, components, utilities;\n")

	// Theme layer
	builder.AddLayer("theme", func(b *CSSBuilder) {
		b.AddBlock(":root {") // Start :root block
		
		// DaisyUI semantic colors
		b.AddSection("DaisyUI semantic colors")
		b.AddPropertyFromMap("color", config.DaisyUIColors)
		b.AddBlock("")
		
		// Font families
		b.AddSection("Font families")
		b.AddPropertyFromMap("font", config.FontFamilies)
		b.AddBlock("")

		// Spacing scale
		b.AddSection("Spacing scale")
		b.AddProperty("--spacing", config.Spacing)
		b.AddBlock("")

		// Standard Tailwind colors
		b.AddSection("Standard Tailwind colors")
		b.AddPropertyFromNestedMap("color", config.Colors)
		b.AddBlock("")

		// Border radius
		b.AddSection("Border radius")
		b.AddPropertyFromMap("radius", config.BorderRadius)
		b.AddBlock("")

		// Text sizes
		b.AddSection("Text sizes")
		b.AddPropertyFromMap("text", config.TextSizes)
		b.AddBlock("")

		// Text line heights (for text size utilities)
		b.AddSection("Text line heights")
		for size, lineHeight := range config.TextLineHeights {
			b.AddProperty("--text-"+size+"--line-height", lineHeight)
		}
		b.AddBlock("")

		// Font weights
		b.AddSection("Font weights")
		b.AddPropertyFromMap("font-weight", config.FontWeights)
		b.AddBlock("")

		// Breakpoints
		b.AddSection("Breakpoints")
		b.AddPropertyFromMap("breakpoint", config.Breakpoints)
		b.AddBlock("")

		// Box shadows
		b.AddSection("Box shadows")
		b.AddPropertyFromMap("shadow", config.Shadows)
		b.AddBlock("")

		// Z-index
		b.AddSection("Z-index")
		b.AddPropertyFromMap("z", config.ZIndex)
		b.AddBlock("")

		// Blur values
		b.AddSection("Blur values")
		b.AddPropertyFromMap("blur", config.BlurValues)
		b.AddBlock("")

		// Line heights
		b.AddSection("Line heights")
		b.AddPropertyFromMap("leading", config.LineHeights)
		b.AddBlock("")

		// Letter spacing
		b.AddSection("Letter spacing")
		b.AddPropertyFromMap("tracking", config.LetterSpacing)
		b.AddBlock("")

		// Animation easings
		b.AddSection("Animation easings")
		b.AddPropertyFromMap("ease", config.Easings)
		b.AddBlock("")

		// Animation durations
		b.AddSection("Animation durations")
		b.AddPropertyFromMap("duration", config.Durations)
		b.AddBlock("")

		// Container sizes
		b.AddSection("Container sizes")
		b.AddPropertyFromMap("container", config.Containers)

		b.lines = append(b.lines[:len(b.lines)-1], "  }") // Close :root
	})

	return builder.String()
}

// GenerateTailwindBase generates the Tailwind CSS base layer with modern resets
func GenerateCssBaseLayer() string {
	builder := NewCSSBuilder()

	builder.AddLayer("base", func(b *CSSBuilder) {
		// Universal box-sizing and border reset
		b.AddRule("*, ::before, ::after",
			"box-sizing: border-box; border-width: 0; border-style: solid; border-color: var(--color-gray-200)")

		// HTML and host element styles
		b.AddRule("html, :host",
			"line-height: 1.5; -webkit-text-size-adjust: 100%; -moz-tab-size: 4; tab-size: 4; font-family: var(--font-sans); font-feature-settings: normal; font-variation-settings: normal; -webkit-tap-highlight-color: transparent")

		// Body reset
		b.AddRule("body",
			"margin: 0; line-height: inherit")

		// Heading reset
		b.AddRule("h1, h2, h3, h4, h5, h6",
			"font-size: inherit; font-weight: inherit")

		// Link reset
		b.AddRule("a",
			"color: inherit; text-decoration: inherit")

		// Form element reset
		b.AddRule("button, input, optgroup, select, textarea",
			"font-family: inherit; font-feature-settings: inherit; font-variation-settings: inherit; font-size: 100%; font-weight: inherit; line-height: inherit; color: inherit; margin: 0; padding: 0")

		// Button and select styles
		b.AddRule("button, select",
			"text-transform: none")

		// Button appearance reset
		b.AddRule("button, [type='button'], [type='reset'], [type='submit']",
			"-webkit-appearance: button; background-color: transparent; background-image: none")

		// Media element styles
		b.AddRule("img, svg, video, canvas, audio, iframe, embed, object",
			"display: block; vertical-align: middle")

		// Responsive media
		b.AddRule("img, video",
			"max-width: 100%; height: auto")

		// Table reset
		b.AddRule("table",
			"text-indent: 0; border-color: inherit; border-collapse: collapse")

		// List reset
		b.AddRule("ol, ul",
			"list-style: none; margin: 0; padding: 0")

		// Fieldset and legend reset
		b.AddRule("fieldset",
			"margin: 0; padding: 0")

		b.AddRule("legend",
			"padding: 0")

		// Textarea resize
		b.AddRule("textarea",
			"resize: vertical")

		// Placeholder color
		b.AddRule("input::placeholder, textarea::placeholder",
			"opacity: 1; color: var(--color-gray-400)")

		// Focus outline reset
		b.AddRule("button:focus, input:focus, optgroup:focus, select:focus, textarea:focus",
			"outline: 2px solid transparent; outline-offset: 2px")

		// Summary cursor
		b.AddRule("summary",
			"cursor: pointer")

		// HR styling
		b.AddRule("hr",
			"height: 0; color: inherit; border-top-width: 1px")

		// Abbr styling
		b.AddRule("abbr:where([title])",
			"text-decoration: underline dotted")

		// Code and pre styling
		b.AddRule("code, kbd, samp, pre",
			"font-family: var(--font-mono); font-size: 1em")

		// Small element
		b.AddRule("small",
			"font-size: 80%")

		// Sub and sup elements
		b.AddRule("sub, sup",
			"font-size: 75%; line-height: 0; position: relative; vertical-align: baseline")

		b.AddRule("sub",
			"bottom: -0.25em")

		b.AddRule("sup",
			"top: -0.5em")
	})

	return builder.String()
}
