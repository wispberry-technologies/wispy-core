package tail

import (
	"fmt"
	"strconv"
	"strings"
)

// CSSProperty represents a CSS property with its value
type CSSProperty struct {
	Name  string
	Value string
}

// CSS utilities and mappings

// tailwindColorMap maps tailwind color names to CSS variables
var tailwindColorMap = map[string]string{
	// Standard Tailwind colors
	"white":    "var(--color-white)",
	"gray-100": "var(--color-gray-100)",
	"gray-300": "var(--color-gray-300)",
	"gray-700": "var(--color-gray-700)",
	"gray-950": "var(--color-gray-950)",
	"sky-300":  "var(--color-sky-300)",
	"sky-400":  "var(--color-sky-400)",
	"sky-800":  "var(--color-sky-800)",

	// DaisyUI semantic colors
	"primary":           "var(--color-primary)",
	"primary-content":   "var(--color-primary-content)",
	"secondary":         "var(--color-secondary)",
	"secondary-content": "var(--color-secondary-content)",
	"accent":            "var(--color-accent)",
	"accent-content":    "var(--color-accent-content)",
	"neutral":           "var(--color-neutral)",
	"neutral-content":   "var(--color-neutral-content)",
	"base-100":          "var(--color-base-100)",
	"base-200":          "var(--color-base-200)",
	"base-300":          "var(--color-base-300)",
	"base-content":      "var(--color-base-content)",
	"info":              "var(--color-info)",
	"info-content":      "var(--color-info-content)",
	"success":           "var(--color-success)",
	"success-content":   "var(--color-success-content)",
	"warning":           "var(--color-warning)",
	"warning-content":   "var(--color-warning-content)",
	"error":             "var(--color-error)",
	"error-content":     "var(--color-error-content)",
}

// tailwindSizeMap maps tailwind size utilities to CSS values
// var tailwindSizeMap = map[string]string{
// 	"xs": "0.75rem",
// 	"sm": "0.875rem",
// 	"md": "1rem",
// 	"lg": "1.125rem",
// 	"xl": "1.25rem",
// }

// tailwindSpacingMultiplier maps tailwind spacing utilities to multipliers
var tailwindSpacingMultiplier = map[string]float64{
	"px": 0.0625, // 1px
	"0":  0,
	"1":  0.25,
	"2":  0.5,
	"3":  0.75,
	"4":  1,
	"5":  1.25,
	"6":  1.5,
	"8":  2,
	"10": 2.5,
	"12": 3,
}

