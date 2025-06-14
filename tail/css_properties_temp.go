package tail

// This file contains additional CSS properties that should be merged into the main css_properties.go file
// These are specifically to address the missing class mappings reported in the compiler

// Extend the getCSSPropertiesForClass function to handle the missing classes
func handleMissingCSSProperties(class string) ([]CSSProperty, bool) {
	switch {
	// Height with px value
	case class == "h-px":
		return []CSSProperty{{"height", "1px"}}, true

	// Max width large
	case class == "max-w-lg":
		return []CSSProperty{{"max-width", "32rem"}}, true

	// Background attachment
	case class == "bg-fixed":
		return []CSSProperty{{"background-attachment", "fixed"}}, true

	// Font family mono
	case class == "font-mono":
		return []CSSProperty{{"font-family", "var(--font-mono)"}}, true

	default:
		return nil, false
	}
}
