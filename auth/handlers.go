package auth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"wispy-core/common"
)

// AuthHandlers contains HTTP handlers for authentication
type AuthHandlers struct {
	authProvider AuthProvider
	middleware   *Middleware
	config       Config
}

// NewAuthHandlers creates new auth handlers
func NewAuthHandlers(authProvider AuthProvider, middleware *Middleware, config Config) *AuthHandlers {
	return &AuthHandlers{
		authProvider: authProvider,
		middleware:   middleware,
		config:       config,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" form:"email" validate:"omitempty,email"`
	Username string `json:"username" form:"username" validate:"omitempty,min=3,max=50"` // Optional, some systems allow login with username
	Password string `json:"password" form:"password" validate:"required,min=6"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
	User    *User  `json:"user,omitempty"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email           string `json:"email" form:"email" validate:"required,email"`
	Username        string `json:"username" form:"username" validate:"required,min=3,max=50"`
	Password        string `json:"password" form:"password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password" validate:"required,eqfield=Password"`
}

// RegisterResponse represents a registration response
type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	User    *User  `json:"user,omitempty"`
}

// APIResponse represents a generic API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// HandleLogin handles form-based login
func (h *AuthHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.PlainTextError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	// Parse the login request based on Content-Type
	var loginReq LoginRequest
	var jsonReq LoginRequest
	if err := parseFormRequest(r, &loginReq, &jsonReq); err != nil {
		errorMsg := formatValidationErrors(err)
		common.PlainTextError(w, http.StatusBadRequest, errorMsg)
		common.Error("Login validation error: %v", err)
		return
	}

	// If JSON was parsed instead, use that
	if loginReq.Email == "" && loginReq.Username == "" && loginReq.Password == "" {
		loginReq = jsonReq
	}

	// Validate the request (additional business logic)
	if err := validateLoginRequest(loginReq); err != nil {
		common.PlainTextError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Handle login with either email or username
	var session *Session
	var err error

	// Try to login with email first
	if loginReq.Email != "" {
		session, err = h.authProvider.Login(r.Context(), loginReq.Email, loginReq.Password)
	} else if loginReq.Username != "" {
		// If username is provided, convert to email
		user, userErr := h.authProvider.GetUserStore().GetUserByUsername(r.Context(), loginReq.Username)
		if userErr == nil {
			session, err = h.authProvider.Login(r.Context(), user.Email, loginReq.Password)
		} else {
			err = userErr
		}
	}

	// Handle login result
	if err != nil {
		common.PlainTextError(w, http.StatusUnauthorized, "Invalid credentials")
		common.Error("Login failed: %v", err)
		return
	}

	// Get user info
	user, err := h.authProvider.GetUserStore().GetUserByID(r.Context(), session.UserID)
	if err != nil {
		common.PlainTextError(w, http.StatusInternalServerError, "User not found")
		common.Error("User not found after login: %v", err)
		return
	}

	// Set auth cookie
	h.middleware.SetAuthCookie(w, session.Token)

	// Return success - convert to JSON for success response only
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Use direct encoding for the success response
	if err := json.NewEncoder(w).Encode(LoginResponse{
		Success: true,
		Token:   session.Token,
		User:    user,
	}); err != nil {
		common.Error("Failed to encode login response: %v", err)
	}
}

// HandleLoginForm handles form-based login from a web form
func (h *AuthHandlers) HandleLoginForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.PlainTextError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	// Parse and validate form data
	var loginReq LoginRequest
	var jsonReq LoginRequest
	if err := parseFormRequest(r, &loginReq, &jsonReq); err != nil {
		common.Error("Failed to parse login form: %v", err)
		errorMsg := formatValidationErrors(err)
		common.RedirectWithMessage(w, r, "/login", "", errorMsg)
		return
	}

	email := loginReq.Email
	password := loginReq.Password
	redirect := r.FormValue("redirect")

	if redirect == "" {
		redirect = "/"
	}

	// Try to login
	session, err := h.authProvider.Login(r.Context(), email, password)
	if err != nil {
		common.Error("Login failed: %v", err)
		common.RedirectWithMessage(w, r, "/login?email="+url.QueryEscape(email), "", "Invalid credentials")
		return
	}

	// Set auth cookie
	h.middleware.SetAuthCookie(w, session.Token)

	// Redirect back to the specified page
	http.Redirect(w, r, redirect, http.StatusFound)
}

