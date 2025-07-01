package core

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"wispy-core/common"
)

var daisyUIComponentMap = map[string]string{
	"alert":       "alert.css",
	"avatar":      "avatar.css",
	"badge":       "badge.css",
	"breadcrumbs": "breadcrumbs.css",
	"btn":         "button.css",
	"cally":       "calendar.css",
	"card":        "card.css",
	"carousel":    "carousel.css",
	"chat":        "chat.css",
	// Add more as you add more component CSS files
}
var daisyUIComponentDir = "wispytail/daisyui-components"

// Tailwind v4 cascade layers
const (
	LayerTheme      = "theme"
	LayerBase       = "base"
	LayerComponents = "components"
	LayerUtilities  = "utilities"
)

// Map for responsive breakpoints - Tailwind v4 compatible
var responsivePrefixes = map[string]string{
	"sm":  "@media (min-width: 640px)",
	"md":  "@media (min-width: 768px)",
	"lg":  "@media (min-width: 1024px)",
	"xl":  "@media (min-width: 1280px)",
	"2xl": "@media (min-width: 1536px)",
	"3xl": "@media (min-width: 1920px)", // New in v4
}

// Container query support - new in v4
var containerQueries = map[string]string{
	"@xs":  "@container (min-width: 20rem)",
	"@sm":  "@container (min-width: 24rem)",
	"@md":  "@container (min-width: 28rem)",
	"@lg":  "@container (min-width: 32rem)",
	"@xl":  "@container (min-width: 36rem)",
	"@2xl": "@container (min-width: 42rem)",
	"@3xl": "@container (min-width: 48rem)",
	"@4xl": "@container (min-width: 56rem)",
	"@5xl": "@container (min-width: 64rem)",
	"@6xl": "@container (min-width: 72rem)",
	"@7xl": "@container (min-width: 80rem)",
}

// For group variants, we’ll handle them specially in buildSelector.
var statePseudoPrefixes = map[string]string{
	"hover":             ":hover",
	"focus":             ":focus",
	"active":            ":active",
	"visited":           ":visited",
	"checked":           ":checked",
	"disabled":          ":disabled",
	"enabled":           ":enabled",
	"read-only":         ":read-only",
	"read-write":        ":read-write",
	"focus-within":      ":focus-within",
	"focus-visible":     ":focus-visible",
	"autofill":          ":autofill",
	"placeholder-shown": ":placeholder-shown",
	"default":           ":default",
	"first":             ":first-child",
	"last":              ":last-child",
	"only":              ":only-child",
	"odd":               ":nth-child(odd)",
	"even":              ":nth-child(even)",
	"first-of-type":     ":first-of-type",
	"last-of-type":      ":last-of-type",
	"only-of-type":      ":only-of-type",
	"empty":             ":empty",
	"open":              "[open]",
	"inert":             "[inert]", // New in v4
	"target":            ":target",
	"scope":             ":scope",
	"indeterminate":     ":indeterminate",
	"valid":             ":valid",
	"invalid":           ":invalid",
	"required":          ":required",
	"optional":          ":optional",
	"in-range":          ":in-range",
	"out-of-range":      ":out-of-range",
	"print":             "@media print",
	"dark":              "@media (prefers-color-scheme: dark)",
}

