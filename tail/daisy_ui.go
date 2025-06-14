package tail

import (
	"strings"
)

// daisyUIComponents contains a list of daisyUI component classes for identification
var daisyUIComponents = map[string]bool{
	"accordion":        true,
	"alert":            true,
	"avatar":           true,
	"badge":            true,
	"breadcrumbs":      true,
	"btn":              true,
	"card":             true,
	"carousel":         true,
	"chat":             true,
	"checkbox":         true,
	"collapse":         true,
	"countdown":        true,
	"diff":             true,
	"divider":          true,
	"dock":             true,
	"drawer":           true,
	"dropdown":         true,
	"fieldset":         true,
	"file-input":       true,
	"filter":           true,
	"footer":           true,
	"hero":             true,
	"indicator":        true,
	"input":            true,
	"join":             true,
	"kbd":              true,
	"label":            true,
	"link":             true,
	"list":             true,
	"loading":          true,
	"mask":             true,
	"menu":             true,
	"mockup-browser":   true,
	"mockup-code":      true,
	"mockup-phone":     true,
	"mockup-window":    true,
	"modal":            true,
	"navbar":           true,
	"progress":         true,
	"radial-progress":  true,
	"radio":            true,
	"range":            true,
	"rating":           true,
	"select":           true,
	"skeleton":         true,
	"stack":            true,
	"stat":             true,
	"status":           true,
	"steps":            true,
	"swap":             true,
	"tab":              true,
	"table":            true,
	"textarea":         true,
	"theme-controller": true,
	"timeline":         true,
	"toast":            true,
	"toggle":           true,
	"validator":        true,
}

// daisyUIPartClasses maps component parts to their base components
var daisyUIPartClasses = map[string]string{
	// Just a few examples
	"collapse-title":   "collapse",
	"collapse-content": "collapse",
	"card-title":       "card",
	"card-body":        "card",
	"card-actions":     "card",
	"badge-outline":    "badge",
	"badge-primary":    "badge",
	"btn-primary":      "btn",
	"btn-success":      "btn",
	"btn-error":        "btn",
	"btn-outline":      "btn",
	"menu-title":       "menu",
	"drawer-toggle":    "drawer",
	"drawer-content":   "drawer",
	"drawer-side":      "drawer",
	"tab-active":       "tab",
}

// isDaisyUIClass checks if a class is a daisyUI component
func isDaisyUIClass(class string) bool {
	// Check if it's a main component
	if daisyUIComponents[class] {
		return true
	}

	// Check if it's a part component
	for partClass := range daisyUIPartClasses {
		if strings.HasPrefix(class, partClass) || class == partClass {
			return true
		}
	}

	return false
}

// getDaisyUIBaseComponent returns the base component for a daisyUI class
func getDaisyUIBaseComponent(class string) string {
	// If it's a main component, return itself
	if daisyUIComponents[class] {
		return class
	}

	// Check if it's a part component
	for partClass, baseComponent := range daisyUIPartClasses {
		if strings.HasPrefix(class, partClass) || class == partClass {
			return baseComponent
		}
	}

	// Find by prefix (e.g., btn-primary -> btn)
	for component := range daisyUIComponents {
		if strings.HasPrefix(class, component+"-") {
			return component
		}
	}

	return ""
}

// categorizeDaisyUIClasses groups daisyUI classes by their base components
func categorizeDaisyUIClasses(classes []string) map[string][]string {
	result := make(map[string][]string)

	for _, class := range classes {
		if isDaisyUIClass(class) {
			baseComponent := getDaisyUIBaseComponent(class)
			if baseComponent != "" {
				result[baseComponent] = append(result[baseComponent], class)
			}
		}
	}

	return result
}

// generateDaisyUICSS generates CSS for daisyUI components
func generateDaisyUICSS(daisyUIClasses map[string][]string) string {
	// In a real implementation, this would generate the CSS for daisyUI components
	// For this example, we'll just return a placeholder
	return "/* daisyUI components would be generated here */"
}
