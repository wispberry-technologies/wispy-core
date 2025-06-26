package models

import "time"

// APIEndpointConfig represents a single API endpoint configuration
type APIEndpointConfig struct {
	Method      string                 `json:"method"`
	URL         string                 `json:"url"`
	Description string                 `json:"description"`
	Handler     string                 `json:"handler"`
	Table       string                 `json:"table,omitempty"`    // Target database table
	Database    string                 `json:"database,omitempty"` // Target database file
	Protected   *ProtectionConfig      `json:"protected,omitempty"`
	QueryParams map[string]ParamConfig `json:"queryParams,omitempty"`
	BodyParams  map[string]ParamConfig `json:"bodyParams,omitempty"`
	Validation  *ValidationConfig      `json:"validation,omitempty"`
	Response    *ResponseConfig        `json:"response,omitempty"`
}

// ProtectionConfig defines authentication and authorization for an endpoint
type ProtectionConfig struct {
	Type           string   `json:"type"`             // "user", "admin", "api_key", "none"
	Roles          []string `json:"roles"`            // Required roles
	SQLColumnMatch string   `json:"sql_column_match"` // Column to match against user ID
	Permissions    []string `json:"permissions"`      // Required permissions
}

// ParamConfig defines parameter configuration and validation
type ParamConfig struct {
	Type        string      `json:"type"` // "string", "integer", "float", "boolean", "array"
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`       // Valid values
	Min         *float64    `json:"min,omitempty"`        // Min value for numbers
	Max         *float64    `json:"max,omitempty"`        // Max value for numbers
	MinLength   *int        `json:"min_length,omitempty"` // Min length for strings
	MaxLength   *int        `json:"max_length,omitempty"` // Max length for strings
	Pattern     string      `json:"pattern,omitempty"`    // Regex pattern for validation
}

// ValidationConfig defines additional validation rules
type ValidationConfig struct {
	UniqueFields   []string `json:"unique_fields,omitempty"`    // Fields that must be unique in table
	RequiredFields []string `json:"required_fields,omitempty"`  // Required fields for create/update
	ReadOnlyFields []string `json:"read_only_fields,omitempty"` // Fields that cannot be modified
}

// ResponseConfig defines response formatting
type ResponseConfig struct {
	Fields   []string          `json:"fields,omitempty"`   // Fields to include in response
	Exclude  []string          `json:"exclude,omitempty"`  // Fields to exclude from response
	Format   string            `json:"format,omitempty"`   // "json", "xml", "csv"
	Metadata map[string]string `json:"metadata,omitempty"` // Additional response metadata
}

// APIEndpointsConfig represents the complete API configuration for a site
type APIEndpointsConfig map[string]APIEndpointConfig

// DatabaseConnectionConfig represents database connection settings for a site
type DatabaseConnectionConfig struct {
	DefaultDB string            `json:"default_db"`
	Databases map[string]string `json:"databases"` // name -> file path
}

// APIRequestContext contains context information for API requests
type APIRequestContext struct {
	SiteInstance *SiteInstance
	Config       *APIEndpointConfig
	UserID       string
	UserRoles    []string
	RequestID    string
	StartTime    time.Time
	Database     string
	Table        string
	RequestBody  map[string]interface{} // Parsed request body
}

// SQLOperation represents the type of SQL operation
type SQLOperation string

const (
	SQLSelect SQLOperation = "SELECT"
	SQLInsert SQLOperation = "INSERT"
	SQLUpdate SQLOperation = "UPDATE"
	SQLDelete SQLOperation = "DELETE"
)

// SQLQueryBuilder helps build dynamic SQL queries
type SQLQueryBuilder struct {
	Operation  SQLOperation
	Table      string
	Fields     []string
	Values     map[string]interface{}
	Conditions []string
	Args       []interface{}
	OrderBy    string
	Limit      *int
	Offset     *int
	Joins      []string
}

// APIError represents an API error response
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      *APIError   `json:"error,omitempty"`
	Metadata   interface{} `json:"metadata,omitempty"`
	Pagination interface{} `json:"pagination,omitempty"`
}

// PaginationInfo represents pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}