// --- Selector & Media Wrapping Helpers ---
// BuildSelector constructs the CSS selector using the full original class (with proper escaping)
// and then applies state pseudo-classes and group variants based on the provided prefixes.
// Enhanced for Tailwind v4 with container queries and new variants
func BuildSelector(originalClass string, prefixes []string) (selector string, mediaQuery string) {
	selector = "." + EscapeClass(originalClass)
	var pseudoClasses []string
	var groupSelector string

	for _, p := range prefixes {
		// Container queries (new in v4) - fix the variable usage
		if strings.HasPrefix(p, "@") {
			if cq, exists := containerQueries[p]; exists {
				mediaQuery = cq
				continue
			}
		}

		if strings.HasPrefix(p, "group-") {
			// Enhanced group-based variants
			switch p {
			case "group-hover":
				groupSelector = ".group:hover "
			case "group-focus":
				groupSelector = ".group:focus "
			case "group-active":
				groupSelector = ".group:active "
			case "group-focus-within":
				groupSelector = ".group:focus-within "
			case "group-focus-visible":
				groupSelector = ".group:focus-visible "
			case "group-aria-expanded":
				groupSelector = ".group[aria-expanded='true'] "
			case "group-aria-selected":
				groupSelector = ".group[aria-selected='true'] "
			case "group-aria-checked":
				groupSelector = ".group[aria-checked='true'] "
			case "group-data-state-open":
				groupSelector = ".group[data-state='open'] "
			case "group-data-state-closed":
				groupSelector = ".group[data-state='closed'] "
			}
		} else if strings.HasPrefix(p, "peer-") {
			// Enhanced peer-based variants
			switch p {
			case "peer-hover":
				groupSelector = ".peer:hover ~ "
			case "peer-focus":
				groupSelector = ".peer:focus ~ "
			case "peer-active":
				groupSelector = ".peer:active ~ "
			case "peer-checked":
				groupSelector = ".peer:checked ~ "
			case "peer-disabled":
				groupSelector = ".peer:disabled ~ "
			case "peer-valid":
				groupSelector = ".peer:valid ~ "
			case "peer-invalid":
				groupSelector = ".peer:invalid ~ "
			case "peer-required":
				groupSelector = ".peer:required ~ "
			case "peer-optional":
				groupSelector = ".peer:optional ~ "
			}
		} else if strings.HasPrefix(p, "aria-") {
			// Handle ARIA attributes - enhanced for v4
			attrName := strings.TrimPrefix(p, "aria-")
			if strings.Contains(attrName, "-") {
				// Handle specific ARIA values like aria-expanded-true
				parts := strings.Split(attrName, "-")
				if len(parts) >= 2 && (parts[len(parts)-1] == "true" || parts[len(parts)-1] == "false") {
					value := parts[len(parts)-1]
					attr := strings.Join(parts[:len(parts)-1], "-")
					pseudoClasses = append(pseudoClasses, fmt.Sprintf("[aria-%s='%s']", attr, value))
				} else {
					pseudoClasses = append(pseudoClasses, fmt.Sprintf("[aria-%s]", attrName))
				}
			} else {
				pseudoClasses = append(pseudoClasses, fmt.Sprintf("[aria-%s='true']", attrName))
			}
		} else if strings.HasPrefix(p, "data-") {
			// Enhanced data attributes handling for v4
			if strings.Contains(p, "=") {
				// Handle data-[attribute=value] syntax
				pseudoClasses = append(pseudoClasses, fmt.Sprintf("[%s]", p))
			} else if strings.Contains(p, "-") && len(strings.Split(p, "-")) > 2 {
				// Handle data-state-open style attributes
				parts := strings.Split(p, "-")
				if len(parts) >= 3 {
					attr := strings.Join(parts[:len(parts)-1], "-")
					value := parts[len(parts)-1]
					pseudoClasses = append(pseudoClasses, fmt.Sprintf("[%s='%s']", attr, value))
				} else {
					pseudoClasses = append(pseudoClasses, fmt.Sprintf("[%s]", p))
				}
			} else {
				pseudoClasses = append(pseudoClasses, fmt.Sprintf("[%s]", p))
			}
		} else if mq, exists := responsivePrefixes[p]; exists {
			// Handle responsive prefixes
			mediaQuery = mq
		} else if pseudo, ok := statePseudoPrefixes[p]; ok {
			// Handle special media queries like print
			if strings.HasPrefix(pseudo, "@media") {
				mediaQuery = pseudo
			} else {
				pseudoClasses = append(pseudoClasses, pseudo)
			}
		} else if strings.HasPrefix(p, "not-") {
			// Enhanced "not-" pseudo-classes for v4
			notState := strings.TrimPrefix(p, "not-")
			if pseudo, exists := statePseudoPrefixes[notState]; exists {
				if strings.HasPrefix(pseudo, "@media") {
					// Don't negate media queries
					continue
				}
				pseudoClasses = append(pseudoClasses, fmt.Sprintf(":not(%s)", pseudo))
			} else if strings.HasPrefix(notState, "data-") {
				// Handle not-data-* variants
				pseudoClasses = append(pseudoClasses, fmt.Sprintf(":not([%s])", notState))
			}
		}
	}

	if groupSelector != "" {
		selector = groupSelector + selector
	}
	if len(pseudoClasses) > 0 {
		selector += strings.Join(pseudoClasses, "")
	}
	return selector, mediaQuery
}

func generateRuleForClass(class string, trie *common.Trie) (rule string, mediaQuery string, ok bool) {
	// Handle colons inside brackets properly - they are not variant separators
	var parts []string
	var currentPart strings.Builder
	inBrackets := 0

	for _, char := range class {
		if char == '[' {
			inBrackets++
			currentPart.WriteRune(char)
		} else if char == ']' {
			inBrackets--
			currentPart.WriteRune(char)
		} else if char == ':' && inBrackets == 0 {
			// Only split on colons outside of brackets
			parts = append(parts, currentPart.String())
			currentPart.Reset()
		} else {
			currentPart.WriteRune(char)
		}
	}
	// Add the last part
	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}

	// If no parts were found, use the original class
	if len(parts) == 0 {
		parts = []string{class}
	}

	base := parts[len(parts)-1]      // get last item
	prefixes := parts[:len(parts)-1] // all except last item

	// Try a static lookup on the base utility.
	if ruleBody, ok := trie.Search(base); ok {
		selector, mediaQuery := BuildSelector(class, prefixes)
		return selector + " { " + ruleBody + " }", mediaQuery, true
	} else {
		if ruleBody, ok := ProcessRecipe(base); ok {
			selector, mediaQuery := BuildSelector(class, prefixes)
			return selector + " { " + ruleBody + " }", mediaQuery, true
		}
		// TODO: taggable from debug env flag
		// fmt.Println("[X]", base)
	}

	return "", "", false
}

var fallbackTrie = common.NewTrie()