// getCSSPropertiesForClass generates CSS properties for a given Tailwind class
func getCSSPropertiesForClass(class string) ([]CSSProperty, error) {
	// First check for missing properties that were added separately
	if props, found := handleMissingCSSProperties(class); found {
		return props, nil
	}

	// Check for arbitrary values in the class
	if strings.Contains(class, "[") && strings.Contains(class, "]") {
		// Extract property and value from arbitrary class
		return handleArbitraryValue(class)
	}

	// Handle responsive and pseudo-class variants
	if strings.Contains(class, ":") {
		parts := strings.Split(class, ":")
		if len(parts) > 1 {
			baseClass := parts[len(parts)-1]
			variant := strings.Join(parts[:len(parts)-1], ":")

			// Get properties for the base class
			props, err := getCSSPropertiesForClass(baseClass)
			if err != nil {
				return nil, err
			}

			// Apply variant wrapper
			return applyVariant(props, variant)
		}
	}

	// Handle common Tailwind utility classes
	switch {
	// Positioning
	case class == "relative":
		return []CSSProperty{{"position", "relative"}}, nil
	case class == "absolute":
		return []CSSProperty{{"position", "absolute"}}, nil
	case class == "fixed":
		return []CSSProperty{{"position", "fixed"}}, nil
	case class == "sticky":
		return []CSSProperty{{"position", "sticky"}}, nil
	case strings.HasPrefix(class, "top-"):
		value := class[4:]
		if value == "px" {
			return []CSSProperty{{"top", "1px"}}, nil
		}
		return handleSpacingProperty("top", value)
	case strings.HasPrefix(class, "right-"):
		value := class[6:]
		if value == "px" {
			return []CSSProperty{{"right", "1px"}}, nil
		}
		return handleSpacingProperty("right", value)
	case strings.HasPrefix(class, "bottom-"):
		value := class[7:]
		if value == "px" {
			return []CSSProperty{{"bottom", "1px"}}, nil
		}
		return handleSpacingProperty("bottom", value)
	case strings.HasPrefix(class, "left-"):
		value := class[5:]
		if value == "px" {
			return []CSSProperty{{"left", "1px"}}, nil
		}
		return handleSpacingProperty("left", value)
	case strings.HasPrefix(class, "-top-"):
		value := class[5:]
		if value == "px" {
			return []CSSProperty{{"top", "-1px"}}, nil
		}
		return handleNegativeSpacingProperty("top", value)
	case strings.HasPrefix(class, "-right-"):
		value := class[7:]
		if value == "px" {
			return []CSSProperty{{"right", "-1px"}}, nil
		}
		return handleNegativeSpacingProperty("right", value)
	case strings.HasPrefix(class, "-bottom-"):
		value := class[8:]
		if value == "px" {
			return []CSSProperty{{"bottom", "-1px"}}, nil
		}
		return handleNegativeSpacingProperty("bottom", value)
	case strings.HasPrefix(class, "-left-"):
		value := class[6:]
		if value == "px" {
			return []CSSProperty{{"left", "-1px"}}, nil
		}
		return handleNegativeSpacingProperty("left", value)
	case strings.HasPrefix(class, "inset-"):
		value := class[6:]
		if value == "0" {
			return []CSSProperty{{"inset", "0"}}, nil
		}
		return handleSpacingProperty("inset", value)

	// Z-index
	case strings.HasPrefix(class, "z-"):
		value := class[2:]
		if value == "0" {
			return []CSSProperty{{"z-index", "0"}}, nil
		}
		if val, err := strconv.Atoi(value); err == nil {
			return []CSSProperty{{"z-index", strconv.Itoa(val)}}, nil
		}

	// Grid positioning
	case class == "col-span-full":
		return []CSSProperty{{"grid-column", "1 / -1"}}, nil
	case strings.HasPrefix(class, "col-span-"):
		value := class[9:]
		if val, err := strconv.Atoi(value); err == nil {
			return []CSSProperty{{"grid-column", fmt.Sprintf("span %d / span %d", val, val)}}, nil
		}
	case strings.HasPrefix(class, "col-start-"):
		value := class[10:]
		if val, err := strconv.Atoi(value); err == nil {
			return []CSSProperty{{"grid-column-start", strconv.Itoa(val)}}, nil
		}
	case strings.HasPrefix(class, "col-end-"):
		value := class[8:]
		if val, err := strconv.Atoi(value); err == nil {
			return []CSSProperty{{"grid-column-end", strconv.Itoa(val)}}, nil
		}
	case class == "row-span-full":
		return []CSSProperty{{"grid-row", "1 / -1"}}, nil
	case strings.HasPrefix(class, "row-span-"):
		value := class[9:]
		if val, err := strconv.Atoi(value); err == nil {
			return []CSSProperty{{"grid-row", fmt.Sprintf("span %d / span %d", val, val)}}, nil
		}
	case strings.HasPrefix(class, "row-start-"):
		value := class[10:]
		if val, err := strconv.Atoi(value); err == nil {
			return []CSSProperty{{"grid-row-start", strconv.Itoa(val)}}, nil
		}
	case strings.HasPrefix(class, "row-end-"):
		value := class[8:]
		if val, err := strconv.Atoi(value); err == nil {
			return []CSSProperty{{"grid-row-end", strconv.Itoa(val)}}, nil
		}

	// Display
	case class == "block":
		return []CSSProperty{{"display", "block"}}, nil
	case class == "inline":
		return []CSSProperty{{"display", "inline"}}, nil
	case class == "inline-block":
		return []CSSProperty{{"display", "inline-block"}}, nil
	case class == "flex":
		return []CSSProperty{{"display", "flex"}}, nil
	case class == "inline-flex":
		return []CSSProperty{{"display", "inline-flex"}}, nil
	case class == "grid":
		return []CSSProperty{{"display", "grid"}}, nil
	case class == "inline-grid":
		return []CSSProperty{{"display", "inline-grid"}}, nil
	case class == "hidden":
		return []CSSProperty{{"display", "none"}}, nil
	case class == "not-dark:hidden":
		return []CSSProperty{{"@media not (prefers-color-scheme: dark)", "display: none;"}}, nil
	case class == "dark:hidden":
		return []CSSProperty{{"@media (prefers-color-scheme: dark)", "display: none;"}}, nil

	// Flex properties
	case class == "flex-row":
		return []CSSProperty{{"flex-direction", "row"}}, nil
	case class == "flex-col":
		return []CSSProperty{{"flex-direction", "column"}}, nil
	case class == "flex-row-reverse":
		return []CSSProperty{{"flex-direction", "row-reverse"}}, nil
	case class == "flex-col-reverse":
		return []CSSProperty{{"flex-direction", "column-reverse"}}, nil
	case class == "flex-wrap":
		return []CSSProperty{{"flex-wrap", "wrap"}}, nil
	case class == "flex-nowrap":
		return []CSSProperty{{"flex-wrap", "nowrap"}}, nil
	case class == "flex-wrap-reverse":
		return []CSSProperty{{"flex-wrap", "wrap-reverse"}}, nil
	case class == "flex-1":
		return []CSSProperty{{"flex", "1 1 0%"}}, nil
	case class == "flex-auto":
		return []CSSProperty{{"flex", "1 1 auto"}}, nil
	case class == "flex-initial":
		return []CSSProperty{{"flex", "0 1 auto"}}, nil
	case class == "flex-none":
		return []CSSProperty{{"flex", "none"}}, nil
	case class == "grow":
		return []CSSProperty{{"flex-grow", "1"}}, nil
	case class == "grow-0":
		return []CSSProperty{{"flex-grow", "0"}}, nil
	case class == "shrink":
		return []CSSProperty{{"flex-shrink", "1"}}, nil
	case class == "shrink-0":
		return []CSSProperty{{"flex-shrink", "0"}}, nil

	// Width and height
	case strings.HasPrefix(class, "w-"):
		value := class[2:]
		return handleSizeProperty("width", value)
	case strings.HasPrefix(class, "h-"):
		value := class[2:]
		return handleSizeProperty("height", value)
	case strings.HasPrefix(class, "min-w-"):
		value := class[6:]
		return handleSizeProperty("min-width", value)
	case strings.HasPrefix(class, "min-h-"):
		value := class[6:]
		return handleSizeProperty("min-height", value)
	case strings.HasPrefix(class, "max-w-"):
		value := class[6:]
		return handleSizeProperty("max-width", value)
	case strings.HasPrefix(class, "max-h-"):
		value := class[6:]
		return handleSizeProperty("max-height", value)

	// Grid template
	case strings.HasPrefix(class, "grid-cols-"):
		if strings.Contains(class, "[") && strings.Contains(class, "]") {
			// Extract the grid-template-columns content
			start := strings.Index(class, "[")
			end := strings.LastIndex(class, "]")
			if start != -1 && end != -1 && start < end {
				value := class[start+1 : end]
				return []CSSProperty{{"grid-template-columns", value}}, nil
			}
		}
	case strings.HasPrefix(class, "grid-rows-"):
		if strings.Contains(class, "[") && strings.Contains(class, "]") {
			// Extract the grid-template-rows content
			start := strings.Index(class, "[")
			end := strings.LastIndex(class, "]")
			if start != -1 && end != -1 && start < end {
				value := class[start+1 : end]
				return []CSSProperty{{"grid-template-rows", value}}, nil
			}
		}

	// Padding
	case strings.HasPrefix(class, "p-"):
		value := class[2:]
		return handleSpacingProperty("padding", value)
	case strings.HasPrefix(class, "pt-"):
		value := class[3:]
		return handleSpacingProperty("padding-top", value)
	case strings.HasPrefix(class, "pr-"):
		value := class[3:]
		return handleSpacingProperty("padding-right", value)
	case strings.HasPrefix(class, "pb-"):
		value := class[3:]
		return handleSpacingProperty("padding-bottom", value)
	case strings.HasPrefix(class, "pl-"):
		value := class[3:]
		return handleSpacingProperty("padding-left", value)
	case strings.HasPrefix(class, "px-"):
		value := class[3:]
		return []CSSProperty{
			{"padding-left", fmt.Sprintf("calc(var(--spacing) * %v)", tailwindSpacingMultiplier[value])},
			{"padding-right", fmt.Sprintf("calc(var(--spacing) * %v)", tailwindSpacingMultiplier[value])},
		}, nil
	case strings.HasPrefix(class, "py-"):
		value := class[3:]
		return []CSSProperty{
			{"padding-top", fmt.Sprintf("calc(var(--spacing) * %v)", tailwindSpacingMultiplier[value])},
			{"padding-bottom", fmt.Sprintf("calc(var(--spacing) * %v)", tailwindSpacingMultiplier[value])},
		}, nil

	// Margin
	case strings.HasPrefix(class, "m-"):
		value := class[2:]
		return handleSpacingProperty("margin", value)
	case strings.HasPrefix(class, "mt-"):
		value := class[3:]
		return handleSpacingProperty("margin-top", value)
	case strings.HasPrefix(class, "mr-"):
		value := class[3:]
		return handleSpacingProperty("margin-right", value)
	case strings.HasPrefix(class, "mb-"):
		value := class[3:]
		return handleSpacingProperty("margin-bottom", value)
	case strings.HasPrefix(class, "ml-"):
		value := class[3:]
		return handleSpacingProperty("margin-left", value)
	case strings.HasPrefix(class, "mx-"):
		value := class[3:]
		return []CSSProperty{
			{"margin-left", fmt.Sprintf("calc(var(--spacing) * %v)", tailwindSpacingMultiplier[value])},
			{"margin-right", fmt.Sprintf("calc(var(--spacing) * %v)", tailwindSpacingMultiplier[value])},
		}, nil
	case strings.HasPrefix(class, "my-"):
		value := class[3:]
		if value == "6" {
			return []CSSProperty{{"margin-block", "calc(var(--spacing) * 6)"}}, nil
		}
		return []CSSProperty{
			{"margin-top", fmt.Sprintf("calc(var(--spacing) * %v)", tailwindSpacingMultiplier[value])},
			{"margin-bottom", fmt.Sprintf("calc(var(--spacing) * %v)", tailwindSpacingMultiplier[value])},
		}, nil

	// Spacing for elements
	case strings.HasPrefix(class, "space-y-"):
		value := class[8:]
		return []CSSProperty{
			{":where(& > :not(:last-child))", fmt.Sprintf(`
    --tw-space-y-reverse: 0;
    margin-block-start: calc(calc(var(--spacing) * %v) * calc(1 - var(--tw-space-y-reverse)));
    margin-block-end: calc(calc(var(--spacing) * %v) * var(--tw-space-y-reverse));`, value, value)},
		}, nil
	case strings.HasPrefix(class, "space-x-"):
		value := class[8:]
		return []CSSProperty{
			{":where(& > :not(:last-child))", fmt.Sprintf(`
    --tw-space-x-reverse: 0;
    margin-inline-start: calc(calc(var(--spacing) * %v) * calc(1 - var(--tw-space-x-reverse)));
    margin-inline-end: calc(calc(var(--spacing) * %v) * var(--tw-space-x-reverse));`, value, value)},
		}, nil

	// Border
	case class == "border":
		return []CSSProperty{{"border-width", "1px"}}, nil
	case class == "border-0":
		return []CSSProperty{{"border-width", "0px"}}, nil
	case class == "border-2":
		return []CSSProperty{{"border-width", "2px"}}, nil
	case class == "border-4":
		return []CSSProperty{{"border-width", "4px"}}, nil
	case class == "border-t":
		return []CSSProperty{{"border-top-width", "1px"}}, nil
	case class == "border-r":
		return []CSSProperty{{"border-right-width", "1px"}}, nil
	case class == "border-b":
		return []CSSProperty{{"border-bottom-width", "1px"}}, nil
	case class == "border-l":
		return []CSSProperty{{"border-left-width", "1px"}}, nil
	case class == "border-x":
		return []CSSProperty{
			{"border-left-width", "1px"},
			{"border-right-width", "1px"},
		}, nil
	case class == "border-y":
		return []CSSProperty{
			{"border-top-width", "1px"},
			{"border-bottom-width", "1px"},
		}, nil
	case strings.HasPrefix(class, "border-"):
		// Skip width-related classes that we handle elsewhere
		if class == "border-0" || class == "border-2" || class == "border-4" {
			break
		}
		// Handle directional border width classes
		if class == "border-t" || class == "border-r" || class == "border-b" || class == "border-l" ||
			class == "border-x" || class == "border-y" {
			break
		}
		// Use the new color handler for border colors
		return HandleColorClass(class)

	// Border radius
	case class == "rounded":
		return []CSSProperty{{"border-radius", "0.25rem"}}, nil
	case class == "rounded-sm":
		return []CSSProperty{{"border-radius", "0.125rem"}}, nil
	case class == "rounded-md":
		return []CSSProperty{{"border-radius", "0.375rem"}}, nil
	case class == "rounded-lg":
		return []CSSProperty{{"border-radius", "0.5rem"}}, nil
	case class == "rounded-xl":
		return []CSSProperty{{"border-radius", "var(--radius-xl)"}}, nil
	case class == "rounded-2xl":
		return []CSSProperty{{"border-radius", "1rem"}}, nil
	case class == "rounded-3xl":
		return []CSSProperty{{"border-radius", "1.5rem"}}, nil
	case class == "rounded-full":
		return []CSSProperty{{"border-radius", "9999px"}}, nil
	case class == "rounded-none":
		return []CSSProperty{{"border-radius", "0"}}, nil
	case class == "rounded-t":
		return []CSSProperty{
			{"border-top-left-radius", "0.25rem"},
			{"border-top-right-radius", "0.25rem"},
		}, nil
	case class == "rounded-r":
		return []CSSProperty{
			{"border-top-right-radius", "0.25rem"},
			{"border-bottom-right-radius", "0.25rem"},
		}, nil
	case class == "rounded-b":
		return []CSSProperty{
			{"border-bottom-right-radius", "0.25rem"},
			{"border-bottom-left-radius", "0.25rem"},
		}, nil
	case class == "rounded-l":
		return []CSSProperty{
			{"border-top-left-radius", "0.25rem"},
			{"border-bottom-left-radius", "0.25rem"},
		}, nil

	// Background
	case strings.HasPrefix(class, "bg-"):
		// Use the new color handler for background colors
		return HandleColorClass(class)

	// Background attachment and size
	case class == "bg-fixed":
		return []CSSProperty{{"background-attachment", "fixed"}}, nil
	case class == "bg-local":
		return []CSSProperty{{"background-attachment", "local"}}, nil
	case class == "bg-scroll":
		return []CSSProperty{{"background-attachment", "scroll"}}, nil
	case strings.HasPrefix(class, "bg-[size:"):
		start := strings.Index(class, "[size:")
		end := strings.LastIndex(class, "]")
		if start != -1 && end != -1 && start < end {
			value := class[start+6 : end]
			return []CSSProperty{{"background-size", value}}, nil
		}
	case strings.HasPrefix(class, "bg-[image:"):
		start := strings.Index(class, "[image:")
		end := strings.LastIndex(class, "]")
		if start != -1 && end != -1 && start < end {
			value := class[start+7 : end]
			return []CSSProperty{{"background-image", value}}, nil
		}

	// Border color
	case strings.HasPrefix(class, "border-(--"):
		start := strings.Index(class, "(--")
		end := strings.LastIndex(class, ")")
		if start != -1 && end != -1 && start < end {
			value := class[start : end+1]
			return []CSSProperty{{"border-color", "var" + value}}, nil
		}
	case strings.HasPrefix(class, "border-x-(--"):
		start := strings.Index(class, "(--")
		end := strings.LastIndex(class, ")")
		if start != -1 && end != -1 && start < end {
			value := class[start : end+1]
			return []CSSProperty{{"border-inline-color", "var" + value}}, nil
		}

	// Custom properties and patterns
	case strings.HasPrefix(class, "[--"):
		end := strings.Index(class, "]")
		if end != -1 {
			value := class[1:end]
			if strings.Contains(class, "]/") {
				// It's a color opacity pattern like [--pattern-fg:var(--color-gray-950)]/5
				opacity := class[end+2:]
				if _, err := strconv.Atoi(opacity); err == nil {
					return []CSSProperty{
						{value, fmt.Sprintf("color-mix(in srgb, oklch(13%% 0.028 261.692) %s%%, transparent)", opacity)},
						{"@supports (color: color-mix(in lab, red, red))",
							fmt.Sprintf("%s: color-mix(in oklab, var(--color-gray-950) %s%%, transparent);", value, opacity)},
					}, nil
				}
			}
			return []CSSProperty{{value, "var(--" + value[4:] + ")"}}, nil
		}

	// Text decoration offset
	case class == "underline-offset-1":
		return []CSSProperty{{"text-underline-offset", "1px"}}, nil
	case class == "underline-offset-2":
		return []CSSProperty{{"text-underline-offset", "2px"}}, nil
	case class == "underline-offset-3":
		return []CSSProperty{{"text-underline-offset", "3px"}}, nil
	case class == "underline-offset-4":
		return []CSSProperty{{"text-underline-offset", "4px"}}, nil
	case class == "underline-offset-8":
		return []CSSProperty{{"text-underline-offset", "8px"}}, nil

	// Text decoration color
	case strings.HasPrefix(class, "decoration-"):
		// Handle decoration thickness cases
		if class == "decoration-0" {
			return []CSSProperty{{"text-decoration-thickness", "0px"}}, nil
		} else if class == "decoration-1" {
			return []CSSProperty{{"text-decoration-thickness", "1px"}}, nil
		} else if class == "decoration-2" {
			return []CSSProperty{{"text-decoration-thickness", "2px"}}, nil
		} else if class == "decoration-4" {
			return []CSSProperty{{"text-decoration-thickness", "4px"}}, nil
		} else if class == "decoration-8" {
			return []CSSProperty{{"text-decoration-thickness", "8px"}}, nil
		}

		// Handle decoration color using new color handler
		return HandleColorClass(class)

	// Max width cases
	case class == "max-w-lg":
		return []CSSProperty{{"max-width", "var(--container-lg)"}}, nil

	// Height special cases
	case class == "h-px":
		return []CSSProperty{{"height", "1px"}}, nil

	// Text decoration
	case class == "underline":
		return []CSSProperty{{"text-decoration-line", "underline"}}, nil

	// Text color
	case strings.HasPrefix(class, "text-"):
		// Don't handle text alignment here - only colors
		if class == "text-left" || class == "text-center" || class == "text-right" || class == "text-justify" {
			break
		}
		// Use the new color handler for text colors
		return HandleColorClass(class)

	// SVG fill
	case strings.HasPrefix(class, "fill-"):
		// Use the new color handler for SVG fill colors
		return HandleColorClass(class)

	// SVG stroke
	case strings.HasPrefix(class, "stroke-"):
		// Use the new color handler for SVG stroke colors
		return HandleColorClass(class)

	// Dark mode variants
	case strings.HasPrefix(class, "dark:"):
		// We'll try to handle all dark mode variants with our color handler first
		if strings.HasPrefix(class[5:], "text-") ||
			strings.HasPrefix(class[5:], "bg-") ||
			strings.HasPrefix(class[5:], "border-") ||
			strings.HasPrefix(class[5:], "fill-") ||
			strings.HasPrefix(class[5:], "stroke-") ||
			strings.HasPrefix(class[5:], "decoration-") {
			return HandleColorClass(class)
		}
	case class == "h-px":
		return []CSSProperty{{"height", "1px"}}, nil
	case class == "underline":
		return []CSSProperty{{"text-decoration-line", "underline"}}, nil
	case class == "font-medium":
		return []CSSProperty{
			{"--tw-font-weight", "var(--font-weight-medium)"},
			{"font-weight", "var(--font-weight-medium)"},
		}, nil
	case class == "font-semibold":
		return []CSSProperty{
			{"--tw-font-weight", "var(--font-weight-semibold)"},
			{"font-weight", "var(--font-weight-semibold)"},
		}, nil
	case class == "text-gray-950":
		return []CSSProperty{{"color", "var(--color-gray-950)"}}, nil
	case class == "text-gray-700":
		return []CSSProperty{{"color", "var(--color-gray-700)"}}, nil
	case class == "stroke-sky-800":
		return []CSSProperty{{"stroke", "var(--color-sky-800)"}}, nil
	case class == "dark:stroke-sky-300":
		return []CSSProperty{{"@media (prefers-color-scheme: dark)", "stroke: var(--color-sky-300);"}}, nil
	case class == "dark:text-gray-300":
		return []CSSProperty{{"@media (prefers-color-scheme: dark)", "color: var(--color-gray-300);"}}, nil
	case class == "dark:text-white":
		return []CSSProperty{{"@media (prefers-color-scheme: dark)", "color: var(--color-white);"}}, nil
	}

	// Handle arbitrary values
	if strings.Contains(class, "[") && strings.Contains(class, "]") {
		return handleArbitraryValue(class)
	}

	// Return empty result for unhandled classes
	return []CSSProperty{}, nil
}

