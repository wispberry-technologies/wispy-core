package tail

import (
	"fmt"
	"strconv"
	"strings"
)

// semanticColorMap maps semantic color names (like daisyUI colors) to CSS variables
var semanticColorMap = map[string]string{
	// DaisyUI semantic colors
	"primary":           "--color-primary",
	"primary-content":   "--color-primary-content",
	"secondary":         "--color-secondary",
	"secondary-content": "--color-secondary-content",
	"accent":            "--color-accent",
	"accent-content":    "--color-accent-content",
	"neutral":           "--color-neutral",
	"neutral-content":   "--color-neutral-content",
	"base-100":          "--color-base-100",
	"base-200":          "--color-base-200",
	"base-300":          "--color-base-300",
	"base-content":      "--color-base-content",
	"info":              "--color-info",
	"info-content":      "--color-info-content",
	"success":           "--color-success",
	"success-content":   "--color-success-content",
	"warning":           "--color-warning",
	"warning-content":   "--color-warning-content",
	"error":             "--color-error",
	"error-content":     "--color-error-content",
}

// GetColorValue handles both standard Tailwind colors and semantic colors
func GetColorValue(colorName string) (string, bool) {
	// First check in tailwindColorMap (standard colors)
	if tailwindColor, ok := tailwindColorMap[colorName]; ok {
		return tailwindColor, true
	}

	// Then check for semantic colors (daisyUI)
	if semanticColor, ok := semanticColorMap[colorName]; ok {
		return "var(" + semanticColor + ")", true
	}

	return "", false
}

// HandleColorClass processes color-related classes like bg-primary, text-base-200, etc.
func HandleColorClass(class string) ([]CSSProperty, error) {
	// Handle dark mode variants
	isDarkMode := false
	if strings.HasPrefix(class, "dark:") {
		isDarkMode = true
		class = class[5:] // Remove 'dark:' prefix
	}

	// Extract property and color parts
	parts := strings.SplitN(class, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid color class format: %s", class)
	}

	// Determine CSS property from class prefix
	var property string
	prefix := parts[0]
	colorName := parts[1]

	switch prefix {
	case "bg":
		property = "background-color"
	case "text":
		property = "color"
	case "border":
		property = "border-color"
	case "fill":
		property = "fill"
	case "stroke":
		property = "stroke"
	case "decoration":
		property = "text-decoration-color"
	default:
		return nil, fmt.Errorf("unsupported color property prefix: %s", prefix)
	}

	// Handle color with opacity like bg-primary/50
	if strings.Contains(colorName, "/") {
		return handleColorWithOpacity(property, colorName, isDarkMode)
	}

	// Get color value
	colorValue, ok := GetColorValue(colorName)
	if ok {
		if isDarkMode {
			// Let applyVariant handle the dark mode wrapping
			baseProps := []CSSProperty{{property, colorValue}}
			return applyVariant(baseProps, "dark")
		}
		return []CSSProperty{{property, colorValue}}, nil
	}

	return nil, fmt.Errorf("unrecognized color: %s", colorName)
}

// handleColorWithOpacity processes color classes with opacity like bg-primary/50
func handleColorWithOpacity(property, colorSpec string, isDarkMode bool) ([]CSSProperty, error) {
	parts := strings.Split(colorSpec, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid color/opacity format: %s", colorSpec)
	}

	colorName := parts[0]
	opacity := parts[1]

	// Get base color value
	colorValue, ok := GetColorValue(colorName)
	if !ok {
		return nil, fmt.Errorf("unrecognized color: %s", colorName)
	}

	// Process opacity value
	opacityValue := opacity
	if opacityValue == "0" {
		opacityValue = "0"
	} else if opacityValue == "100" {
		opacityValue = "1"
	} else {
		// Convert percentage to decimal
		if val, err := strconv.Atoi(opacityValue); err == nil {
			opacityValue = fmt.Sprintf("%.2f", float64(val)/100.0)
		}
	}

	// Generate CSS with color-mix for modern browsers
	cssValue := fmt.Sprintf("%s / %s", colorValue, opacityValue)
	if isDarkMode {
		// Let applyVariant handle the dark mode wrapping
		baseProps := []CSSProperty{{property, cssValue}}
		return applyVariant(baseProps, "dark")
	}
	return []CSSProperty{{property, cssValue}}, nil
}
