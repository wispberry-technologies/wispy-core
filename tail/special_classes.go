package tail

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// handleSpecialClasses processes classes with special syntax like arbitrary values and complex variants
func handleSpecialClasses(classes []string) []string {
	result := make([]string, 0, len(classes))

	for _, class := range classes {
		// Process the class and add it to the result
		processedClass := processClass(class)
		result = append(result, processedClass)
	}

	return result
}

// processClass handles special syntax for a single class
func processClass(class string) string {
	// Handle arbitrary values like "w-[10px]"
	if strings.Contains(class, "[") && strings.Contains(class, "]") {
		// Mark as processed but keep the original syntax
		return class
	}

	// Handle complex variants like "hover:dark:bg-black"
	if strings.Contains(class, ":") {
		variants := strings.Split(class, ":")
		baseClass := variants[len(variants)-1]

		// Process the base class if needed
		processedBase := processClass(baseClass)
		if processedBase != baseClass {
			variants[len(variants)-1] = processedBase
			return strings.Join(variants, ":")
		}
	}

	return class
}

// Patterns for class extraction and categorization
var (
	// Match arbitrary values like w-[10px]
	arbitraryValuePattern = regexp.MustCompile(`\[.+?\]`)

	// Match color with opacity like bg-red-500/50
	colorOpacityPattern = regexp.MustCompile(`^(.+?)\/(\d+)$`)

	// Match responsive variants like md:flex
	responsiveVariantPattern = regexp.MustCompile(`^(sm|md|lg|xl|2xl):`)

	// Match state variants like hover:bg-blue-500
	stateVariantPattern = regexp.MustCompile(`^(hover|focus|active|disabled|group-hover|peer-hover):`)
)

// determineClassType identifies what kind of class we're dealing with
func determineClassType(class string) string {
	switch {
	case arbitraryValuePattern.MatchString(class):
		return "arbitrary"
	case colorOpacityPattern.MatchString(class):
		return "color-opacity"
	case responsiveVariantPattern.MatchString(class):
		return "responsive"
	case stateVariantPattern.MatchString(class):
		return "state"
	default:
		return "standard"
	}
}

// extractArbitraryValueFromClass extracts the value from an arbitrary class
func extractArbitraryValueFromClass(class string) string {
	matches := arbitraryValuePattern.FindStringSubmatch(class)
	if len(matches) > 0 {
		// Remove the brackets
		value := matches[0]
		return value[1 : len(value)-1]
	}
	return ""
}

// parseColorOpacity extracts the color and opacity from a class like bg-red-500/50
func parseColorOpacity(class string) (string, string) {
	matches := colorOpacityPattern.FindStringSubmatch(class)
	if len(matches) == 3 {
		return matches[1], matches[2]
	}
	return class, ""
}

// extractBaseClassFromVariant gets the base class from a variant like hover:bg-blue-500
func extractBaseClassFromVariant(class string) string {
	parts := strings.Split(class, ":")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return class
}

// handlePatternFgClass handles classes that use the --pattern-fg variable
func handlePatternFgClass(class string) ([]CSSProperty, error) {
	// Classes like [--pattern-fg:var(--color-gray-950)]/5 or dark:[--pattern-fg:var(--color-white)]/10
	if strings.Contains(class, "[--pattern-fg:") {
		// Check if it has an opacity value
		if strings.Contains(class, "]/") {
			// Extract the opacity value
			opacityStart := strings.LastIndex(class, "]/") + 2
			opacity := class[opacityStart:]

			// Extract the color variable
			start := strings.Index(class, "[--pattern-fg:") + 14
			end := strings.LastIndex(class, "]")
			colorValue := class[start:end]

			// Create CSS variable with opacity
			return []CSSProperty{{
				Name: "--pattern-fg",
				Value: fmt.Sprintf("color-mix(in srgb, %s %s%%, transparent); @supports (color: color-mix(in lab, red, red)) { --pattern-fg: color-mix(in oklab, %s %s%%, transparent); }",
					colorValue, opacity, colorValue, opacity),
			}}, nil
		}

		// Just the pattern without opacity
		start := strings.Index(class, "[--pattern-fg:") + 14
		end := strings.LastIndex(class, "]")
		colorValue := class[start:end]

		return []CSSProperty{{
			Name:  "--pattern-fg",
			Value: colorValue,
		}}, nil
	}

	// Classes that use the pattern variable like bg-(--pattern-fg)
	if strings.Contains(class, "-(--pattern-fg)") {
		prefix := strings.Split(class, "-(--pattern-fg)")[0]

		switch prefix {
		case "bg":
			return []CSSProperty{{"background-color", "var(--pattern-fg)"}}, nil
		case "border":
			return []CSSProperty{{"border-color", "var(--pattern-fg)"}}, nil
		case "border-x":
			return []CSSProperty{{"border-inline-color", "var(--pattern-fg)"}}, nil
		case "text":
			return []CSSProperty{{"color", "var(--pattern-fg)"}}, nil
		}
	}

	return nil, fmt.Errorf("unsupported pattern-fg class: %s", class)
}