// applyVariant wraps CSS properties with the appropriate variant pseudo-class or media query
func applyVariant(properties []CSSProperty, variant string) ([]CSSProperty, error) {
	result := make([]CSSProperty, 0, len(properties))

	// Store the pseudo-class to be added to selector in the CSS output
	pseudoClass := ""
	switch variant {
	case "hover":
		pseudoClass = ":hover"
	case "focus":
		pseudoClass = ":focus"
	case "active":
		pseudoClass = ":active"
	}

	// If we have a pseudo-class, use it for all properties
	if pseudoClass != "" {
		for _, prop := range properties {
			// We still need to return a standard CSSProperty, but we'll modify
			// how these are handled in the CSS output
			result = append(result, CSSProperty{
				// Store special handling info in global context
				Name:  prop.Name,
				Value: prop.Value,
			})
		}
		return result, nil
	}

	// Handle media queries and other variants
	switch variant {
	case "dark":
		// Apply dark mode variant to all properties
		for _, prop := range properties {
			result = append(result, CSSProperty{
				Name:  "@media (prefers-color-scheme: dark)",
				Value: fmt.Sprintf("%s: %s;", prop.Name, prop.Value),
			})
		}
	case "not-dark":
		// Apply not dark mode variant to all properties
		for _, prop := range properties {
			result = append(result, CSSProperty{
				Name:  "@media not (prefers-color-scheme: dark)",
				Value: fmt.Sprintf("%s: %s;", prop.Name, prop.Value),
			})
		}
	case "sm":
		// Apply small screen variant
		for _, prop := range properties {
			result = append(result, CSSProperty{
				Name:  "@media (min-width: 640px)",
				Value: fmt.Sprintf("%s: %s;", prop.Name, prop.Value),
			})
		}
	case "md":
		// Apply medium screen variant
		for _, prop := range properties {
			result = append(result, CSSProperty{
				Name:  "@media (min-width: 768px)",
				Value: fmt.Sprintf("%s: %s;", prop.Name, prop.Value),
			})
		}
	case "lg":
		// Apply large screen variant
		for _, prop := range properties {
			result = append(result, CSSProperty{
				Name:  "@media (min-width: 1024px)",
				Value: fmt.Sprintf("%s: %s;", prop.Name, prop.Value),
			})
		}
	case "xl":
		// Apply extra-large screen variant
		for _, prop := range properties {
			result = append(result, CSSProperty{
				Name:  "@media (min-width: 1280px)",
				Value: fmt.Sprintf("%s: %s;", prop.Name, prop.Value),
			})
		}
	case "2xl":
		// Apply 2x extra-large screen variant
		for _, prop := range properties {
			result = append(result, CSSProperty{
				Name:  "@media (min-width: 1536px)",
				Value: fmt.Sprintf("%s: %s;", prop.Name, prop.Value),
			})
		}
	default:
		// For compound variants like hover:focus:, recursively apply each variant
		if strings.Contains(variant, ":") {
			parts := strings.Split(variant, ":")
			if len(parts) > 1 {
				lastVariant := parts[len(parts)-1]
				remainingVariants := strings.Join(parts[:len(parts)-1], ":")

				// Apply the last variant first
				intermediate, err := applyVariant(properties, lastVariant)
				if err != nil {
					return nil, err
				}

				// Then apply the remaining variants
				return applyVariant(intermediate, remainingVariants)
			}
		}

		// If it's an unknown variant, just pass through the properties
		result = properties
	}

	return result, nil
}

