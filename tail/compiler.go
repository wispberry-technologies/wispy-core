package tail

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

// extractClasses extracts all Tailwind classes from HTML content
func extractClasses(htmlContent string) ([]string, error) {
	classSet := make(map[string]struct{})

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				if attr.Key == "class" {
					// Split class names and add them to our set
					for _, className := range strings.Fields(attr.Val) {
						classSet[className] = struct{}{}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)

	// Convert set to slice
	uniqueClasses := make([]string, 0, len(classSet))
	for class := range classSet {
		uniqueClasses = append(uniqueClasses, class)
	}

	return uniqueClasses, nil
}

// Define categories for Tailwind class ordering
type classCategory int

const (
	// Categories based on Tailwind's internal ordering
	layoutCategory classCategory = iota
	positionCategory
	displayCategory
	sizingCategory
	spacingCategory
	backgroundCategory
	borderCategory
	flexGridCategory
	boxDecorationCategory
	tableCategory
	transformEffectsCategory
	transitionCategory
	typographyCategory
	interactivityCategory
	svgCategory
	accessibilityCategory
	mediaQueriesCategory
	otherCategory
)

// classOrder maps class prefixes to their category order
var classOrder = map[string]classCategory{
	// Layout
	"container": layoutCategory,
	"box-":      layoutCategory,
	"float-":    layoutCategory,
	"overflow-": layoutCategory,
	"absolute":  positionCategory,
	"relative":  positionCategory,
	"sticky":    positionCategory,
	"fixed":     positionCategory,
	"top-":      positionCategory,
	"right-":    positionCategory,
	"bottom-":   positionCategory,
	"left-":     positionCategory,
	"z-":        positionCategory,
	"inset-":    positionCategory,

	// Display & Box Model
	"block":   displayCategory,
	"inline":  displayCategory,
	"flex":    displayCategory,
	"grid":    displayCategory,
	"hidden":  displayCategory,
	"visible": displayCategory,
	"w-":      sizingCategory,
	"h-":      sizingCategory,
	"min-":    sizingCategory,
	"max-":    sizingCategory,

	// Spacing
	"p":      spacingCategory,
	"pt-":    spacingCategory,
	"pr-":    spacingCategory,
	"pb-":    spacingCategory,
	"pl-":    spacingCategory,
	"px-":    spacingCategory,
	"py-":    spacingCategory,
	"m":      spacingCategory,
	"mt-":    spacingCategory,
	"mr-":    spacingCategory,
	"mb-":    spacingCategory,
	"ml-":    spacingCategory,
	"mx-":    spacingCategory,
	"my-":    spacingCategory,
	"space-": spacingCategory,

	// Background
	"bg-": backgroundCategory,

	// Border
	"border":  borderCategory,
	"rounded": borderCategory,

	// Flex and Grid
	"flex-":    flexGridCategory,
	"grid-":    flexGridCategory,
	"col-":     flexGridCategory,
	"row-":     flexGridCategory,
	"gap-":     flexGridCategory,
	"justify-": flexGridCategory,
	"items-":   flexGridCategory,

	// Typography
	"font-":       typographyCategory,
	"text-":       typographyCategory,
	"leading-":    typographyCategory,
	"tracking-":   typographyCategory,
	"align-":      typographyCategory,
	"whitespace-": typographyCategory,
	"underline":   typographyCategory,
	"decoration-": typographyCategory,

	// Effects
	"shadow":   transformEffectsCategory,
	"opacity-": transformEffectsCategory,

	// SVG
	"fill-":   svgCategory,
	"stroke-": svgCategory,

	// Interactive states
	"hover:":  interactivityCategory,
	"focus:":  interactivityCategory,
	"active:": interactivityCategory,
	"group-":  interactivityCategory,
	"peer-":   interactivityCategory,

	// Responsive modifiers
	"sm:":   mediaQueriesCategory,
	"md:":   mediaQueriesCategory,
	"lg:":   mediaQueriesCategory,
	"xl:":   mediaQueriesCategory,
	"2xl:":  mediaQueriesCategory,
	"dark:": mediaQueriesCategory,
}

// getCategoryForClass determines the category of a Tailwind class
func getCategoryForClass(class string) classCategory {
	// Check for matching prefixes
	for prefix, category := range classOrder {
		if strings.HasPrefix(class, prefix) {
			return category
		}
	}

	// Handle special cases
	switch {
	case strings.HasPrefix(class, "p-"):
		return spacingCategory
	case strings.HasPrefix(class, "m-"):
		return spacingCategory
	}

	// Default to other category
	return otherCategory
}

// sortClasses sorts Tailwind classes according to their categories
func sortClasses(classes []string) []string {
	// Create a copy to avoid modifying the input slice
	sorted := make([]string, len(classes))
	copy(sorted, classes)

	// Sort classes by their categories and then alphabetically within each category
	sort.SliceStable(sorted, func(i, j int) bool {
		categoryI := getCategoryForClass(sorted[i])
		categoryJ := getCategoryForClass(sorted[j])

		if categoryI != categoryJ {
			return categoryI < categoryJ
		}

		// If same category, sort alphabetically
		return sorted[i] < sorted[j]
	})

	return sorted
}

// generateCSS creates optimized CSS from sorted Tailwind classes
func generateCSS(classes []string) (string, error) {
	// Process special classes (arbitrary values, complex variants)
	processedClasses := handleSpecialClasses(classes)

	// Separate daisyUI classes from regular Tailwind classes
	var tailwindClasses []string
	daisyUIClassesMap := make(map[string][]string)
	unresolvedClasses := make(map[string]struct{})

	for _, class := range processedClasses {
		if isDaisyUIClass(class) {
			baseComponent := getDaisyUIBaseComponent(class)
			daisyUIClassesMap[baseComponent] = append(daisyUIClassesMap[baseComponent], class)
		} else {
			tailwindClasses = append(tailwindClasses, class)

			// First check for special pattern classes with --pattern-fg
			if strings.Contains(class, "--pattern-fg") || strings.Contains(class, "(--pattern-fg)") {
				props, err := handlePatternFgClass(class)
				if err == nil && len(props) > 0 {
					continue // Successfully handled
				}
			}

			// Check for font size with line height (text-sm/7)
			if strings.HasPrefix(class, "text-") && strings.Contains(class, "/") {
				props, err := handleFontLineHeight(class)
				if err == nil && len(props) > 0 {
					continue // Successfully handled
				}
			}

			// Check for color with opacity (fill-sky-400/25)
			if strings.Contains(class, "/") &&
				(strings.HasPrefix(class, "fill-") ||
					strings.HasPrefix(class, "stroke-") ||
					strings.HasPrefix(class, "text-") ||
					strings.HasPrefix(class, "bg-")) {
				props, err := handleColorOpacity(class)
				if err == nil && len(props) > 0 {
					continue // Successfully handled
				}
			}

			// Check for complex grid templates
			if (strings.HasPrefix(class, "grid-cols-") || strings.HasPrefix(class, "grid-rows-")) &&
				strings.Contains(class, "[") && strings.Contains(class, "]") {
				props, err := handleComplexGridTemplate(class)
				if err == nil && len(props) > 0 {
					continue // Successfully handled
				}
			}

			// Finally, check if we have CSS rules for this class
			props, _ := getCSSPropertiesForClass(class)
			if len(props) == 0 {
				unresolvedClasses[class] = struct{}{}
			}
		}
	}

	// Log warnings for unresolved classes
	if len(unresolvedClasses) > 0 {
		fmt.Println("WARNING: The following Tailwind classes could not be resolved:")
		for class := range unresolvedClasses {
			fmt.Printf("  - %s\n", class)
		}
	}

	// Create a buffer for our CSS output
	var buffer bytes.Buffer

	buffer.WriteString("/*! tailwindcss v4.1.10 | MIT License | https://tailwindcss.com */\n")
	buffer.WriteString("@layer properties;\n")
	buffer.WriteString("@layer theme, base, components, utilities;\n")

	// Theme layer
	buffer.WriteString("@layer theme {\n")
	buffer.WriteString("  :root, :host {\n")
	buffer.WriteString("    --font-sans: ui-sans-serif, system-ui, sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol',\n")
	buffer.WriteString("    'Noto Color Emoji';\n")
	buffer.WriteString("    --font-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New',\n")
	buffer.WriteString("    monospace;\n")
	buffer.WriteString("    --color-sky-300: oklch(82.8% 0.111 230.318);\n")
	buffer.WriteString("    --color-sky-400: oklch(74.6% 0.16 232.661);\n")
	buffer.WriteString("    --color-sky-800: oklch(44.3% 0.11 240.79);\n")
	buffer.WriteString("    --color-gray-100: oklch(96.7% 0.003 264.542);\n")
	buffer.WriteString("    --color-gray-300: oklch(87.2% 0.01 258.338);\n")
	buffer.WriteString("    --color-gray-700: oklch(37.3% 0.034 259.733);\n")
	buffer.WriteString("    --color-gray-950: oklch(13% 0.028 261.692);\n")
	buffer.WriteString("    --color-white: #fff;\n")
	buffer.WriteString("    --spacing: 0.25rem;\n")
	buffer.WriteString("    --container-lg: 32rem;\n")
	buffer.WriteString("    --text-sm: 0.875rem;\n")
	buffer.WriteString("    --font-weight-medium: 500;\n")
	buffer.WriteString("    --font-weight-semibold: 600;\n")
	buffer.WriteString("    --radius-xl: 0.75rem;\n")
	buffer.WriteString("    --default-font-family: var(--font-sans);\n")
	buffer.WriteString("    --default-mono-font-family: var(--font-mono);\n")
	buffer.WriteString("  }\n")
	buffer.WriteString("}\n")

	// Base layer
	buffer.WriteString("@layer base {\n")
	buffer.WriteString("  *, ::after, ::before, ::backdrop, ::file-selector-button {\n")
	buffer.WriteString("    box-sizing: border-box;\n")
	buffer.WriteString("    margin: 0;\n")
	buffer.WriteString("    padding: 0;\n")
	buffer.WriteString("    border: 0 solid;\n")
	buffer.WriteString("  }\n")

	// Add more base styles from example-output.css
	buffer.WriteString("  html, :host {\n")
	buffer.WriteString("    line-height: 1.5;\n")
	buffer.WriteString("    -webkit-text-size-adjust: 100%;\n")
	buffer.WriteString("    tab-size: 4;\n")
	buffer.WriteString("    font-family: var(--default-font-family, ui-sans-serif, system-ui, sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol', 'Noto Color Emoji');\n")
	buffer.WriteString("    font-feature-settings: var(--default-font-feature-settings, normal);\n")
	buffer.WriteString("    font-variation-settings: var(--default-font-variation-settings, normal);\n")
	buffer.WriteString("    -webkit-tap-highlight-color: transparent;\n")
	buffer.WriteString("  }\n")
	buffer.WriteString("}\n")

	// Components layer (for daisyUI)
	buffer.WriteString("@layer components {\n")
	if len(daisyUIClassesMap) > 0 {
		daisyUICSS := generateDaisyUICSS(daisyUIClassesMap)
		buffer.WriteString(daisyUICSS)
	}
	buffer.WriteString("}\n")

	// Utilities layer - generate CSS for each Tailwind class
	buffer.WriteString("@layer utilities {\n")

	// Create a map to deduplicate identical CSS rules
	cssRules := make(map[string]bool)

	// For each tailwind class, generate the CSS
	for _, class := range tailwindClasses {
		// Skip if class is in the unresolved list
		if _, ok := unresolvedClasses[class]; ok {
			continue
		}

		// Get CSS properties for this class
		properties, err := getCSSPropertiesForClass(class)
		if err != nil {
			// Skip this class if there was an error
			continue
		}

		if len(properties) == 0 {
			// No properties found for this class
			continue
		}

		// Generate CSS for this class
		var cssRule strings.Builder
		cssRule.WriteString("  .")

		// Extract variant (for hover:, focus:, etc.)
		variant := ""

		// Check for variants like hover:, focus:, etc.
		if strings.Contains(class, ":") {
			parts := strings.Split(class, ":")
			if len(parts) > 1 {
				// For hover:bg-red-500, variant="hover"
				variant = parts[0]
			}
		}

		// Escape special characters in class names for CSS selectors
		escapedClass := strings.ReplaceAll(class, ":", "\\:")
		escapedClass = strings.ReplaceAll(escapedClass, "%", "\\%")
		escapedClass = strings.ReplaceAll(escapedClass, "/", "\\/")
		escapedClass = strings.ReplaceAll(escapedClass, ".", "\\.")
		escapedClass = strings.ReplaceAll(escapedClass, "[", "\\[")
		escapedClass = strings.ReplaceAll(escapedClass, "]", "\\]")
		escapedClass = strings.ReplaceAll(escapedClass, ")", "\\)")
		escapedClass = strings.ReplaceAll(escapedClass, "(", "\\(")

		// Add pseudo-class to selector for variants
		pseudoClass := ""
		if variant == "hover" {
			pseudoClass = ":hover"
		} else if variant == "focus" {
			pseudoClass = ":focus"
		} else if variant == "active" {
			pseudoClass = ":active"
		}

		cssRule.WriteString(escapedClass)
		if pseudoClass != "" {
			cssRule.WriteString(pseudoClass)
		}
		cssRule.WriteString(" { ")

		// Add each property
		for _, prop := range properties {
			// Check if this is a media query or a complex CSS block
			if strings.HasPrefix(prop.Name, "@media") {
				// This is a media query
				cssRule.WriteString(prop.Name)
				cssRule.WriteString(" { ")
				cssRule.WriteString(prop.Value)
				cssRule.WriteString(" }")
			} else if strings.Contains(prop.Value, "\n") || strings.Contains(prop.Value, "{") {
				// This is a complex property, probably containing nested rules
				cssRule.WriteString(prop.Name)
				cssRule.WriteString(" { ")
				cssRule.WriteString(prop.Value)
				cssRule.WriteString(" } ")
			} else {
				// This is a simple property
				cssRule.WriteString(prop.Name)
				cssRule.WriteString(": ")
				cssRule.WriteString(prop.Value)
				cssRule.WriteString("; ")
			}
		}

		cssRule.WriteString("}\n")

		// Add to output only if we haven't seen this rule before
		rule := cssRule.String()
		if !cssRules[rule] {
			buffer.WriteString(rule)
			cssRules[rule] = true
		}
	}

	buffer.WriteString("}\n")

	// Add properties layer
	buffer.WriteString("@property --tw-space-y-reverse {\n")
	buffer.WriteString("  syntax: \"*\";\n")
	buffer.WriteString("  inherits: false;\n")
	buffer.WriteString("  initial-value: 0;\n")
	buffer.WriteString("}\n")

	buffer.WriteString("@property --tw-space-x-reverse {\n")
	buffer.WriteString("  syntax: \"*\";\n")
	buffer.WriteString("  inherits: false;\n")
	buffer.WriteString("  initial-value: 0;\n")
	buffer.WriteString("}\n")

	buffer.WriteString("@property --tw-border-style {\n")
	buffer.WriteString("  syntax: \"*\";\n")
	buffer.WriteString("  inherits: false;\n")
	buffer.WriteString("  initial-value: solid;\n")
	buffer.WriteString("}\n")

	return buffer.String(), nil
}
