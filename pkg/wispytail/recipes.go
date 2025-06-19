package wispytail

import (
	"fmt"
	"strings"
	"unicode"
)

type ClassRecipe struct {
	Attribute string
}

// ProcessRecipe handles dynamic class generation with Tailwind v4 enhancements
func ProcessRecipe(input string) (string, bool) {
	// --- Robust arbitrary value utility support (Tailwind v4) ---

	// Ensure input is long enough to contain a valid class
	if len(input) < 3 {
		return "", false
	}

	// 1. CSS custom property definitions with opacity like [--pattern-fg:var(--color-gray-950)]/5
	if strings.Contains(input, "/") && strings.Contains(input, "[--") && strings.Contains(input, ":") {
		parts := strings.Split(input, "/")
		if len(parts) == 2 {
			customPropPart := parts[0]
			opacity := parts[1]
			if strings.HasPrefix(customPropPart, "[") && strings.HasSuffix(customPropPart, "]") {
				propDef := customPropPart[1 : len(customPropPart)-1]
				propParts := strings.SplitN(propDef, ":", 2)
				if len(propParts) == 2 {
					propName := propParts[0]
					propValue := propParts[1]
					if strings.Contains(propValue, "var(--color-") {
						return fmt.Sprintf("%s: color-mix(in oklab, %s %s%%, transparent)", propName, propValue, opacity), true
					}
					return fmt.Sprintf("%s: %s", propName, propValue), true
				}
			}
		}
	}

	// 2. Color/opacity patterns (e.g., bg-primary/50, text-red-500/75)
	if strings.Contains(input, "/") && !strings.HasPrefix(input, "w-") && !strings.HasPrefix(input, "h-") && !strings.Contains(input, "[--") {
		parts := strings.Split(input, "/")
		if len(parts) == 2 {
			baseClass := parts[0]
			opacity := parts[1]
			if strings.HasPrefix(baseClass, "text-") {
				colorName := strings.TrimPrefix(baseClass, "text-")
				return fmt.Sprintf("color: color-mix(in oklab, var(--color-%s) %s%%, transparent)", colorName, opacity), true
			} else if strings.HasPrefix(baseClass, "bg-") {
				colorName := strings.TrimPrefix(baseClass, "bg-")
				return fmt.Sprintf("background-color: color-mix(in oklab, var(--color-%s) %s%%, transparent)", colorName, opacity), true
			} else if strings.HasPrefix(baseClass, "border-") {
				colorName := strings.TrimPrefix(baseClass, "border-")
				return fmt.Sprintf("border-color: color-mix(in oklab, var(--color-%s) %s%%, transparent)", colorName, opacity), true
			} else if strings.HasPrefix(baseClass, "stroke-") {
				colorName := strings.TrimPrefix(baseClass, "stroke-")
				return fmt.Sprintf("stroke: color-mix(in oklab, var(--color-%s) %s%%, transparent)", colorName, opacity), true
			} else if strings.HasPrefix(baseClass, "fill-") {
				colorName := strings.TrimPrefix(baseClass, "fill-")
				return fmt.Sprintf("fill: color-mix(in oklab, var(--color-%s) %s%%, transparent)", colorName, opacity), true
			}
		}
	}

	// 3. Arbitrary value utilities: e.g. min-h-[80vh], max-w-[50vw], text-[2rem], rounded-[12px], etc.
	// Match: <prefix>-['[']value[']'] (e.g. min-h-[80vh])
	if idx := strings.Index(input, "-["); idx > 0 && strings.HasSuffix(input, "]") {
		prefix := input[:idx+1] // include the dash
		value := input[idx+2 : len(input)-1]

		// Map Tailwind utility prefixes to CSS properties
		switch prefix {
		case "min-h-[":
			return fmt.Sprintf("min-height: %s", value), true
		case "max-h-[":
			return fmt.Sprintf("max-height: %s", value), true
		case "h-[":
			return fmt.Sprintf("height: %s", value), true
		case "min-w-[":
			return fmt.Sprintf("min-width: %s", value), true
		case "max-w-[":
			return fmt.Sprintf("max-width: %s", value), true
		case "w-[":
			return fmt.Sprintf("width: %s", value), true
		case "text-[":
			return fmt.Sprintf("font-size: %s", value), true
		case "rounded-[":
			return fmt.Sprintf("border-radius: %s", value), true
		case "p-[":
			return fmt.Sprintf("padding: %s", value), true
		case "m-[":
			return fmt.Sprintf("margin: %s", value), true
		case "gap-[":
			return fmt.Sprintf("gap: %s", value), true
		case "leading-[":
			return fmt.Sprintf("line-height: %s", value), true
		case "tracking-[":
			return fmt.Sprintf("letter-spacing: %s", value), true
		case "z-[":
			return fmt.Sprintf("z-index: %s", value), true
		case "top-[":
			return fmt.Sprintf("top: %s", value), true
		case "bottom-[":
			return fmt.Sprintf("bottom: %s", value), true
		case "left-[":
			return fmt.Sprintf("left: %s", value), true
		case "right-[":
			return fmt.Sprintf("right: %s", value), true
		case "inset-[":
			return fmt.Sprintf("inset: %s", value), true
		case "border-[":
			return fmt.Sprintf("border-width: %s", value), true
		case "border-t-[":
			return fmt.Sprintf("border-top-width: %s", value), true
		case "border-b-[":
			return fmt.Sprintf("border-bottom-width: %s", value), true
		case "border-l-[":
			return fmt.Sprintf("border-left-width: %s", value), true
		case "border-r-[":
			return fmt.Sprintf("border-right-width: %s", value), true
		case "opacity-[":
			return fmt.Sprintf("opacity: %s", value), true
		case "shadow-[":
			return fmt.Sprintf("box-shadow: %s", value), true
		case "duration-[":
			return fmt.Sprintf("animation-duration: %s", value), true
		case "delay-[":
			return fmt.Sprintf("animation-delay: %s", value), true
		case "ease-[":
			return fmt.Sprintf("animation-timing-function: %s", value), true
		case "content-[":
			return fmt.Sprintf("content: %s", value), true
		// Add more as needed for other arbitrary value utilities
		default:
			// Fallback: try to parse as a custom property
			return fmt.Sprintf("%s: %s", strings.TrimSuffix(prefix, "-["), value), true
		}
	}

	// 4. Handle arbitrary values with parentheses (value) for legacy/rare cases
	if idx := strings.Index(input, "-("); idx > 0 && strings.HasSuffix(input, ")") {
		prefix := input[:idx+1]
		value := input[idx+2 : len(input)-1]
		return fmt.Sprintf("%s: %s", strings.TrimSuffix(prefix, "-("), value), true
	}

	// 5. Fallback to original logic for bracket/parenthesis or standard recipes
	if (strings.Contains(input, "[") && strings.Contains(input, "]")) || (strings.Contains(input, "(") && strings.Contains(input, ")")) {
		var startBracket, endBracket int
		if strings.Contains(input, "[") && strings.Contains(input, "]") {
			startBracket = strings.Index(input, "[")
			endBracket = strings.Index(input, "]")
		} else {
			startBracket = strings.Index(input, "(")
			endBracket = strings.Index(input, ")")
		}
		if startBracket != -1 && endBracket != -1 && endBracket > startBracket {
			prefix := input[:startBracket]
			value := input[startBracket+1 : endBracket]
			switch prefix {
			case "h-":
				return fmt.Sprintf("height: %s", value), true
			case "w-":
				return fmt.Sprintf("width: %s", value), true
			case "grid-cols-":
				gridValue := strings.ReplaceAll(value, "_", " ")
				return fmt.Sprintf("grid-template-columns: %s", gridValue), true
			case "grid-rows-":
				gridValue := strings.ReplaceAll(value, "_", " ")
				return fmt.Sprintf("grid-template-rows: %s", gridValue), true
			case "border-":
				if strings.HasPrefix(value, "--") {
					return fmt.Sprintf("border-color: var(%s)", value), true
				}
				return fmt.Sprintf("border-color: %s", value), true
			case "border-x-":
				if strings.HasPrefix(value, "--") {
					return fmt.Sprintf("border-left-color: var(%s); border-right-color: var(%s)", value, value), true
				}
				return fmt.Sprintf("border-left-color: %s; border-right-color: %s", value, value), true
			case "bg-":
				if strings.HasPrefix(value, "--") {
					return fmt.Sprintf("background-color: var(%s)", value), true
				}
				if strings.HasPrefix(value, "size:") {
					sizeValue := strings.TrimPrefix(value, "size:")
					sizeValue = strings.ReplaceAll(sizeValue, "_", " ")
					return fmt.Sprintf("background-size: %s", sizeValue), true
				}
				if strings.HasPrefix(value, "image:") {
					imageValue := strings.TrimPrefix(value, "image:")
					imageValue = strings.ReplaceAll(imageValue, "_", " ")
					return fmt.Sprintf("background-image: %s", imageValue), true
				}
				return fmt.Sprintf("background-color: %s", value), true
			}
		}
	}

	// Handle direct color classes without opacity
	if strings.HasPrefix(input, "text-") && !strings.Contains(input, "-") {
		colorName := strings.TrimPrefix(input, "text-")
		return fmt.Sprintf("color: var(--color-%s)", colorName), true
	} else if strings.HasPrefix(input, "bg-") && !strings.Contains(input, "-") {
		colorName := strings.TrimPrefix(input, "bg-")
		return fmt.Sprintf("background-color: var(--color-%s)", colorName), true
	} else if strings.HasPrefix(input, "border-") && len(strings.Split(input, "-")) == 2 {
		colorName := strings.TrimPrefix(input, "border-")
		return fmt.Sprintf("border-color: var(--color-%s)", colorName), true
	}

	lastChar := input[len(input)-1]

	// Check if number - enhanced for v4 arbitrary values
	if unicode.IsDigit(rune(lastChar)) {
		j := strings.LastIndexByte(input, byte('-'))
		if j == -1 {
			return "", false
		}

		start := input[:j+1]
		value := input[j+1:]

		// Standard recipes
		if rc, exists := StdRecipes[start]; exists {
			return fmt.Sprintf(rc.Attribute, value), true
		} else {
			return "", false
		}

		// Handle rounded bracket recipes
	} else if rune(lastChar) == ')' {
		j := strings.LastIndexByte(input, byte('-'))
		prefix := input[:j+1]
		value := input[j+1:]
		exists := false
		var rc ClassRecipe

		// Standard recipes
		if rc, exists = BracketRecipes[prefix]; exists {
			return fmt.Sprintf(rc.Attribute, value), true
		} else {
			return "", false
		}
	}
	return "", false
}