// handleSpacingProperty generates CSS for spacing properties
func handleSpacingProperty(property, value string) ([]CSSProperty, error) {
	if multiplier, ok := tailwindSpacingMultiplier[value]; ok {
		return []CSSProperty{{property, fmt.Sprintf("calc(var(--spacing) * %v)", multiplier)}}, nil
	}

	// Try to parse a direct number (like mb-3)
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		floatValue, _ := strconv.ParseFloat(value, 64)
		return []CSSProperty{{property, fmt.Sprintf("calc(var(--spacing) * %v)", floatValue)}}, nil
	}

	// Handle arbitrary values for spacing
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		cleanValue := value[1 : len(value)-1]
		return []CSSProperty{{property, cleanValue}}, nil
	}

	return []CSSProperty{}, fmt.Errorf("unrecognized spacing value: %s", value)
}

// handleSizeProperty generates CSS for width/height properties
func handleSizeProperty(property, value string) ([]CSSProperty, error) {
	switch value {
	case "full":
		return []CSSProperty{{property, "100%"}}, nil
	case "screen":
		if property == "height" || property == "min-height" {
			return []CSSProperty{{property, "100vh"}}, nil
		}
		return []CSSProperty{{property, "100vw"}}, nil
	case "auto":
		return []CSSProperty{{property, "auto"}}, nil
	}

	// Handle size with arbitrary values
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		cleanValue := value[1 : len(value)-1]
		return []CSSProperty{{property, cleanValue}}, nil
	}

	// Try to parse a direct number (like w-6)
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		floatValue, _ := strconv.ParseFloat(value, 64)
		return []CSSProperty{{property, fmt.Sprintf("calc(var(--spacing) * %v)", floatValue)}}, nil
	}

	return []CSSProperty{}, fmt.Errorf("unrecognized size value: %s", value)
}

