// Package template provides filter functions for template rendering
package template

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// UpcaseFilter converts a string to uppercase
func UpcaseFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String {
		if str, ok := value.(string); ok {
			return strings.ToUpper(str)
		}
	}
	return value
}

// DowncaseFilter converts a string to lowercase
func DowncaseFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String {
		if str, ok := value.(string); ok {
			return strings.ToLower(str)
		}
	}
	return value
}

// SplitFilter splits a string by a delimiter
func SplitFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String && len(args) > 0 {
		str := value.(string)

		// If the delimiter has quotes, remove them
		delimiter := args[0]
		if len(delimiter) >= 2 && (delimiter[0] == '"' && delimiter[len(delimiter)-1] == '"' ||
			delimiter[0] == '\'' && delimiter[len(delimiter)-1] == '\'') {
			delimiter = delimiter[1 : len(delimiter)-1]
		}

		result := strings.Split(str, delimiter)

		// Trim spaces from results
		for i := range result {
			result[i] = strings.TrimSpace(result[i])
		}

		// Convert to []interface{} to be compatible with ForTemplate
		interfaceSlice := make([]interface{}, len(result))
		for i, v := range result {
			interfaceSlice[i] = v
		}
		return interfaceSlice
	}
	return value
}

// RemoveFilter removes all occurrences of a substring from a string
func RemoveFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String && len(args) > 0 {
		str := value.(string)
		return strings.ReplaceAll(str, args[0], "")
	}
	return value
}

// ReplaceFilter replaces all occurrences of a substring with another substring
func ReplaceFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String && len(args) > 1 {
		str := value.(string)
		return strings.ReplaceAll(str, args[0], args[1])
	}
	return value
}

// StripFilter removes all HTML tags from a string
func StripFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String {
		str := value.(string)
		re := regexp.MustCompile("<[^>]*>")
		return re.ReplaceAllString(str, "")
	}
	return value
}

// TrimFilter removes whitespace from both ends of a string
func TrimFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String {
		str := value.(string)
		return strings.TrimSpace(str)
	}
	return value
}

// JSONFilter converts a value to a JSON string
func JSONFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	slog.Debug("JSON filter called", "value", value, "type", valueType)

	jsonBytes, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		slog.Error("JSON marshaling failed", "error", err)
		return fmt.Sprintf("%+v", value)
	}
	result := string(jsonBytes)
	slog.Debug("JSON filter result", "result", result)
	return result
}

// AppendFilter appends a string to another string
func AppendFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String && len(args) > 0 {
		str := value.(string)
		// Clean quotes from argument if present
		arg := args[0]
		if len(arg) >= 2 && (arg[0] == '"' && arg[len(arg)-1] == '"' ||
			arg[0] == '\'' && arg[len(arg)-1] == '\'') {
			arg = arg[1 : len(arg)-1]
		}

		// Return the concatenated string
		return str + arg
	}
	return value
}

// PrependFilter prepends a string to another string
func PrependFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String && len(args) > 0 {
		str := value.(string)
		// Clean quotes from argument if present
		arg := args[0]
		if len(arg) >= 2 && (arg[0] == '"' && arg[len(arg)-1] == '"' ||
			arg[0] == '\'' && arg[len(arg)-1] == '\'') {
			arg = arg[1 : len(arg)-1]
		}

		// Return the concatenated string
		return arg + str
	}
	return value
}

// TruncateFilter truncates a string to a specified length and appends an ellipsis if needed
func TruncateFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String && len(args) > 0 {
		str := value.(string)
		length, err := strconv.Atoi(args[0])
		if err != nil || length <= 0 {
			return value
		}

		if len(str) <= length {
			return str
		}

		result := str[:length]
		ellipsis := "..."
		if len(args) > 1 {
			ellipsis = args[1]
		}

		return result + ellipsis
	}
	return value
}