var StdRecipes = map[string]ClassRecipe{
	// Enhanced Width and Height with logical properties (v4)
	"size-":  {Attribute: "block-size: calc(var(--spacing) * %[1]s); inline-size: calc(var(--spacing) * %[1]s);"},
	"w-":     {Attribute: "inline-size: calc(var(--spacing) * %s);"},
	"h-":     {Attribute: "block-size: calc(var(--spacing) * %s);"},
	"max-w-": {Attribute: "max-inline-size: calc(var(--spacing) * %s);"},
	"max-h-": {Attribute: "max-block-size: calc(var(--spacing) * %s);"},
	"min-w-": {Attribute: "min-inline-size: calc(var(--spacing) * %s);"},
	"min-h-": {Attribute: "min-block-size: calc(var(--spacing) * %s);"},

	// Enhanced Padding with logical properties (v4)
	"p-":  {Attribute: "padding: calc(var(--spacing) * %s);"},
	"px-": {Attribute: "padding-inline: calc(var(--spacing) * %s);"},
	"py-": {Attribute: "padding-block: calc(var(--spacing) * %s);"},
	"pt-": {Attribute: "padding-block-start: calc(var(--spacing) * %s);"},
	"pb-": {Attribute: "padding-block-end: calc(var(--spacing) * %s);"},
	"pl-": {Attribute: "padding-inline-start: calc(var(--spacing) * %s);"},
	"pr-": {Attribute: "padding-inline-end: calc(var(--spacing) * %s);"},
	"ps-": {Attribute: "padding-inline-start: calc(var(--spacing) * %s);"},
	"pe-": {Attribute: "padding-inline-end: calc(var(--spacing) * %s);"},

	// Enhanced Margin with logical properties (v4)
	"m-":  {Attribute: "margin: calc(var(--spacing) * %s);"},
	"mx-": {Attribute: "margin-inline: calc(var(--spacing) * %s);"},
	"my-": {Attribute: "margin-block: calc(var(--spacing) * %s);"},
	"mt-": {Attribute: "margin-block-start: calc(var(--spacing) * %s);"},
	"mb-": {Attribute: "margin-block-end: calc(var(--spacing) * %s);"},
	"ml-": {Attribute: "margin-inline-start: calc(var(--spacing) * %s);"},
	"mr-": {Attribute: "margin-inline-end: calc(var(--spacing) * %s);"},
	"ms-": {Attribute: "margin-inline-start: calc(var(--spacing) * %s);"},
	"me-": {Attribute: "margin-inline-end: calc(var(--spacing) * %s);"},

	// Enhanced negative margins with logical properties (v4)
	"-m-":  {Attribute: "margin: calc(var(--spacing) * %s * -1);"},
	"-mx-": {Attribute: "margin-inline: calc(var(--spacing) * %s * -1);"},
	"-my-": {Attribute: "margin-block: calc(var(--spacing) * %s * -1);"},
	"-mt-": {Attribute: "margin-block-start: calc(var(--spacing) * %s * -1);"},
	"-mb-": {Attribute: "margin-block-end: calc(var(--spacing) * %s * -1);"},
	"-ml-": {Attribute: "margin-inline-start: calc(var(--spacing) * %s * -1);"},
	"-mr-": {Attribute: "margin-inline-end: calc(var(--spacing) * %s * -1);"},
	"-ms-": {Attribute: "margin-inline-start: calc(var(--spacing) * %s * -1);"},
	"-me-": {Attribute: "margin-inline-end: calc(var(--spacing) * %s * -1);"},

	// Enhanced Grid with subgrid support (v4)
	"grid-":      {Attribute: "display: grid;"},
	"cols-":      {Attribute: "grid-template-columns: repeat(%s, minmax(0, 1fr));"},
	"rows-":      {Attribute: "grid-template-rows: repeat(%s, minmax(0, 1fr));"},
	"col-span-":  {Attribute: "grid-column: span %s / span %s;"},
	"row-span-":  {Attribute: "grid-row: span %s / span %s;"},
	"col-start-": {Attribute: "grid-column-start: %s;"},
	"col-end-":   {Attribute: "grid-column-end: %s;"},
	"row-start-": {Attribute: "grid-row-start: %s;"},
	"row-end-":   {Attribute: "grid-row-end: %s;"},
	"gap-":       {Attribute: "gap: calc(var(--spacing) * %s);"},
	"gap-x-":     {Attribute: "column-gap: calc(var(--spacing) * %s);"},
	"gap-y-":     {Attribute: "row-gap: calc(var(--spacing) * %s);"},

	// Enhanced positioning with logical properties (v4)
	"top-":     {Attribute: "inset-block-start: calc(var(--spacing) * %s);"},
	"bottom-":  {Attribute: "inset-block-end: calc(var(--spacing) * %s);"},
	"left-":    {Attribute: "inset-inline-start: calc(var(--spacing) * %s);"},
	"right-":   {Attribute: "inset-inline-end: calc(var(--spacing) * %s);"},
	"start-":   {Attribute: "inset-inline-start: calc(var(--spacing) * %s);"},
	"end-":     {Attribute: "inset-inline-end: calc(var(--spacing) * %s);"},
	"inset-":   {Attribute: "inset: calc(var(--spacing) * %s);"},
	"inset-x-": {Attribute: "inset-inline: calc(var(--spacing) * %s);"},
	"inset-y-": {Attribute: "inset-block: calc(var(--spacing) * %s);"},

	// Negative positioning
	"-top-":    {Attribute: "inset-block-start: calc(var(--spacing) * %s * -1);"},
	"-bottom-": {Attribute: "inset-block-end: calc(var(--spacing) * %s * -1);"},
	"-left-":   {Attribute: "inset-inline-start: calc(var(--spacing) * %s * -1);"},
	"-right-":  {Attribute: "inset-inline-end: calc(var(--spacing) * %s * -1);"},
	"-start-":  {Attribute: "inset-inline-start: calc(var(--spacing) * %s * -1);"},
	"-end-":    {Attribute: "inset-inline-end: calc(var(--spacing) * %s * -1);"},

	// 3D Transform utilities (new in v4)
	"rotate-x-":    {Attribute: "transform: rotateX(%sdeg);"},
	"rotate-y-":    {Attribute: "transform: rotateY(%sdeg);"},
	"rotate-z-":    {Attribute: "transform: rotateZ(%sdeg);"},
	"scale-x-":     {Attribute: "transform: scaleX(%s);"},
	"scale-y-":     {Attribute: "transform: scaleY(%s);"},
	"scale-z-":     {Attribute: "transform: scaleZ(%s);"},
	"translate-z-": {Attribute: "transform: translateZ(calc(var(--spacing) * %s));"},

	// Enhanced text and font utilities (v4)
	"text-":     {Attribute: "font-size: var(--text-%s);"},
	"leading-":  {Attribute: "line-height: %s;"},
	"tracking-": {Attribute: "letter-spacing: %sem;"},
	"font-":     {Attribute: "font-weight: var(--font-weight-%s);"},

	// Enhanced border utilities (v4)
	"border-":   {Attribute: "border-width: %spx;"},
	"border-t-": {Attribute: "border-block-start-width: %spx;"},
	"border-b-": {Attribute: "border-block-end-width: %spx;"},
	"border-l-": {Attribute: "border-inline-start-width: %spx;"},
	"border-r-": {Attribute: "border-inline-end-width: %spx;"},
	"border-s-": {Attribute: "border-inline-start-width: %spx;"},
	"border-e-": {Attribute: "border-inline-end-width: %spx;"},
	"rounded-":  {Attribute: "border-radius: var(--radius-%s);"},

	// Enhanced opacity and backdrop utilities (v4)
	"opacity-":          {Attribute: "opacity: 0.%s;"},
	"backdrop-blur-":    {Attribute: "backdrop-filter: blur(%spx);"},
	"backdrop-opacity-": {Attribute: "backdrop-filter: opacity(0.%s);"},

	// Container queries (new in v4)
	"@container-type-": {Attribute: "container-type: %s;"},
	"@container-name-": {Attribute: "container-name: %s;"},

	// Field sizing (new in v4)
	"field-sizing-": {Attribute: "field-sizing: %s;"},
}

