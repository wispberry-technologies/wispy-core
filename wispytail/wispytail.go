package wispytail

import (
	"regexp"
	"strings"
	"wispy-core/common"
	"wispy-core/wispytail/core"
)

// ThemeConfig represents the configuration for generating Tailwind CSS theme
type ThemeConf *core.ThemeConfig

func DefaultThemeConfig() ThemeConf {
	return core.DefaultThemeConfig()
}

func GenerateCssBaseLayer() string {
	// Generate the base CSS layer for Tailwind
	// This includes the default styles and utility classes
	return core.GenerateCssBaseLayer()
}

func GenerateThemeLayer(themeConfig ThemeConf) string {
	// Generate the theme layer CSS based on the provided theme configuration
	if themeConfig == nil {
		themeConfig = core.DefaultThemeConfig()
	}
	return core.GenerateThemeLayer(themeConfig)
}

// func (th *ThemeConfig) DefaultThemeConfig() *ThemeConfig {
// 	if th == nil {
// 		return DefaultThemeConfig()
// 	}
// 	return th
// }

var baseTrie *common.Trie

func init() {
	// Initialize the BaseTrie with utility classes
	baseTrie = common.NewTrie()
	baseTrie = core.BuildFullTrie(baseTrie)
}

func GetBaseTrie() *common.Trie {
	if baseTrie == nil {
		baseTrie = common.NewTrie()
		baseTrie = core.BuildFullTrie(baseTrie)
	}
	return baseTrie
}

func Generate(input string, themeConfig ThemeConf, trie *common.Trie) string {
	// Generate CSS for the given classes using the provided theme configuration
	extractedClasses := ExtractClasses(input)
	css := core.GenerateCSSFromClasses(extractedClasses, themeConfig, trie)

	return css
}

// GenerateWithBaseTheme generates CSS using the base theme and the provided theme configuration
// It combines the base CSS, theme CSS, and utility classes into a single CSS string.
// This is useful for generating a complete CSS file that includes both the base styles and the theme styles.
func GenerateWithBaseTheme(input string, themeConfig ThemeConf, trie *common.Trie) string {
	// Generate CSS for the given classes using the provided theme configuration
	extractedClasses := ExtractClasses(input)
	baseCss := core.GenerateCssBaseLayer()
	themeCss := core.GenerateThemeLayer(themeConfig)
	utilCss := core.GenerateCSSFromClasses(extractedClasses, themeConfig, trie)

	css := baseCss + "\n" + themeCss + "\n" + utilCss

	return css
}

func PopulateTrieWithUtils(trie *common.Trie) *common.Trie {
	newTrie := core.BuildFullTrie(trie)
	return newTrie
}

// --- CSS Generation ---
// ExtractClasses parses HTML from the reader and extracts unique class names in order.
func ExtractClasses(input string) []string {
	seen := make(map[string]bool) // Track unique class names
	var classList []string        // Preserve order
	regex := regexp.MustCompile(`class\s*=\s*"([^"]+)"`)
	matches := regex.FindAllString(input, -1)
	for _, match := range matches {
		match = match[7 : len(match)-1]      // Remove 'class="' and '"'
		classes := strings.Split(match, " ") // Split class names by space
		for _, class := range classes {
			if !seen[class] {
				seen[class] = true                   // Mark class as seen
				classList = append(classList, class) // Add to list
			}
		}
	}

	return classList
}
