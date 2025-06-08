package common

import (
	"context"
	"net/http"
)

// -----------------
// Helper functions for HTTP cookies
// ----------
// // getSameSiteAttribute converts string to http.SameSite
// func getSameSiteAttribute(sameSite string) http.SameSite {
// 	switch sameSite {
// 	case "Strict":
// 		return http.SameSiteStrictMode
// 	case "Lax":
// 		return http.SameSiteLaxMode
// 	case "None":
// 		return http.SameSiteNoneMode
// 	default:
// 		return http.SameSiteStrictMode
// 	}
// }

// -----------------
// Helper functions for request context
// -----------------
type contextKey string

const (
	routeParamsKey = contextKey("routeParams")
	matchedPageKey = contextKey("matchedPage")
)

// SetRouteParam sets a route parameter in the request context
func SetRouteParam(ctx context.Context, key, value string) context.Context {
	params, ok := ctx.Value(routeParamsKey).(map[string]string)
	if !ok {
		params = make(map[string]string)
	}
	params[key] = value
	return context.WithValue(ctx, routeParamsKey, params)
}

// SetMatchedPage sets the matched page in the request context
func SetMatchedPage(ctx context.Context, page *Page) context.Context {
	return context.WithValue(ctx, matchedPageKey, page)
}

// GetRouteParam gets a route parameter from the request context
func GetRouteParam(r *http.Request, key string) string {
	if params, ok := r.Context().Value(routeParamsKey).(map[string]string); ok {
		return params[key]
	}
	return ""
}

// GetMatchedPage gets the matched page from the request context
func GetMatchedPage(r *http.Request) *Page {
	if page, ok := r.Context().Value(matchedPageKey).(*Page); ok {
		return page
	}
	return nil
}