// Enhanced bracket recipes for Tailwind v4 arbitrary values
var BracketRecipes = map[string]ClassRecipe{
	// Background utilities
	"bg-url-": {Attribute: "background-image: url(%s);"},
	"bg-":     {Attribute: "background: %s;"},

	// Enhanced color utilities with arbitrary values
	"text-color-":   {Attribute: "color: %s;"},
	"bg-color-":     {Attribute: "background-color: %s;"},
	"border-color-": {Attribute: "border-color: %s;"},
	"ring-color-":   {Attribute: "--tw-ring-color: %s;"},

	// Enhanced sizing with arbitrary values
	"w-":     {Attribute: "width: %s;"},
	"h-":     {Attribute: "height: %s;"},
	"size-":  {Attribute: "width: %s; height: %s;"},
	"max-w-": {Attribute: "max-width: %s;"},
	"max-h-": {Attribute: "max-height: %s;"},
	"min-w-": {Attribute: "min-width: %s;"},
	"min-h-": {Attribute: "min-height: %s;"},

	// Enhanced spacing with arbitrary values
	"p-":  {Attribute: "padding: %s;"},
	"px-": {Attribute: "padding-inline: %s;"},
	"py-": {Attribute: "padding-block: %s;"},
	"pt-": {Attribute: "padding-block-start: %s;"},
	"pb-": {Attribute: "padding-block-end: %s;"},
	"pl-": {Attribute: "padding-inline-start: %s;"},
	"pr-": {Attribute: "padding-inline-end: %s;"},
	"ps-": {Attribute: "padding-inline-start: %s;"},
	"pe-": {Attribute: "padding-inline-end: %s;"},

	"m-":  {Attribute: "margin: %s;"},
	"mx-": {Attribute: "margin-inline: %s;"},
	"my-": {Attribute: "margin-block: %s;"},
	"mt-": {Attribute: "margin-block-start: %s;"},
	"mb-": {Attribute: "margin-block-end: %s;"},
	"ml-": {Attribute: "margin-inline-start: %s;"},
	"mr-": {Attribute: "margin-inline-end: %s;"},
	"ms-": {Attribute: "margin-inline-start: %s;"},
	"me-": {Attribute: "margin-inline-end: %s;"},

	// Enhanced typography with arbitrary values
	"text-size-":   {Attribute: "font-size: %s;"},
	"leading-":     {Attribute: "line-height: %s;"},
	"tracking-":    {Attribute: "letter-spacing: %s;"},
	"font-weight-": {Attribute: "font-weight: %s;"},

	// Enhanced transforms with arbitrary values (v4)
	"rotate-":      {Attribute: "transform: rotate(%s);"},
	"rotate-x-":    {Attribute: "transform: rotateX(%s);"},
	"rotate-y-":    {Attribute: "transform: rotateY(%s);"},
	"rotate-z-":    {Attribute: "transform: rotateZ(%s);"},
	"scale-":       {Attribute: "transform: scale(%s);"},
	"scale-x-":     {Attribute: "transform: scaleX(%s);"},
	"scale-y-":     {Attribute: "transform: scaleY(%s);"},
	"scale-z-":     {Attribute: "transform: scaleZ(%s);"},
	"translate-x-": {Attribute: "transform: translateX(%s);"},
	"translate-y-": {Attribute: "transform: translateY(%s);"},
	"translate-z-": {Attribute: "transform: translateZ(%s);"},
	"skew-x-":      {Attribute: "transform: skewX(%s);"},
	"skew-y-":      {Attribute: "transform: skewY(%s);"},

	// Enhanced positioning with arbitrary values
	"top-":     {Attribute: "inset-block-start: %s;"},
	"bottom-":  {Attribute: "inset-block-end: %s;"},
	"left-":    {Attribute: "inset-inline-start: %s;"},
	"right-":   {Attribute: "inset-inline-end: %s;"},
	"start-":   {Attribute: "inset-inline-start: %s;"},
	"end-":     {Attribute: "inset-inline-end: %s;"},
	"inset-":   {Attribute: "inset: %s;"},
	"inset-x-": {Attribute: "inset-inline: %s;"},
	"inset-y-": {Attribute: "inset-block: %s;"},

	// Enhanced grid with arbitrary values
	"cols-":      {Attribute: "grid-template-columns: %s;"},
	"rows-":      {Attribute: "grid-template-rows: %s;"},
	"col-span-":  {Attribute: "grid-column: %s;"},
	"row-span-":  {Attribute: "grid-row: %s;"},
	"col-start-": {Attribute: "grid-column-start: %s;"},
	"col-end-":   {Attribute: "grid-column-end: %s;"},
	"row-start-": {Attribute: "grid-row-start: %s;"},
	"row-end-":   {Attribute: "grid-row-end: %s;"},
	"gap-":       {Attribute: "gap: %s;"},
	"gap-x-":     {Attribute: "column-gap: %s;"},
	"gap-y-":     {Attribute: "row-gap: %s;"},

	// Enhanced flex with arbitrary values
	"basis-":  {Attribute: "flex-basis: %s;"},
	"grow-":   {Attribute: "flex-grow: %s;"},
	"shrink-": {Attribute: "flex-shrink: %s;"},
	"order-":  {Attribute: "order: %s;"},

	// Enhanced effects with arbitrary values
	"shadow-":           {Attribute: "box-shadow: %s;"},
	"drop-shadow-":      {Attribute: "filter: drop-shadow(%s);"},
	"blur-":             {Attribute: "filter: blur(%s);"},
	"brightness-":       {Attribute: "filter: brightness(%s);"},
	"contrast-":         {Attribute: "filter: contrast(%s);"},
	"grayscale-":        {Attribute: "filter: grayscale(%s);"},
	"hue-rotate-":       {Attribute: "filter: hue-rotate(%s);"},
	"invert-":           {Attribute: "filter: invert(%s);"},
	"saturate-":         {Attribute: "filter: saturate(%s);"},
	"sepia-":            {Attribute: "filter: sepia(%s);"},
	"backdrop-blur-":    {Attribute: "backdrop-filter: blur(%s);"},
	"backdrop-opacity-": {Attribute: "backdrop-filter: opacity(%s);"},

	// Enhanced border with arbitrary values
	"border-width-": {Attribute: "border-width: %s;"},
	"border-t-":     {Attribute: "border-block-start-width: %s;"},
	"border-b-":     {Attribute: "border-block-end-width: %s;"},
	"border-l-":     {Attribute: "border-inline-start-width: %s;"},
	"border-r-":     {Attribute: "border-inline-end-width: %s;"},
	"border-s-":     {Attribute: "border-inline-start-width: %s;"},
	"border-e-":     {Attribute: "border-inline-end-width: %s;"},
	"rounded-":      {Attribute: "border-radius: %s;"},
	"rounded-t-":    {Attribute: "border-start-start-radius: %s; border-start-end-radius: %s;"},
	"rounded-b-":    {Attribute: "border-end-start-radius: %s; border-end-end-radius: %s;"},
	"rounded-l-":    {Attribute: "border-start-start-radius: %s; border-end-start-radius: %s;"},
	"rounded-r-":    {Attribute: "border-start-end-radius: %s; border-end-end-radius: %s;"},
	"rounded-tl-":   {Attribute: "border-start-start-radius: %s;"},
	"rounded-tr-":   {Attribute: "border-start-end-radius: %s;"},
	"rounded-bl-":   {Attribute: "border-end-start-radius: %s;"},
	"rounded-br-":   {Attribute: "border-end-end-radius: %s;"},

	// Enhanced opacity with arbitrary values
	"opacity-": {Attribute: "opacity: %s;"},

	// Z-index with arbitrary values
	"z-": {Attribute: "z-index: %s;"},

	// Enhanced animation with arbitrary values (v4)
	"animate-":  {Attribute: "animation: %s;"},
	"duration-": {Attribute: "animation-duration: %s;"},
	"delay-":    {Attribute: "animation-delay: %s;"},
	"ease-":     {Attribute: "animation-timing-function: %s;"},

	// Content utilities with arbitrary values (v4)
	"content-": {Attribute: "content: %s;"},

	// Arbitrary CSS properties (v4)
	"css-": {Attribute: "%s;"},
}
