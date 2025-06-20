package core

type contextKey string

const (
	// InstanceKey is the context key for the current instance
	InstanceKey = contextKey("instance")
	// PageKey is the context key for the current page
	PageKey = contextKey("page")
)