// GenerateCSS accepts a set of class names and the trie, returning Tailwind v4 compatible CSS with cascade layers.
func ResolveClasses(classes []string, trie *common.Trie) string {
	if trie == nil {
		trie = fallbackTrie
	}

	var buffer bytes.Buffer
	var utilityRules []string
	mediaQueryRules := make(map[string][]string)
	containerQueryRules := make(map[string][]string)
	var mediaQueryList []string // Maintain insertion order
	var containerQueryList []string

	shouldDebug := common.GetEnvBool("DEBUG_WISPY_TAIL", false) // Set to true to enable debug output

	if shouldDebug {
		fmt.Println("Missing classes:")
	}
	for _, className := range classes {
		if rule, mediaQuery, ok := generateRuleForClass(className, trie); ok {
			if mediaQuery == "" {
				// Rules without media queries go into utilities layer
				utilityRules = append(utilityRules, rule)
			} else if strings.HasPrefix(mediaQuery, "@container") {
				// Container queries get their own handling
				if _, exists := containerQueryRules[mediaQuery]; !exists {
					containerQueryList = append(containerQueryList, mediaQuery)
				}
				containerQueryRules[mediaQuery] = append(containerQueryRules[mediaQuery], rule)
			} else {
				// Regular media queries
				if _, exists := mediaQueryRules[mediaQuery]; !exists {
					mediaQueryList = append(mediaQueryList, mediaQuery)
				}
				mediaQueryRules[mediaQuery] = append(mediaQueryRules[mediaQuery], rule)
			}
		} else {
			if shouldDebug {
				fmt.Printf("[X] %s\n", className)
			}
		}
	}

	// Output utilities layer with default rules
	if len(utilityRules) > 0 {
		buffer.WriteString("@layer utilities {\n")
		for _, rule := range utilityRules {
			buffer.WriteString("  " + rule + "\n")
		}
		buffer.WriteString("}\n")
	}

	// Sort media queries by Tailwind priority
	sort.SliceStable(mediaQueryList, func(i, j int) bool {
		return MediaQueryPriority(mediaQueryList[i]) < MediaQueryPriority(mediaQueryList[j])
	})

	// Output rules grouped by ordered media queries within utilities layer
	for _, mq := range mediaQueryList {
		buffer.WriteString("@layer utilities {\n")
		buffer.WriteString("  " + mq + " {\n")
		for _, rule := range mediaQueryRules[mq] {
			buffer.WriteString("    " + rule + "\n")
		}
		buffer.WriteString("  }\n")
		buffer.WriteString("}\n")
	}

	// Output container queries within utilities layer
	for _, cq := range containerQueryList {
		buffer.WriteString("@layer utilities {\n")
		buffer.WriteString("  " + cq + " {\n")
		for _, rule := range containerQueryRules[cq] {
			buffer.WriteString("    " + rule + "\n")
		}
		buffer.WriteString("  }\n")
		buffer.WriteString("}\n")
	}

	return buffer.String()
}

// Enhanced priority for Tailwind v4 media queries including container queries
func MediaQueryPriority(mq string) int {
	priority := map[string]int{
		"@media print":               0, // Print styles first
		"@media (min-width: 640px)":  1, // sm
		"@media (min-width: 768px)":  2, // md
		"@media (min-width: 1024px)": 3, // lg
		"@media (min-width: 1280px)": 4, // xl
		"@media (min-width: 1536px)": 5, // 2xl
		"@media (min-width: 1920px)": 6, // 3xl (new in v4)
	}
	if p, exists := priority[mq]; exists {
		return p
	}

	// Container queries get higher priority
	if strings.HasPrefix(mq, "@container") {
		return 50
	}

	return 99 // Default lowest priority for unknown media queries
}

// escapeClass escapes special characters (such as colon and square brackets) in class names.
func EscapeClass(class string) string {
	s := strings.ReplaceAll(class, "\\", "\\\\")
	s = strings.ReplaceAll(s, ":", "\\:")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "]", "\\]")
	s = strings.ReplaceAll(s, ".", "\\.")
	s = strings.ReplaceAll(s, "/", "\\/")
	return s
}

// GenerateFullCSSv4 combines base CSS with utility classes for complete Tailwind v4 output
func GenerateCSSFromClasses(classes []string, config *ThemeConfig, trie *common.Trie) string {
	var buffer bytes.Buffer

	// Add base CSS first
	// buffer.WriteString(generateThemeLayer(config))
	// buffer.WriteString(generateCssBaseLayer())

	// --- daisyUI component CSS auto-inclusion ---
	// Track which component CSS files have been included
	included := make(map[string]bool)
	for _, className := range classes {
		// Only check the root/base part (before any colon/variant)
		base := className
		if idx := strings.Index(base, ":"); idx > 0 {
			base = base[:idx]
		}
		if cssFile, ok := daisyUIComponentMap[base]; ok && !included[cssFile] {
			cssPath := filepath.Join(daisyUIComponentDir, cssFile)
			if cssContent, err := os.ReadFile(cssPath); err == nil {
				buffer.WriteString("\n/* daisyUI component: " + base + " */\n")
				buffer.Write(cssContent)
				buffer.WriteString("\n")
				included[cssFile] = true
			}
		}
	}

	// Add utility classes
	buffer.WriteString(ResolveClasses(classes, trie))

	return buffer.String()
}

// // PopulateBasicUtilities adds common Tailwind utilities to the trie
// func PopulateBasicUtilities(trie *common.Trie) {
// 	// Container utility
// 	trie.Insert("container", "width: 100%; margin-left: auto; margin-right: auto; max-width: 1536px")