// handleNegativeSpacingProperty generates CSS for negative spacing properties
func handleNegativeSpacingProperty(property, value string) ([]CSSProperty, error) {
	if multiplier, ok := tailwindSpacingMultiplier[value]; ok {
		return []CSSProperty{{property, fmt.Sprintf("calc(var(--spacing) * -%v)", multiplier)}}, nil
	}

	// Try to parse a direct number (like -top-3)
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		floatValue, _ := strconv.ParseFloat(value, 64)
		return []CSSProperty{{property, fmt.Sprintf("calc(var(--spacing) * -%v)", floatValue)}}, nil
	}

	return []CSSProperty{}, fmt.Errorf("unrecognized negative spacing value: %s", value)
}

// handleArbitraryValue parses arbitrary Tailwind values like "w-[10px]"
func handleArbitraryValue(class string) ([]CSSProperty, error) {
	// Extract the property name and value
	startBracket := strings.Index(class, "[")
	endBracket := strings.LastIndex(class, "]")

	if startBracket == -1 || endBracket == -1 || startBracket > endBracket {
		return nil, fmt.Errorf("invalid arbitrary value format: %s", class)
	}

	property := class[:startBracket]
	value := class[startBracket+1 : endBracket]

	// Replace underscores with spaces in the value (commonly used in Tailwind to represent spaces)
	processedValue := strings.ReplaceAll(value, "_", " ")

	// Check for prefixed arbitrary values like bg-[image:url(...)]
	colonIndex := strings.Index(processedValue, ":")
	if colonIndex != -1 {
		prefix := processedValue[:colonIndex]
		valueContent := processedValue[colonIndex+1:]

		// Handle special cases for prefixed arbitrary values
		switch {
		case property == "bg-" && prefix == "image":
			return []CSSProperty{{"background-image", valueContent}}, nil
		case property == "bg-" && prefix == "size":
			return []CSSProperty{{"background-size", valueContent}}, nil
		}
	}

	// Map common prefixes to CSS properties
	var cssProperty string
	switch {
	case property == "w-":
		cssProperty = "width"
	case property == "h-":
		cssProperty = "height"
	case property == "p-":
		cssProperty = "padding"
	case property == "m-":
		cssProperty = "margin"
	case property == "bg-":
		cssProperty = "background"
	case property == "text-":
		cssProperty = "color" // Could be font-size too, but typically color
	default:
		return nil, fmt.Errorf("unsupported arbitrary property: %s", property)
	}

	return []CSSProperty{{cssProperty, processedValue}}, nil
}
