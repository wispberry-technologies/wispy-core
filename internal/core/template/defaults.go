package template

import "wispy-core/pkg/models"

// GetDefaultFilters returns an empty filter map for now
func GetDefaultFilters() models.FilterMap {
	return models.FilterMap{
		"upcase":     UpcaseFilter,
		"downcase":   DowncaseFilter,
		"split":      SplitFilter,
		"remove":     RemoveFilter,
		"replace":    ReplaceFilter,
		"strip":      StripFilter,
		"trim":       TrimFilter,
		"append":     AppendFilter,
		"prepend":    PrependFilter,
		"truncate":   TruncateFilter,
		"slice":      SliceFilter,
		"join":       JoinFilter,
		"capitalize": CapitalizeFilter,
		"default":    DefaultValueFilter,
	}
}

// GetDefaultFunctions returns the default set of template functions
func GetDefaultFunctions() models.FunctionMap {
	return models.FunctionMap{
		"if":      IfTemplate,
		"for":     ForTag,
		"define":  DefineTag,
		"include": IncludeTag,
		"render":  RenderTag,
		"block":   BlockTag,
		"extend":  ExtendTag,
		// Add more built-in functions here
	}
}