// 	// Display utilities
// 	trie.Insert("block", "display: block")
// 	trie.Insert("inline", "display: inline")
// 	trie.Insert("inline-block", "display: inline-block")
// 	trie.Insert("flex", "display: flex")
// 	trie.Insert("inline-flex", "display: inline-flex")
// 	trie.Insert("grid", "display: grid")
// 	trie.Insert("inline-grid", "display: inline-grid")
// 	trie.Insert("contents", "display: contents")
// 	trie.Insert("hidden", "display: none")
// 	trie.Insert("flow-root", "display: flow-root")

// 	// Position utilities
// 	trie.Insert("static", "position: static")
// 	trie.Insert("fixed", "position: fixed")
// 	trie.Insert("absolute", "position: absolute")
// 	trie.Insert("relative", "position: relative")
// 	trie.Insert("sticky", "position: sticky")

// 	// Negative positioning utilities
// 	trie.Insert("-top-px", "top: -1px")
// 	trie.Insert("-bottom-px", "bottom: -1px")
// 	trie.Insert("-left-px", "left: -1px")
// 	trie.Insert("-right-px", "right: -1px")

// 	// Flex utilities
// 	trie.Insert("flex-row", "flex-direction: row")
// 	trie.Insert("flex-col", "flex-direction: column")
// 	trie.Insert("flex-wrap", "flex-wrap: wrap")
// 	trie.Insert("flex-nowrap", "flex-wrap: nowrap")
// 	trie.Insert("items-start", "align-items: flex-start")
// 	trie.Insert("items-center", "align-items: center")
// 	trie.Insert("items-end", "align-items: flex-end")
// 	trie.Insert("items-stretch", "align-items: stretch")
// 	trie.Insert("justify-start", "justify-content: flex-start")
// 	trie.Insert("justify-center", "justify-content: center")
// 	trie.Insert("justify-end", "justify-content: flex-end")
// 	trie.Insert("justify-between", "justify-content: space-between")
// 	trie.Insert("justify-around", "justify-content: space-around")
// 	trie.Insert("justify-evenly", "justify-content: space-evenly")

// 	// Text utilities
// 	trie.Insert("text-left", "text-align: left")
// 	trie.Insert("text-center", "text-align: center")
// 	trie.Insert("text-right", "text-align: right")
// 	trie.Insert("text-justify", "text-align: justify")
// 	trie.Insert("text-start", "text-align: start")
// 	trie.Insert("text-end", "text-align: end")

// 	// Font utilities
// 	trie.Insert("font-sans", "font-family: var(--font-sans)")
// 	trie.Insert("font-serif", "font-family: var(--font-serif)")
// 	trie.Insert("font-mono", "font-family: var(--font-mono)")
// 	trie.Insert("font-thin", "font-weight: var(--font-weight-thin)")
// 	trie.Insert("font-light", "font-weight: var(--font-weight-light)")
// 	trie.Insert("font-normal", "font-weight: var(--font-weight-normal)")
// 	trie.Insert("font-medium", "font-weight: var(--font-weight-medium)")
// 	trie.Insert("font-semibold", "font-weight: var(--font-weight-semibold)")
// 	trie.Insert("font-bold", "font-weight: var(--font-weight-bold)")
// 	trie.Insert("font-extrabold", "font-weight: var(--font-weight-extrabold)")
// 	trie.Insert("font-black", "font-weight: var(--font-weight-black)")

// 	// Text size utilities using CSS variables
// 	trie.Insert("text-xs", "font-size: var(--text-xs)")
// 	trie.Insert("text-sm", "font-size: var(--text-sm)")
// 	trie.Insert("text-base", "font-size: var(--text-base)")
// 	trie.Insert("text-lg", "font-size: var(--text-lg)")
// 	trie.Insert("text-xl", "font-size: var(--text-xl)")
// 	trie.Insert("text-2xl", "font-size: var(--text-2xl)")
// 	trie.Insert("text-3xl", "font-size: var(--text-3xl)")
// 	trie.Insert("text-4xl", "font-size: var(--text-4xl)")
// 	trie.Insert("text-5xl", "font-size: var(--text-5xl)")
// 	trie.Insert("text-6xl", "font-size: var(--text-6xl)")
// 	trie.Insert("text-7xl", "font-size: var(--text-7xl)")
// 	trie.Insert("text-8xl", "font-size: var(--text-8xl)")
// 	trie.Insert("text-9xl", "font-size: var(--text-9xl)")

// 	// Color utilities using CSS variables
// 	trie.Insert("text-primary", "color: var(--color-primary)")
// 	trie.Insert("text-secondary", "color: var(--color-secondary)")
// 	trie.Insert("text-accent", "color: var(--color-accent)")
// 	trie.Insert("text-neutral", "color: var(--color-neutral)")
// 	trie.Insert("text-base-content", "color: var(--color-base-content)")
// 	trie.Insert("text-info", "color: var(--color-info)")
// 	trie.Insert("text-success", "color: var(--color-success)")
// 	trie.Insert("text-warning", "color: var(--color-warning)")
// 	trie.Insert("text-error", "color: var(--color-error)")

