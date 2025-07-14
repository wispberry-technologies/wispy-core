package app

import (
	"fmt"
	"net/http"
	"wispy-core/auth"
	"wispy-core/common"
	"wispy-core/core/tenant/app/providers"
	"wispy-core/tpl"

	"github.com/go-chi/chi/v5"
)

func FormsHandler(cms WispyCms) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, err := auth.UserFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		engine := cms.GetTemplateEngine()
		pagePath := "forms/index.html"
		layoutPath := "default.html"

		// Try to load the template first to check if it exists
		_, err = engine.LoadTemplate(pagePath)
		if err != nil {
			// Template doesn't exist, fall back to index.html
			pagePath = "index.html"
		}

		// Create data provider
		domain := common.NormalizeHost(r.Host)
		siteInstance, err := cms.GetSiteManager().GetSite(domain)
		if err != nil {
			http.Error(w, "Site not found for domain "+domain, http.StatusNotFound)
			return
		}

		providerManager := providers.NewProviderManager(siteInstance)
		defer providerManager.Close()

		// Get query parameters for search and filters
		searchQuery := r.URL.Query().Get("search")
		statusFilter := r.URL.Query().Get("status")
		sortBy := r.URL.Query().Get("sort")

		// Get forms stats
		formsProvider := providerManager.GetFormsProvider()
		formsStats, err := formsProvider.GetFormsStats(r.Context())
		if err != nil {
			common.Error("Failed to get forms stats: %v", err)
			// Use default stats on error
			formsStats = &providers.FormsStats{
				TotalForms:       0,
				SubmissionsToday: 0,
				ResponseRate:     0,
			}
		}

		// Get forms
		forms, err := formsProvider.GetForms(r.Context(), 0) // 0 = no limit
		if err != nil {
			common.Error("Failed to get forms: %v", err)
			// Use default forms on error
			forms = []providers.FormItem{}
		}

		// Convert forms to table format
		formRows := make([]map[string]interface{}, 0, len(forms))
		for _, form := range forms {
			statusBadge := ""
			if form.Status == "Active" {
				statusBadge = `<span class="badge badge-success">Active</span>`
			} else if form.Status == "Draft" {
				statusBadge = `<span class="badge badge-warning">Draft</span>`
			} else {
				statusBadge = `<span class="badge badge-error">Disabled</span>`
			}

			formRows = append(formRows, map[string]interface{}{
				"id": form.ID,
				"columns": []map[string]interface{}{
					{
						"text": form.Title,
						"html": fmt.Sprintf(`<div><strong>%s</strong><br><span class="text-sm text-base-content/70">%s</span></div>`, form.Title, form.Description),
					},
					{
						"text": form.Created,
					},
					{
						"text": fmt.Sprintf("%d", form.Submissions),
					},
					{
						"html": statusBadge,
					},
				},
			})
		}

		data := tpl.TemplateData{
			Title:       "Forms",
			Description: "Manage Forms",
			Site: tpl.SiteData{
				Name:    "Wispy CMS",
				Domain:  domain,
				BaseURL: "https://" + domain,
			},
			Content: "",
			Data: map[string]interface{}{
				"__styles":     []string{},
				"__scripts":    []string{},
				"__inlineCSS":  "",
				"user":         user,
				"pageTitle":    "Forms",
				"Search":       searchQuery,
				"StatusFilter": statusFilter,
				"SortBy":       sortBy,
				"Stats": map[string]interface{}{
					"totalForms":       formsStats.TotalForms,
					"submissionsToday": formsStats.SubmissionsToday,
					"responseRate":     formsStats.ResponseRate,
				},
				"Forms": formRows,
				"Pagination": map[string]interface{}{
					"current": 1,
					"total":   1,
					"hasPrev": false,
					"hasNext": false,
				},
			},
		}

		// Add provider functions to template context
		templateContext := providerManager.CreateTemplateContext(r.Context())
		for key, value := range templateContext {
			data.Data[key] = value
		}

		state, err := renderCMSTemplate(engine, pagePath, layoutPath, data, cms.GetTheme())
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		state.SetHeadTitle("Wispy CMS ~ Forms")
		tpl.HtmlBaseRender(w, state)
	}
}

