// Package assets provides functions for asset management
package assets

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"wispy-core/internal/models"
)

// AssetType represents the type of asset
type AssetType string

const (
	// CSS is a CSS asset type
	CSS AssetType = "css"
	// JS is a JavaScript asset type
	JS AssetType = "js"
	// CSSInline is an inline CSS asset type
	CSSInline AssetType = "css-inline"
	// JSInline is an inline JavaScript asset type
	JSInline AssetType = "js-inline"
)

// AssetLocation represents where in the document the asset should be placed
type AssetLocation string

const (
	// Head location puts the asset in the document head
	Head AssetLocation = "head"
	// Footer location puts the asset at the end of the document body
	Footer AssetLocation = "footer"
	// PreFooter location puts the asset before other footer assets
	PreFooter AssetLocation = "pre-footer"
)

// Asset represents a CSS or JavaScript asset
type Asset struct {
	Type     AssetType
	Path     string
	Content  string
	Location AssetLocation
	Priority int
}

// ImportAsset imports a CSS or JavaScript asset
// This is a pure function that adds an asset to the template context
func ImportAsset(ctx *models.TemplateContext, assetType AssetType, path string, location AssetLocation) error {
	// Validate asset type
	switch assetType {
	case CSS, JS, CSSInline, JSInline:
		// Valid types
	default:
		return fmt.Errorf("invalid asset type: %s", assetType)
	}

	// Validate path
	if path == "" {
		return fmt.Errorf("asset path cannot be empty")
	}

	// Create map for asset type if it doesn't exist
	assetMap := ctx.InternalContext.Flags
	if _, ok := assetMap["assets"]; !ok {
		assetMap["assets"] = make(map[AssetType][]Asset)
	}

	assets := assetMap["assets"].(map[AssetType][]Asset)

	// Check for duplicates
	for _, asset := range assets[assetType] {
		if asset.Path == path {
			// Asset already imported
			return nil
		}
	}

	// Add asset
	newAsset := Asset{
		Type:     assetType,
		Path:     path,
		Location: location,
		Priority: 0,
	}

	assets[assetType] = append(assets[assetType], newAsset)
	return nil
}

// LoadAssetContent loads the content of an asset from disk or URL
// This has side effects (file IO, network requests) and returns content
func LoadAssetContent(siteBasePath string, asset Asset) (string, error) {
	// Check if asset is a URL
	if strings.HasPrefix(asset.Path, "http://") || strings.HasPrefix(asset.Path, "https://") {
		// Remote asset
		return loadRemoteAsset(asset.Path)
	}

	// Local asset
	return loadLocalAsset(siteBasePath, asset.Path)
}

// loadRemoteAsset loads an asset from a URL
func loadRemoteAsset(path string) (string, error) {
	// Security check: only allow https URLs
	if !strings.HasPrefix(path, "https://") {
		return "", fmt.Errorf("only HTTPS URLs are allowed for remote assets")
	}

	// Parse URL
	parsedURL, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Fetch asset
	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return "", fmt.Errorf("error fetching remote asset: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading remote asset: %w", err)
	}

	return string(body), nil
}

// loadLocalAsset loads an asset from the local filesystem
func loadLocalAsset(siteBasePath, path string) (string, error) {
	// Security check: ensure path is within site's assets or public directory
	if !(strings.HasPrefix(path, "assets/") || strings.HasPrefix(path, "public/")) {
		return "", fmt.Errorf("local assets must be in assets/ or public/ directories")
	}

	// Build full path
	fullPath := filepath.Join(siteBasePath, path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("asset file not found: %s", path)
	}

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("error reading asset file: %w", err)
	}

	return string(content), nil
}