// 	trie.Insert("bg-primary", "background-color: var(--color-primary)")
// 	trie.Insert("bg-secondary", "background-color: var(--color-secondary)")
// 	trie.Insert("bg-accent", "background-color: var(--color-accent)")
// 	trie.Insert("bg-neutral", "background-color: var(--color-neutral)")
// 	trie.Insert("bg-base-100", "background-color: var(--color-base-100)")
// 	trie.Insert("bg-base-200", "background-color: var(--color-base-200)")
// 	trie.Insert("bg-base-300", "background-color: var(--color-base-300)")
// 	trie.Insert("bg-info", "background-color: var(--color-info)")
// 	trie.Insert("bg-success", "background-color: var(--color-success)")
// 	trie.Insert("bg-warning", "background-color: var(--color-warning)")
// 	trie.Insert("bg-error", "background-color: var(--color-error)")

// 	// Border utilities
// 	trie.Insert("border", "border-width: 1px")
// 	trie.Insert("border-0", "border-width: 0")
// 	trie.Insert("border-2", "border-width: 2px")
// 	trie.Insert("border-4", "border-width: 4px")
// 	trie.Insert("border-8", "border-width: 8px")
// 	trie.Insert("border-solid", "border-style: solid")
// 	trie.Insert("border-dashed", "border-style: dashed")
// 	trie.Insert("border-dotted", "border-style: dotted")
// 	trie.Insert("border-double", "border-style: double")
// 	trie.Insert("border-none", "border-style: none")

// 	// Rounded utilities using CSS variables
// 	trie.Insert("rounded", "border-radius: var(--radius-md)")
// 	trie.Insert("rounded-none", "border-radius: var(--radius-none)")
// 	trie.Insert("rounded-sm", "border-radius: var(--radius-sm)")
// 	trie.Insert("rounded-md", "border-radius: var(--radius-md)")
// 	trie.Insert("rounded-lg", "border-radius: var(--radius-lg)")
// 	trie.Insert("rounded-xl", "border-radius: var(--radius-xl)")
// 	trie.Insert("rounded-2xl", "border-radius: var(--radius-2xl)")
// 	trie.Insert("rounded-3xl", "border-radius: var(--radius-3xl)")
// 	trie.Insert("rounded-full", "border-radius: var(--radius-full)")

// 	// Opacity utilities
// 	trie.Insert("opacity-0", "opacity: 0")
// 	trie.Insert("opacity-25", "opacity: 0.25")
// 	trie.Insert("opacity-50", "opacity: 0.5")
// 	trie.Insert("opacity-75", "opacity: 0.75")
// 	trie.Insert("opacity-100", "opacity: 1")

// 	// Transition utilities (v4 enhanced)
// 	trie.Insert("transition", "transition-property: color, background-color, border-color, text-decoration-color, fill, stroke, opacity, box-shadow, transform, filter, backdrop-filter")
// 	trie.Insert("transition-none", "transition-property: none")
// 	trie.Insert("transition-all", "transition-property: all")
// 	trie.Insert("transition-colors", "transition-property: color, background-color, border-color, text-decoration-color, fill, stroke")
// 	trie.Insert("transition-opacity", "transition-property: opacity")
// 	trie.Insert("transition-shadow", "transition-property: box-shadow")
// 	trie.Insert("transition-transform", "transition-property: transform")

// 	// Transform utilities (enhanced for v4)
// 	trie.Insert("transform", "transform: var(--tw-transform)")
// 	trie.Insert("transform-gpu", "transform: var(--tw-transform)")
// 	trie.Insert("transform-none", "transform: none")

// 	// Scale utilities
// 	trie.Insert("scale-0", "transform: scale(0)")
// 	trie.Insert("scale-50", "transform: scale(0.5)")
// 	trie.Insert("scale-75", "transform: scale(0.75)")
// 	trie.Insert("scale-90", "transform: scale(0.9)")
// 	trie.Insert("scale-95", "transform: scale(0.95)")
// 	trie.Insert("scale-100", "transform: scale(1)")
// 	trie.Insert("scale-105", "transform: scale(1.05)")
// 	trie.Insert("scale-110", "transform: scale(1.1)")
// 	trie.Insert("scale-125", "transform: scale(1.25)")
// 	trie.Insert("scale-150", "transform: scale(1.5)")

// 	// Cursor utilities
// 	trie.Insert("cursor-auto", "cursor: auto")
// 	trie.Insert("cursor-default", "cursor: default")
// 	trie.Insert("cursor-pointer", "cursor: pointer")
// 	trie.Insert("cursor-wait", "cursor: wait")
// 	trie.Insert("cursor-text", "cursor: text")
// 	trie.Insert("cursor-move", "cursor: move")
// 	trie.Insert("cursor-help", "cursor: help")
// 	trie.Insert("cursor-not-allowed", "cursor: not-allowed")