// SliceFilter returns a substring or subarray of the input
func SliceFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if len(args) < 1 {
		return value
	}

	// Convert args to integers
	indices := make([]int, 0, len(args))
	for _, arg := range args {
		idx, err := strconv.Atoi(arg)
		if err != nil {
			return value
		}
		indices = append(indices, idx)
	}

	// Handle string slicing
	if valueType.Kind() == reflect.String {
		if str, ok := value.(string); ok {
			if len(indices) == 1 {
				// One argument: start from this index to end
				if indices[0] >= 0 && indices[0] < len(str) {
					return str[indices[0]:]
				}
			} else if len(indices) >= 2 {
				// Two arguments: start and end indices
				start, end := indices[0], indices[1]
				if start >= 0 && end <= len(str) && start <= end {
					return str[start:end]
				}
			}
		}
	}

	// Handle slice types
	if valueType.Kind() == reflect.Slice {
		slice := reflect.ValueOf(value)
		length := slice.Len()

		if len(indices) == 1 {
			// One argument: start from this index to end
			if indices[0] >= 0 && indices[0] < length {
				return slice.Slice(indices[0], length).Interface()
			}
		} else if len(indices) >= 2 {
			// Two arguments: start and end indices
			start, end := indices[0], indices[1]
			if start >= 0 && end <= length && start <= end {
				return slice.Slice(start, end).Interface()
			}
		}
	}

	return value
}

// JoinFilter joins a slice of strings with a separator
func JoinFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	// Handle array/slice of strings
	if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Array {
		slice := reflect.ValueOf(value)
		length := slice.Len()

		// Convert all elements to strings
		strSlice := make([]string, length)
		for i := 0; i < length; i++ {
			item := slice.Index(i).Interface()
			if str, ok := item.(string); ok {
				strSlice[i] = str
			} else {
				// Just use empty string for non-string values
				strSlice[i] = ""
			}
		}

		// Join with separator
		separator := " "
		if len(args) > 0 {
			separator = args[0]
		}

		return strings.Join(strSlice, separator)
	}

	return value
}

// CapitalizeFilter capitalizes the first character of a string
func CapitalizeFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if valueType.Kind() == reflect.String {
		str := value.(string)
		if str == "" {
			return str
		}
		return strings.ToUpper(str[:1]) + str[1:]
	}
	return value
}

// DefaultValueFilter returns a default value if the input is nil or empty
func DefaultValueFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	// Handle nil
	if value == nil && len(args) > 0 {
		return args[0]
	}

	// Handle empty string
	if valueType.Kind() == reflect.String {
		str := value.(string)
		if str == "" && len(args) > 0 {
			return args[0]
		}
	}

	return value
}

// ContainsFunction checks if a container (string, slice, array, map) contains a target value
func ContainsFilter(value interface{}, valueType reflect.Type, args []string) interface{} {
	if len(args) < 1 || value == nil {
		return false
	}

	target := args[0]

	// For slices and arrays
	if valueType.Kind() == reflect.Slice || valueType.Kind() == reflect.Array {
		valueSlice := reflect.ValueOf(value)
		if !valueSlice.IsValid() {
			return false
		}

		for i := 0; i < valueSlice.Len(); i++ {
			item := valueSlice.Index(i).Interface()

			// Try direct comparison
			if fmt.Sprintf("%v", item) == target {
				return true
			}

			// Try string comparison
			if itemStr, ok := item.(string); ok && itemStr == target {
				return true
			}
		}
		return false
	}

	// For strings
	if valueType.Kind() == reflect.String {
		valueStr, ok := value.(string)
		if !ok {
			return false
		}
		return strings.Contains(valueStr, target)
	}

	// For maps
	if valueType.Kind() == reflect.Map {
		valueMap := reflect.ValueOf(value)
		if !valueMap.IsValid() {
			return false
		}

		// Check if target is a key in the map
		for _, key := range valueMap.MapKeys() {
			if key.String() == target {
				return true
			}
		}
	}

	return false
}