// handleFontLineHeight handles classes like text-sm/7 which set both font-size and line-height
func handleFontLineHeight(class string) ([]CSSProperty, error) {
	parts := strings.Split(class, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid font-size/line-height format: %s", class)
	}

	fontSizeClass := parts[0] // e.g., text-sm
	lineHeight := parts[1]    // e.g., 7

	// Get the font size
	var fontSize string
	switch {
	case fontSizeClass == "text-xs":
		fontSize = "0.75rem"
	case fontSizeClass == "text-sm":
		fontSize = "var(--text-sm)"
	case fontSizeClass == "text-base":
		fontSize = "1rem"
	case fontSizeClass == "text-lg":
		fontSize = "1.125rem"
	case fontSizeClass == "text-xl":
		fontSize = "1.25rem"
	case fontSizeClass == "text-2xl":
		fontSize = "1.5rem"
	default:
		return nil, fmt.Errorf("unsupported font size: %s", fontSizeClass)
	}

	// Set the line height based on spacing
	var lineHeightValue string
	if _, err := strconv.Atoi(lineHeight); err == nil {
		lineHeightValue = fmt.Sprintf("calc(var(--spacing) * %s)", lineHeight)
	} else {
		lineHeightValue = lineHeight
	}

	return []CSSProperty{
		{"font-size", fontSize},
		{"line-height", lineHeightValue},
	}, nil
}

// handleColorOpacity handles classes with opacity like fill-sky-400/25
func handleColorOpacity(class string) ([]CSSProperty, error) {
	// Check for opacity format
	if !strings.Contains(class, "/") {
		return nil, fmt.Errorf("not a color opacity class: %s", class)
	}

	parts := strings.Split(class, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid color opacity format: %s", class)
	}

	baseClass := parts[0] // e.g., fill-sky-400
	opacity := parts[1]   // e.g., 25

	// Extract the property name and color
	var property string
	var colorName string

	if strings.HasPrefix(baseClass, "fill-") {
		property = "fill"
		colorName = baseClass[5:]
	} else if strings.HasPrefix(baseClass, "stroke-") {
		property = "stroke"
		colorName = baseClass[7:]
	} else if strings.HasPrefix(baseClass, "text-") {
		property = "color"
		colorName = baseClass[5:]
	} else if strings.HasPrefix(baseClass, "bg-") {
		property = "background-color"
		colorName = baseClass[3:]
	} else {
		return nil, fmt.Errorf("unsupported color property: %s", baseClass)
	}

	// Get the Tailwind color
	colorValue, ok := tailwindColorMap[colorName]
	if !ok {
		return nil, fmt.Errorf("unknown tailwind color: %s", colorName)
	}

	// Generate the CSS with both SRGB and OKLAB color mix support
	return []CSSProperty{{
		Name: property,
		Value: fmt.Sprintf("color-mix(in srgb, %s %s%%, transparent); @supports (color: color-mix(in lab, red, red)) { %s: color-mix(in oklab, %s %s%%, transparent); }",
			colorValue, opacity, property, colorValue, opacity),
	}}, nil
}

// handleComplexGridTemplate handles grid templates with arbitrary values
func handleComplexGridTemplate(class string) ([]CSSProperty, error) {
	// Handle grid-cols-[1fr_2.5rem_auto_2.5rem_1fr]
	if strings.HasPrefix(class, "grid-cols-[") && strings.HasSuffix(class, "]") {
		value := class[10 : len(class)-1]
		return []CSSProperty{{"grid-template-columns", value}}, nil
	}

	// Handle grid-rows-[1fr_1px_auto_1px_1fr]
	if strings.HasPrefix(class, "grid-rows-[") && strings.HasSuffix(class, "]") {
		value := class[10 : len(class)-1]
		return []CSSProperty{{"grid-template-rows", value}}, nil
	}

	return nil, fmt.Errorf("unsupported grid template: %s", class)
}