// HandleRegister handles user registration
func (h *AuthHandlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.PlainTextError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check if registration is allowed
	if !h.config.AllowSignup {
		common.PlainTextError(w, http.StatusForbidden, "Registration is disabled")
		return
	}
	// Parse the registration request based on Content-Type
	var regReq RegisterRequest
	var jsonReq RegisterRequest
	if err := parseFormRequest(r, &regReq, &jsonReq); err != nil {
		errorMsg := formatValidationErrors(err)
		common.PlainTextError(w, http.StatusBadRequest, errorMsg)
		common.Error("Registration validation error: %v", err)
		return
	}

	// If JSON was parsed instead, use that
	if regReq.Email == "" && regReq.Username == "" && regReq.Password == "" {
		regReq = jsonReq
	}

	// Validate the registration request (additional business logic)
	if err := validateRegisterRequest(regReq); err != nil {
		common.PlainTextError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Register the user
	user, err := h.authProvider.Register(r.Context(), regReq.Email, regReq.Username, regReq.Password)
	if err != nil {
		common.PlainTextError(w, http.StatusBadRequest, "Registration failed")
		common.Error("Registration failed: %v", err)
		return
	}

	// Return success as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(RegisterResponse{
		Success: true,
		Message: "Registration successful",
		User:    user,
	}); err != nil {
		common.Error("Failed to encode registration response: %v", err)
	}
}

// HandleRegisterForm handles user registration from a web form
func (h *AuthHandlers) HandleRegisterForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.PlainTextError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check if registration is allowed
	if !h.config.AllowSignup {
		common.PlainTextError(w, http.StatusForbidden, "Registration is disabled")
		return
	}
	// Parse and validate form data
	var regReq RegisterRequest
	var jsonReq RegisterRequest
	if err := parseFormRequest(r, &regReq, &jsonReq); err != nil {
		common.Error("Failed to parse registration form: %v", err)

		// Prepare the base URL for potential error redirects
		redirectBase := "/register"
		redirectParams := "?"

		// Get submitted values even if validation failed
		email := r.FormValue("email")
		username := r.FormValue("username")

		if email != "" {
			redirectParams += "email=" + url.QueryEscape(email)
		}
		if username != "" {
			if len(redirectParams) > 1 {
				redirectParams += "&"
			}
			redirectParams += "username=" + url.QueryEscape(username)
		}
		if len(redirectParams) > 1 {
			redirectBase += redirectParams
		}

		errorMsg := formatValidationErrors(err)
		common.RedirectWithMessage(w, r, redirectBase, "", errorMsg)
		return
	}

	email := regReq.Email
	username := regReq.Username
	password := regReq.Password
	redirect := r.FormValue("redirect")

	if redirect == "" {
		redirect = "/login"
	}

	// Prepare the base URL for potential error redirects (in case of other errors)
	redirectBase := "/register"
	redirectParams := "?"
	if email != "" {
		redirectParams += "email=" + url.QueryEscape(email)
	}
	if username != "" {
		if len(redirectParams) > 1 {
			redirectParams += "&"
		}
		redirectParams += "username=" + url.QueryEscape(username)
	}
	if len(redirectParams) > 1 {
		redirectBase += redirectParams
	}

	// Register the user
	_, err := h.authProvider.Register(r.Context(), email, username, password)
	if err != nil {
		common.Error("Registration failed: %v", err)
		common.RedirectWithMessage(w, r, redirectBase, "", err.Error())
		return
	}

	// Redirect to login page with success message
	common.RedirectWithMessage(w, r, "/login?email="+url.QueryEscape(email), "Registration successful", "")
}

// HandleLogout handles user logout
func (h *AuthHandlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Look for the token in the cookie
	cookie, err := r.Cookie(h.config.CookieName)
	if err == nil {
		// If found, log out the session
		if err := h.authProvider.Logout(r.Context(), cookie.Value); err != nil {
			common.Error("Failed to logout session: %v", err)
			// Continue with logout despite the error
		}
	}

	// Clear the auth cookie
	h.middleware.ClearAuthCookie(w)

	// For web requests, redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleMe returns the current user's information
func (h *AuthHandlers) HandleMe(w http.ResponseWriter, r *http.Request) {
	// Get the user from the request context
	user, err := UserFromContext(r.Context())
	if err != nil {
		common.PlainTextError(w, http.StatusUnauthorized, "Unauthorized")
		common.Error("Auth failed in HandleMe: %v", err)
		return
	}

	// Don't return the password hash
	user.Password = ""

	// Return the user as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    user,
	}); err != nil {
		common.Error("Failed to encode user data: %v", err)
	}
}

// HandleRefreshToken refreshes the user's session token
func (h *AuthHandlers) HandleRefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get the session from the request context
	session, err := SessionFromContext(r.Context())
	if err != nil {
		common.PlainTextError(w, http.StatusUnauthorized, "Unauthorized")
		common.Error("Auth failed in HandleRefreshToken: %v", err)
		return
	}

	// Refresh the session
	newSession, err := h.authProvider.RefreshSession(r.Context(), session.Token)
	if err != nil {
		common.Error("Failed to refresh token: %v", err)
		common.PlainTextError(w, http.StatusInternalServerError, "Failed to refresh token")
		return
	}

	// Set the new auth cookie
	h.middleware.SetAuthCookie(w, newSession.Token)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Message: "Token refreshed",
		Data: map[string]interface{}{
			"token":      newSession.Token,
			"expires_at": newSession.ExpiresAt,
		},
	}); err != nil {
		common.Error("Failed to encode token refresh response: %v", err)
	}
}
