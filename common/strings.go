package common

import (
	"crypto/rand"
	"fmt"
	"path/filepath"
	"strings"
)

// GetMapKeys returns a slice of sorted keys from a map[string]interface{}
func GetMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// SplitTagParts splits contents of a template tag into the tag name and its arguments.
// while respecting spaces, quotes round brackets
// It assumes the first part is the tag name and the rest are arguments.
// For example, "{{ if condition arg1 arg2 }}" would return "if" and ["condition", "arg1", "arg2"].
func SplitTagParts(contents string) (tag string, args []string) {
	var result []string
	var current strings.Builder
	inQuote := false
	inBrackets := 0 // Track nested brackets

	for _, char := range contents {
		switch char {
		case '(':
			inBrackets++ // Increment bracket count
			current.WriteRune(char)
		case ')':
			if inBrackets > 0 {
				inBrackets-- // Decrement bracket count
				current.WriteRune(char)
			} else {
				// If we encounter a closing bracket without an opening one, treat it as a normal character
				current.WriteRune(char)
			}
		case '"':
			inQuote = !inQuote // Toggle quote mode
		case ' ':
			if !inQuote {
				if current.Len() > 0 {
					result = append(result, current.String())
					current.Reset()
				}
				continue
			}
		}
		current.WriteRune(char)
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result[0], result[1:]
}

// DEPRECATED: FieldsRespectQuotes is deprecated, use SplitTagParts instead.
// FieldsRespectQuotes splits a string by spaces while respecting quoted substrings and removing empty values.
func FieldsRespectQuotes(s string) []string {
	var result []string
	var current strings.Builder
	inQuote := false

	for _, char := range s {
		switch char {
		case '"':
			inQuote = !inQuote // Toggle quote mode
		case ' ':
			if !inQuote {
				if current.Len() > 0 {
					result = append(result, current.String())
					current.Reset()
				}
				continue
			}
		}
		current.WriteRune(char)
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// ParseKeyValuePairs handles the key=value pairs parsing logic
func ParseKeyValuePairs(pairs []string) map[string]string {
	options := make(map[string]string)

	for _, pair := range pairs {
		if strings.Contains(pair, "=") {
			parts := strings.SplitN(pair, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			options[key] = value
		} else {
			options[pair] = "true"
		}
	}

	return options
}

// isQuotedString checks if a string is surrounded by single or double quotes
func IsQuotedString(s string) bool {
	if len(s) < 2 {
		return false
	}

	firstChar := s[0]
	lastChar := s[len(s)-1]

	// Check for matching quotes (either single or double)
	return (firstChar == '"' && lastChar == '"') || (firstChar == '\'' && lastChar == '\'')
}

// WrapBraces wraps a string with double curly braces, e.g., WrapBraces("foo") => "{{foo%}"
func WrapBraces(s string) string {
	return "{{" + s + "}}"
}

func WrapDoubleQuotes(s string) string {
	return "\"" + s + "\""
}

// GenerateUUID returns a random v4 UUID string
func GenerateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// isTemplateFile checks if a file is a template file
func IsTemplateFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".html" || ext == ".tmpl" || ext == ".tpl" || ext == ".gohtml"
}

func NormalizeTemplateName(rootDirPath, templatePath string) string {
	relName, _ := filepath.Rel(rootDirPath, templatePath)
	relName = strings.TrimSuffix(relName, filepath.Ext(relName)) // Remove file extension
	relName = strings.TrimPrefix(relName, rootDirPath+"/")
	return relName
}

// PathToRoute converts a file path to a URL route
func PathToRoute(pagePath string) string {
	// Remove .html extension
	route := strings.TrimSuffix(pagePath, ".html")

	// Convert file separators to URL separators
	route = strings.ReplaceAll(route, "\\", "/")

	// Handle index files
	route = strings.TrimSuffix(route, "/index")

	// Ensure route starts with /
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}

	// Handle root index
	if route == "/index" {
		route = "/"
	}

	return route
}
