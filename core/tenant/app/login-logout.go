package app

import (
	"net/http"
	"net/url"
	"wispy-core/auth"
	"wispy-core/common"
	"wispy-core/config"
	"wispy-core/tpl"

	"github.com/go-playground/validator/v10"
)

// validator instance
var validate = validator.New()

func LoginHandler(cms WispyCms) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if user is already logged in
		_, err := auth.UserFromContext(r.Context())
		if err == nil {
			// User is already logged in, redirect to dashboard
			http.Redirect(w, r, "/wispy-cms/dashboard", http.StatusFound)
			return
		}

		if r.Method == http.MethodPost {
			handleLoginPost(w, r, cms)
			return
		}

		// Show login form
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		engine := cms.GetTemplateEngine()
		pagePath := "login.html"
		layoutPath := "default.html"

		// Get error message from query parameters
		errorParam := r.URL.Query().Get("error")
		var errorMessage string
		switch errorParam {
		case "invalid_data":
			errorMessage = "Please check your email and password."
		case "login_failed":
			errorMessage = "Invalid email or password. Please try again."
		case "parse_error":
			errorMessage = "There was an error processing your request. Please try again."
		default:
			errorMessage = ""
		}

		// Preserve form values on error
		email := r.URL.Query().Get("email")

		data := tpl.TemplateData{
			Title:       "Login",
			Description: "Login to Wispy CMS",
			Site: tpl.SiteData{
				Name:    "Wispy CMS",
				Domain:  r.Host,
				BaseURL: "https://" + r.Host,
			},
			Content: "",
			Data: map[string]interface{}{
				"__styles":     []string{},
				"__scripts":    []string{},
				"__inlineCSS":  "",
				"errorMessage": errorMessage,
				"hasError":     errorParam != "",
				"email":        email,
			},
		}

		state, err := renderCMSTemplate(engine, pagePath, layoutPath, data, cms.GetTheme())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		state.SetHeadTitle("Wispy CMS ~ Login")
		tpl.HtmlBaseRender(w, state)
	}
}

func handleLoginPost(w http.ResponseWriter, r *http.Request, cms WispyCms) {
	gConfig := config.GetGlobalConfig()
	authProvider := gConfig.GetCoreAuth()
	authMiddleware := gConfig.GetCoreAuthMiddleware()

	// Parse form data
	if err := r.ParseForm(); err != nil {
		common.Error("Failed to parse form data: %v", err)
		http.Redirect(w, r, "/wispy-cms/login?error=parse_error", http.StatusFound)
		return
	}

	loginReq := &auth.LoginRequest{
		Email:    r.FormValue("email"),
		Username: r.FormValue("username"),
		Password: r.FormValue("password"),
	}

	// Validate the request using the go-playground/validator
	if err := validate.Struct(loginReq); err != nil {
		common.Error("Login validation failed: %v", err)
		redirectURL := "/wispy-cms/login?error=invalid_data"
		if loginReq.Email != "" {
			redirectURL += "&email=" + url.QueryEscape(loginReq.Email)
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// Determine login identifier (email or username)
	loginIdentifier := loginReq.Email
	if loginIdentifier == "" {
		loginIdentifier = loginReq.Username
	}

	// Attempt login
	session, err := authProvider.Login(r.Context(), loginIdentifier, loginReq.Password)
	if err != nil {
		common.Error("Login failed: %v", err)
		redirectURL := "/wispy-cms/login?error=login_failed"
		if loginReq.Email != "" {
			redirectURL += "&email=" + url.QueryEscape(loginReq.Email)
		} else if loginReq.Username != "" {
			redirectURL += "&email=" + url.QueryEscape(loginReq.Username)
		}
		http.Redirect(w, r, redirectURL, http.StatusFound)
		return
	}

	// Set auth cookie
	authMiddleware.SetAuthCookie(w, session.Token)

	// Redirect to dashboard
	http.Redirect(w, r, "/wispy-cms/dashboard", http.StatusFound)
}

func LogoutHandler(cms WispyCms) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gConfig := config.GetGlobalConfig()
		authProvider := gConfig.GetCoreAuth()
		authMiddleware := gConfig.GetCoreAuthMiddleware()

		// Get current session
		session, err := auth.SessionFromContext(r.Context())
		if err == nil {
			// Logout from auth provider
			if err := authProvider.Logout(r.Context(), session.Token); err != nil {
				common.Error("Failed to logout session: %v", err)
			}
		}

		// Clear auth cookie
		authMiddleware.ClearAuthCookie(w)

		// Redirect to login page
		http.Redirect(w, r, "/wispy-cms/login", http.StatusFound)
	}
}