// 	// Overflow utilities
// 	trie.Insert("overflow-auto", "overflow: auto")
// 	trie.Insert("overflow-hidden", "overflow: hidden")
// 	trie.Insert("overflow-clip", "overflow: clip")
// 	trie.Insert("overflow-visible", "overflow: visible")
// 	trie.Insert("overflow-scroll", "overflow: scroll")
// 	trie.Insert("overflow-x-auto", "overflow-x: auto")
// 	trie.Insert("overflow-y-auto", "overflow-y: auto")
// 	trie.Insert("overflow-x-hidden", "overflow-x: hidden")
// 	trie.Insert("overflow-y-hidden", "overflow-y: hidden")
// 	trie.Insert("overflow-x-clip", "overflow-x: clip")
// 	trie.Insert("overflow-y-clip", "overflow-y: clip")
// 	trie.Insert("overflow-x-visible", "overflow-x: visible")
// 	trie.Insert("overflow-y-visible", "overflow-y: visible")
// 	trie.Insert("overflow-x-scroll", "overflow-x: scroll")
// 	trie.Insert("overflow-y-scroll", "overflow-y: scroll")

// 	// Text decoration utilities
// 	trie.Insert("decoration-0", "text-decoration-thickness: 0px")
// 	trie.Insert("decoration-1", "text-decoration-thickness: 1px")
// 	trie.Insert("decoration-2", "text-decoration-thickness: 2px")
// 	trie.Insert("decoration-4", "text-decoration-thickness: 4px")
// 	trie.Insert("decoration-8", "text-decoration-thickness: 8px")
// 	trie.Insert("decoration-auto", "text-decoration-thickness: auto")
// 	trie.Insert("decoration-from-font", "text-decoration-thickness: from-font")

// 	// Text decoration colors for sky palette
// 	trie.Insert("decoration-sky-50", "text-decoration-color: var(--color-sky-50)")
// 	trie.Insert("decoration-sky-100", "text-decoration-color: var(--color-sky-100)")
// 	trie.Insert("decoration-sky-200", "text-decoration-color: var(--color-sky-200)")
// 	trie.Insert("decoration-sky-300", "text-decoration-color: var(--color-sky-300)")
// 	trie.Insert("decoration-sky-400", "text-decoration-color: var(--color-sky-400)")
// 	trie.Insert("decoration-sky-500", "text-decoration-color: var(--color-sky-500)")
// 	trie.Insert("decoration-sky-600", "text-decoration-color: var(--color-sky-600)")
// 	trie.Insert("decoration-sky-700", "text-decoration-color: var(--color-sky-700)")
// 	trie.Insert("decoration-sky-800", "text-decoration-color: var(--color-sky-800)")
// 	trie.Insert("decoration-sky-900", "text-decoration-color: var(--color-sky-900)")
// 	trie.Insert("decoration-sky-950", "text-decoration-color: var(--color-sky-950)")

// 	// Stroke utilities for sky palette
// 	trie.Insert("stroke-sky-50", "stroke: var(--color-sky-50)")
// 	trie.Insert("stroke-sky-100", "stroke: var(--color-sky-100)")
// 	trie.Insert("stroke-sky-200", "stroke: var(--color-sky-200)")
// 	trie.Insert("stroke-sky-300", "stroke: var(--color-sky-300)")
// 	trie.Insert("stroke-sky-400", "stroke: var(--color-sky-400)")
// 	trie.Insert("stroke-sky-500", "stroke: var(--color-sky-500)")
// 	trie.Insert("stroke-sky-600", "stroke: var(--color-sky-600)")
// 	trie.Insert("stroke-sky-700", "stroke: var(--color-sky-700)")
// 	trie.Insert("stroke-sky-800", "stroke: var(--color-sky-800)")
// 	trie.Insert("stroke-sky-900", "stroke: var(--color-sky-900)")
// 	trie.Insert("stroke-sky-950", "stroke: var(--color-sky-950)")

// 	// Fill utilities for sky palette
// 	trie.Insert("fill-sky-50", "fill: var(--color-sky-50)")
// 	trie.Insert("fill-sky-100", "fill: var(--color-sky-100)")
// 	trie.Insert("fill-sky-200", "fill: var(--color-sky-200)")
// 	trie.Insert("fill-sky-300", "fill: var(--color-sky-300)")
// 	trie.Insert("fill-sky-400", "fill: var(--color-sky-400)")
// 	trie.Insert("fill-sky-500", "fill: var(--color-sky-500)")
// 	trie.Insert("fill-sky-600", "fill: var(--color-sky-600)")
// 	trie.Insert("fill-sky-700", "fill: var(--color-sky-700)")
// 	trie.Insert("fill-sky-800", "fill: var(--color-sky-800)")
// 	trie.Insert("fill-sky-900", "fill: var(--color-sky-900)")
// 	trie.Insert("fill-sky-950", "fill: var(--color-sky-950)")

