package common

import "strings"

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

// WrapBraces wraps a string with double curly braces, e.g., WrapBraces("foo") => "{%foo%}"
func WrapBraces(s string) string {
	return "{{" + s + "}}"
}

// WrapTemplateDelims wraps a string with template delimiters, e.g., WrapTemplateDelims("foo") => "{%foo%}"
func WrapTemplateDelims(s string) string {
	return "{%" + s + "%}"
}