func FormSubmissionsHandler(cms WispyCms) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, err := auth.UserFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		engine := cms.GetTemplateEngine()
		pagePath := "forms/submissions/index.html"
		layoutPath := "default.html" // Create data provider
		domain := common.NormalizeHost(r.Host)
		siteInstance, err := cms.GetSiteManager().GetSite(domain)
		if err != nil {
			http.Error(w, "Site not found for domain "+domain, http.StatusNotFound)
			return
		}

		providerManager := providers.NewProviderManager(siteInstance)
		defer providerManager.Close()

		// Get query parameters for filtering
		searchQuery := r.URL.Query().Get("search")
		formFilter := r.URL.Query().Get("form")
		statusFilter := r.URL.Query().Get("status")
		dateRange := r.URL.Query().Get("date_range")

		// Get submissions stats
		formsProvider := providerManager.GetFormsProvider()
		submissionsStats, err := formsProvider.GetSubmissionsStats(r.Context())
		if err != nil {
			common.Error("Failed to get submissions stats: %v", err)
			// Use default stats on error
			submissionsStats = &providers.SubmissionsStats{
				TotalSubmissions:  0,
				WeeklySubmissions: 0,
				WeeklyChange:      "→ 0%",
				UnreadSubmissions: 0,
				SpamFiltered:      0,
			}
		}

		// Get submissions
		submissions, err := formsProvider.GetSubmissions(r.Context(), formFilter, statusFilter, 0) // 0 = no limit
		if err != nil {
			common.Error("Failed to get submissions: %v", err)
			// Use default submissions on error
			submissions = []providers.SubmissionItem{}
		}

		// Convert submissions to table format
		submissionRows := make([]map[string]interface{}, 0, len(submissions))
		for _, submission := range submissions {
			statusBadge := ""
			if submission.Status == "New" {
				statusBadge = `<span class="badge badge-warning">New</span>`
			} else if submission.Status == "Read" {
				statusBadge = `<span class="badge badge-success">Read</span>`
			} else {
				statusBadge = `<span class="badge badge-info">Processed</span>`
			}

			nameWithInitials := fmt.Sprintf(`<div class="flex items-center gap-3">
				<div class="avatar placeholder">
					<div class="bg-neutral text-neutral-content rounded-full w-8 h-8">
						<span class="text-xs">%s</span>
					</div>
				</div>
				<div>
					<div class="font-medium">%s</div>
					<div class="text-sm text-base-content/70">%s</div>
				</div>
			</div>`, submission.Initials, submission.Name, submission.Email)

			submissionRows = append(submissionRows, map[string]interface{}{
				"id": submission.ID,
				"columns": []map[string]interface{}{
					{
						"text": submission.FormName,
					},
					{
						"html": nameWithInitials,
					},
					{
						"text": submission.Email,
					},
					{
						"text": submission.Subject,
					},
					{
						"text": submission.Date,
						"html": fmt.Sprintf(`<div>
							<div>%s</div>
							<div class="text-sm text-base-content/70">%s</div>
						</div>`, submission.Date, submission.TimeAgo),
					},
					{
						"html": statusBadge,
					},
				},
			})
		}

		data := tpl.TemplateData{
			Title:       "Form Submissions",
			Description: "View Form Submissions",
			Site: tpl.SiteData{
				Name:    "Wispy CMS",
				Domain:  domain,
				BaseURL: "https://" + domain,
			},
			Content: "",
			Data: map[string]interface{}{
				"__styles":    []string{},
				"__scripts":   []string{},
				"__inlineCSS": "",
				"user":        user,
				"pageTitle":   "Form Submissions",
				"Stats": map[string]interface{}{
					"totalSubmissions":  submissionsStats.TotalSubmissions,
					"weeklySubmissions": submissionsStats.WeeklySubmissions,
					"weeklyChange":      submissionsStats.WeeklyChange,
					"unreadSubmissions": submissionsStats.UnreadSubmissions,
					"spamFiltered":      submissionsStats.SpamFiltered,
				},
				"FormFilter":   formFilter,
				"DateRange":    dateRange,
				"StatusFilter": statusFilter,
				"Search":       searchQuery,
				"Submissions":  submissionRows,
				"Pagination": map[string]interface{}{
					"current": 1,
					"total":   1,
					"hasPrev": false,
					"hasNext": false,
				},
			},
		}

		// Add provider functions to template context
		templateContext := providerManager.CreateTemplateContext(r.Context())
		for key, value := range templateContext {
			data.Data[key] = value
		}

		state, err := renderCMSTemplate(engine, pagePath, layoutPath, data, cms.GetTheme())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		state.SetHeadTitle("Wispy CMS ~ Form Submissions")
		tpl.HtmlBaseRender(w, state)
	}
}

func FormSubmissionByIdHandler(cms WispyCms) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get current user from context
		user, err := auth.UserFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get form ID from URL parameter
		formID := chi.URLParam(r, "formID")
		if formID == "" {
			http.Error(w, "Form ID is required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		engine := cms.GetTemplateEngine()
		pagePath := "forms/submissions/index.html" // Use same template for now
		layoutPath := "default.html"

		// Create provider manager
		domain := common.NormalizeHost(r.Host)
		siteInstance, err := cms.GetSiteManager().GetSite(domain)
		if err != nil {
			http.Error(w, "Site not found for domain "+domain, http.StatusNotFound)
			return
		}

		providerManager := providers.NewProviderManager(siteInstance)
		defer providerManager.Close()

		data := tpl.TemplateData{
			Title:       "Form Submission Details",
			Description: "View Form Submission Details",
			Site: tpl.SiteData{
				Name:    "Wispy CMS",
				Domain:  domain,
				BaseURL: "https://" + domain,
			},
			Content: "",
			Data: map[string]interface{}{
				"__styles":    []string{},
				"__scripts":   []string{},
				"__inlineCSS": "",
				"user":        user,
				"pageTitle":   "Form Submission Details",
				"formID":      formID,
			},
		}

		// Add provider functions to template context
		templateContext := providerManager.CreateTemplateContext(r.Context())
		for key, value := range templateContext {
			data.Data[key] = value
		}

		state, err := renderCMSTemplate(engine, pagePath, layoutPath, data, cms.GetTheme())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		state.SetHeadTitle("Wispy CMS ~ Form Submission Details")
		tpl.HtmlBaseRender(w, state)
	}
}