// 	// Space utilities
// 	trie.Insert("space-y-0", "> :not([hidden]) ~ :not([hidden]) { margin-top: 0; margin-bottom: 0; }")
// 	trie.Insert("space-y-1", "> :not([hidden]) ~ :not([hidden]) { margin-top: calc(var(--spacing) * 1); }")
// 	trie.Insert("space-y-2", "> :not([hidden]) ~ :not([hidden]) { margin-top: calc(var(--spacing) * 2); }")
// 	trie.Insert("space-y-3", "> :not([hidden]) ~ :not([hidden]) { margin-top: calc(var(--spacing) * 3); }")
// 	trie.Insert("space-y-4", "> :not([hidden]) ~ :not([hidden]) { margin-top: calc(var(--spacing) * 4); }")
// 	trie.Insert("space-y-5", "> :not([hidden]) ~ :not([hidden]) { margin-top: calc(var(--spacing) * 5); }")
// 	trie.Insert("space-y-6", "> :not([hidden]) ~ :not([hidden]) { margin-top: calc(var(--spacing) * 6); }")
// 	trie.Insert("space-y-8", "> :not([hidden]) ~ :not([hidden]) { margin-top: calc(var(--spacing) * 8); }")
// 	trie.Insert("space-y-10", "> :not([hidden]) ~ :not([hidden]) { margin-top: calc(var(--spacing) * 10); }")
// 	trie.Insert("space-y-12", "> :not([hidden]) ~ :not([hidden]) { margin-top: calc(var(--spacing) * 12); }")

// 	trie.Insert("space-x-0", "> :not([hidden]) ~ :not([hidden]) { margin-left: 0; margin-right: 0; }")
// 	trie.Insert("space-x-1", "> :not([hidden]) ~ :not([hidden]) { margin-left: calc(var(--spacing) * 1); }")
// 	trie.Insert("space-x-2", "> :not([hidden]) ~ :not([hidden]) { margin-left: calc(var(--spacing) * 2); }")
// 	trie.Insert("space-x-3", "> :not([hidden]) ~ :not([hidden]) { margin-left: calc(var(--spacing) * 3); }")
// 	trie.Insert("space-x-4", "> :not([hidden]) ~ :not([hidden]) { margin-left: calc(var(--spacing) * 4); }")
// 	trie.Insert("space-x-6", "> :not([hidden]) ~ :not([hidden]) { margin-left: calc(var(--spacing) * 6); }")

// 	// Max width utilities
// 	trie.Insert("max-w-none", "max-width: none")
// 	trie.Insert("max-w-xs", "max-width: 20rem")
// 	trie.Insert("max-w-sm", "max-width: 24rem")
// 	trie.Insert("max-w-md", "max-width: 28rem")
// 	trie.Insert("max-w-lg", "max-width: 32rem")
// 	trie.Insert("max-w-xl", "max-width: 36rem")
// 	trie.Insert("max-w-2xl", "max-width: 42rem")
// 	trie.Insert("max-w-3xl", "max-width: 48rem")
// 	trie.Insert("max-w-4xl", "max-width: 56rem")
// 	trie.Insert("max-w-5xl", "max-width: 64rem")
// 	trie.Insert("max-w-6xl", "max-width: 72rem")
// 	trie.Insert("max-w-7xl", "max-width: 80rem")
// 	trie.Insert("max-w-full", "max-width: 100%")
// 	trie.Insert("max-w-min", "max-width: min-content")
// 	trie.Insert("max-w-max", "max-width: max-content")
// 	trie.Insert("max-w-fit", "max-width: fit-content")
// 	trie.Insert("max-w-prose", "max-width: 65ch")
// 	trie.Insert("max-w-screen-sm", "max-width: 640px")
// 	trie.Insert("max-w-screen-md", "max-width: 768px")
// 	trie.Insert("max-w-screen-lg", "max-width: 1024px")
// 	trie.Insert("max-w-screen-xl", "max-width: 1280px")
// 	trie.Insert("max-w-screen-2xl", "max-width: 1536px")

// 	// Grid utilities for complex layouts
// 	trie.Insert("grid-cols-1", "grid-template-columns: repeat(1, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-2", "grid-template-columns: repeat(2, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-3", "grid-template-columns: repeat(3, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-4", "grid-template-columns: repeat(4, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-5", "grid-template-columns: repeat(5, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-6", "grid-template-columns: repeat(6, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-7", "grid-template-columns: repeat(7, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-8", "grid-template-columns: repeat(8, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-9", "grid-template-columns: repeat(9, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-10", "grid-template-columns: repeat(10, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-11", "grid-template-columns: repeat(11, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-12", "grid-template-columns: repeat(12, minmax(0, 1fr))")
// 	trie.Insert("grid-cols-none", "grid-template-columns: none")
// 	trie.Insert("grid-cols-subgrid", "grid-template-columns: subgrid")

// 	// Grid column span utilities
// 	trie.Insert("col-span-full", "grid-column: 1 / -1")
// 	trie.Insert("col-span-1", "grid-column: span 1 / span 1")
// 	trie.Insert("col-span-2", "grid-column: span 2 / span 2")
// 	trie.Insert("col-span-3", "grid-column: span 3 / span 3")
// 	trie.Insert("col-span-4", "grid-column: span 4 / span 4")
// 	trie.Insert("col-span-5", "grid-column: span 5 / span 5")
// 	trie.Insert("col-span-6", "grid-column: span 6 / span 6")
// 	trie.Insert("col-span-7", "grid-column: span 7 / span 7")
// 	trie.Insert("col-span-8", "grid-column: span 8 / span 8")
// 	trie.Insert("col-span-9", "grid-column: span 9 / span 9")
// 	trie.Insert("col-span-10", "grid-column: span 10 / span 10")
// 	trie.Insert("col-span-11", "grid-column: span 11 / span 11")
// 	trie.Insert("col-span-12", "grid-column: span 12 / span 12")

