package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"wispy-core/common"
	"wispy-core/core"
	"wispy-core/models"
)

// Define colors for logging
const (
	colorCyan   = "\033[36m"
	colorGrey   = "\033[90m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Panicf("%sWarning: Error loading .env file: %v%s", colorRed, err, colorReset)
	}

	// get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		panic("Failed to get current working directory: " + err.Error())
	}
	// Set the WISPY_CORE_ROOT environment variable to the current directory
	os.Setenv("WISPY_CORE_ROOT", currentDir)
}

// RegisterPageRoutes registers a handler for each page in all sites, using the correct layout.
func RegisterPageRoutes(r chi.Router, sites map[string]*models.SiteInstance) {
	startTime := time.Now()
	log.Printf("[INFO] Registering page routes...")
	pageCount := 0
	for _, site := range sites {
		for url, page := range site.Pages {
			pageCount++
			// Capture variables for closure
			pageCopy := page
			siteCopy := site
			urlCopy := url
			r.Get(urlCopy, func(w http.ResponseWriter, req *http.Request) {
				start := time.Now()
				ctx := &models.TemplateContext{Data: map[string]interface{}{"Page": pageCopy, "Site": siteCopy.Site}, Request: req}

				// Determine layout to use
				layoutName := pageCopy.Layout
				if layoutName == "" {
					layoutName = "default"
				}

				layoutPath := common.RootSitesPath(siteCopy.Domain, "layouts", layoutName+".html")
				layoutContent := ""
				if f, err := os.ReadFile(layoutPath); err == nil {
					layoutContent = string(f)
				}

				var result string
				var errors = []error{}
				engine := core.NewTemplateEngine(core.DefaultFunctionMap)
				if layoutContent != "" {
					result, errors = core.Render(pageCopy.Content+layoutContent, engine, ctx)
					if len(errors) > 0 {
						log.Printf("[ERROR] Failed to render page %s: %v", urlCopy, errors)
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						return
					}
				} else {
					result, _ = core.Render(pageCopy.Content, engine, ctx)
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(result))
				dur := time.Since(start)
				log.Printf("[INFO] Rendered %s in %s", urlCopy, dur)
			})
		}
	}
	log.Printf("[INFO] Registered %d page routes in %s", pageCount, time.Since(startTime))
}

func main() {
	// Get configuration from environment
	port := common.GetEnv("PORT", "8080")
	host := common.GetEnv("HOST", "localhost")
	sitesPath := common.MustGetEnv("SITES_PATH") // Required for system to function
	env := common.GetEnv("ENV", "development")

	// Enable rate limiting based on config
	requestsPerSecond := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_SECOND", 12)
	requestsPerMinute := common.GetEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 240)

	// Log startup information
	log.Printf("%s🚀 Starting Wispy Core CMS%s", colorCyan, colorReset)
	log.Printf("%s📁 Sites directory: %s%s", colorGrey, sitesPath, colorReset)
	log.Printf("%s🌍 Environment: %s%s", colorGrey, env, colorReset)
	log.Printf("%s🔧 Host: %s, Port: %s%s", colorGrey, host, port, colorReset)
	log.Printf("%s📊 Rate limiting: %d req/sec, %d req/min%s", colorGrey, requestsPerSecond, requestsPerMinute, colorReset)

	// Load all sites and their pages from the sites directory
	sites, err := core.LoadAllSites(sitesPath)
	if err != nil {
		log.Fatalf("Failed to load sites: %v", err)
	}

	// Set up the chi router and register routes for each page
	r := chi.NewRouter()
	RegisterPageRoutes(r, sites)

	// Start the HTTP server
	addr := host + ":" + port
	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	log.Printf("%s✅ Server starting on http://%s%s", colorGreen, addr, colorReset)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("%sServer failed to start: %v%s", colorRed, err, colorReset)
	}
}