// 	trie.Insert("grid-rows-1", "grid-template-rows: repeat(1, minmax(0, 1fr))")
// 	trie.Insert("grid-rows-2", "grid-template-rows: repeat(2, minmax(0, 1fr))")
// 	trie.Insert("grid-rows-3", "grid-template-rows: repeat(3, minmax(0, 1fr))")
// 	trie.Insert("grid-rows-4", "grid-template-rows: repeat(4, minmax(0, 1fr))")
// 	trie.Insert("grid-rows-5", "grid-template-rows: repeat(5, minmax(0, 1fr))")
// 	trie.Insert("grid-rows-6", "grid-template-rows: repeat(6, minmax(0, 1fr))")
// 	trie.Insert("grid-rows-none", "grid-template-rows: none")
// 	trie.Insert("grid-rows-subgrid", "grid-template-rows: subgrid")

// 	// Height utilities including line height based
// 	trie.Insert("h-0", "height: 0px")
// 	trie.Insert("h-px", "height: 1px")
// 	trie.Insert("h-0.5", "height: calc(var(--spacing) * 0.5)")
// 	trie.Insert("h-1", "height: calc(var(--spacing) * 1)")
// 	trie.Insert("h-1.5", "height: calc(var(--spacing) * 1.5)")
// 	trie.Insert("h-2", "height: calc(var(--spacing) * 2)")
// 	trie.Insert("h-2.5", "height: calc(var(--spacing) * 2.5)")
// 	trie.Insert("h-3", "height: calc(var(--spacing) * 3)")
// 	trie.Insert("h-3.5", "height: calc(var(--spacing) * 3.5)")
// 	trie.Insert("h-4", "height: calc(var(--spacing) * 4)")
// 	trie.Insert("h-5", "height: calc(var(--spacing) * 5)")
// 	trie.Insert("h-6", "height: calc(var(--spacing) * 6)")
// 	trie.Insert("h-7", "height: calc(var(--spacing) * 7)")
// 	trie.Insert("h-8", "height: calc(var(--spacing) * 8)")
// 	trie.Insert("h-9", "height: calc(var(--spacing) * 9)")
// 	trie.Insert("h-10", "height: calc(var(--spacing) * 10)")
// 	trie.Insert("h-11", "height: calc(var(--spacing) * 11)")
// 	trie.Insert("h-12", "height: calc(var(--spacing) * 12)")
// 	trie.Insert("h-14", "height: calc(var(--spacing) * 14)")
// 	trie.Insert("h-16", "height: calc(var(--spacing) * 16)")
// 	trie.Insert("h-20", "height: calc(var(--spacing) * 20)")
// 	trie.Insert("h-24", "height: calc(var(--spacing) * 24)")
// 	trie.Insert("h-28", "height: calc(var(--spacing) * 28)")
// 	trie.Insert("h-32", "height: calc(var(--spacing) * 32)")
// 	trie.Insert("h-36", "height: calc(var(--spacing) * 36)")
// 	trie.Insert("h-40", "height: calc(var(--spacing) * 40)")
// 	trie.Insert("h-44", "height: calc(var(--spacing) * 44)")
// 	trie.Insert("h-48", "height: calc(var(--spacing) * 48)")
// 	trie.Insert("h-52", "height: calc(var(--spacing) * 52)")
// 	trie.Insert("h-56", "height: calc(var(--spacing) * 56)")
// 	trie.Insert("h-60", "height: calc(var(--spacing) * 60)")
// 	trie.Insert("h-64", "height: calc(var(--spacing) * 64)")
// 	trie.Insert("h-72", "height: calc(var(--spacing) * 72)")
// 	trie.Insert("h-80", "height: calc(var(--spacing) * 80)")
// 	trie.Insert("h-96", "height: calc(var(--spacing) * 96)")
// 	trie.Insert("h-auto", "height: auto")
// 	trie.Insert("h-1/2", "height: 50%")
// 	trie.Insert("h-1/3", "height: 33.333333%")
// 	trie.Insert("h-2/3", "height: 66.666667%")
// 	trie.Insert("h-1/4", "height: 25%")
// 	trie.Insert("h-2/4", "height: 50%")
// 	trie.Insert("h-3/4", "height: 75%")
// 	trie.Insert("h-1/5", "height: 20%")
// 	trie.Insert("h-2/5", "height: 40%")
// 	trie.Insert("h-3/5", "height: 60%")
// 	trie.Insert("h-4/5", "height: 80%")
// 	trie.Insert("h-1/6", "height: 16.666667%")
// 	trie.Insert("h-2/6", "height: 33.333333%")
// 	trie.Insert("h-3/6", "height: 50%")
// 	trie.Insert("h-4/6", "height: 66.666667%")
// 	trie.Insert("h-5/6", "height: 83.333333%")
// 	trie.Insert("h-full", "height: 100%")
// 	trie.Insert("h-screen", "height: 100vh")
// 	trie.Insert("h-svh", "height: 100svh")
// 	trie.Insert("h-lvh", "height: 100lvh")
// 	trie.Insert("h-dvh", "height: 100dvh")
// 	trie.Insert("h-min", "height: min-content")
// 	trie.Insert("h-max", "height: max-content")
// 	trie.Insert("h-fit", "height: fit-content")
// }
