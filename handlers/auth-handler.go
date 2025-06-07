package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wispberry-technologies/wispy-core/auth"
	"github.com/wispberry-technologies/wispy-core/common"
	"github.com/wispberry-technologies/wispy-core/models"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	siteManager *common.SiteManager
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(siteManager *common.SiteManager) *AuthHandler {
	return &AuthHandler{
		siteManager: siteManager,
	}
}

// getDomainFromRequest extracts the domain from the request
func (ah *AuthHandler) getDomainFromRequest(r *http.Request) string {
	// Try to get domain from Host header
	host := r.Host
	if host == "" {
		// Fallback to localhost for development
		return "localhost"
	}
	return host
}

// Login handles user login requests
func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		ah.writeErrorResponse(w, "Email and password are required", http.StatusBadRequest, nil)
		return
	}

	// Get domain and client info
	domain := ah.getDomainFromRequest(r)
	ipAddress := auth.GetClientIP(r)
	userAgent := auth.GetUserAgent(r)

	// Attempt login
	user, session, err := ah.siteManager.Login(domain, req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		ah.writeErrorResponse(w, "Invalid credentials", http.StatusUnauthorized, err)
		return
	}

	// Set session cookie - we need to create a session manager to set the cookie
	db, err := ah.siteManager.GetDB(domain, "users")
	if err != nil {
		ah.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	config := auth.NewAuthConfig()
	sessionManager := auth.NewSessionManager(db, config)
	sessionManager.SetSessionCookie(w, session.Token)

	// Return success response
	response := models.AuthResponse{
		Success: true,
		Message: "Login successful",
		User: &models.UserInfo{
			ID:          user.ID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}

	ah.writeJSONResponse(w, response, http.StatusOK)
}

// Register handles user registration requests
func (ah *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		ah.writeErrorResponse(w, "Email and password are required", http.StatusBadRequest, nil)
		return
	}

	// Get domain
	domain := ah.getDomainFromRequest(r)

	// Attempt registration
	user, err := ah.siteManager.Register(domain, req.Email, req.Password, req.FirstName, req.LastName, req.DisplayName)
	if err != nil {
		ah.writeErrorResponse(w, "Registration failed", http.StatusBadRequest, err)
		return
	}

	// Return success response
	response := models.AuthResponse{
		Success: true,
		Message: "Registration successful",
		User: &models.UserInfo{
			ID:          user.ID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}

	ah.writeJSONResponse(w, response, http.StatusCreated)
}

// Logout handles user logout requests
func (ah *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	domain := ah.getDomainFromRequest(r)

	session, err := ah.siteManager.GetSessionFromRequest(domain, r)
	if err != nil {
		ah.writeErrorResponse(w, "No valid session found", http.StatusUnauthorized, err)
		return
	}

	// Create session manager to handle logout
	db, err := ah.siteManager.GetDB(domain, "users")
	if err != nil {
		ah.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	config := auth.NewAuthConfig()
	sessionManager := auth.NewSessionManager(db, config)

	if err := sessionManager.DeleteSession(session.ID); err != nil {
		ah.writeErrorResponse(w, "Failed to logout", http.StatusInternalServerError, err)
		return
	}

	sessionManager.ClearSessionCookie(w)

	response := models.AuthResponse{
		Success: true,
		Message: "Logout successful",
	}

	ah.writeJSONResponse(w, response, http.StatusOK)
}

// LogoutAll handles logging out from all sessions
func (ah *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	domain := ah.getDomainFromRequest(r)

	session, err := ah.siteManager.GetSessionFromRequest(domain, r)
	if err != nil {
		ah.writeErrorResponse(w, "No valid session found", http.StatusUnauthorized, err)
		return
	}

	// Get user from session
	user, _, err := ah.siteManager.ValidateSession(domain, session.Token)
	if err != nil {
		ah.writeErrorResponse(w, "Invalid session", http.StatusUnauthorized, err)
		return
	}

	// Create session manager to handle logout all
	db, err := ah.siteManager.GetDB(domain, "users")
	if err != nil {
		ah.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	config := auth.NewAuthConfig()
	sessionManager := auth.NewSessionManager(db, config)

	if err := sessionManager.DeleteAllUserSessions(user.ID); err != nil {
		ah.writeErrorResponse(w, "Failed to logout from all sessions", http.StatusInternalServerError, err)
		return
	}

	sessionManager.ClearSessionCookie(w)

	response := models.AuthResponse{
		Success: true,
		Message: "Logged out from all sessions",
	}

	ah.writeJSONResponse(w, response, http.StatusOK)
}

// GetCurrentUser returns the current authenticated user
func (ah *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		ah.writeErrorResponse(w, "User not found in context", http.StatusInternalServerError, nil)
		return
	}

	userInfo := &models.UserInfo{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	ah.writeJSONResponse(w, userInfo, http.StatusOK)
}

// UpdateProfile handles updating user profile
func (ah *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		ah.writeErrorResponse(w, "User not found in context", http.StatusInternalServerError, nil)
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	// Update user fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}

	// Get database and update user
	domain := ah.getDomainFromRequest(r)
	db, err := ah.siteManager.GetDB(domain, "users")
	if err != nil {
		ah.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	repository := auth.NewUserRepository(db)
	if err := repository.UpdateUser(user); err != nil {
		ah.writeErrorResponse(w, "Failed to update profile", http.StatusInternalServerError, err)
		return
	}

	response := models.AuthResponse{
		Success: true,
		Message: "Profile updated successfully",
		User: &models.UserInfo{
			ID:          user.ID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
	}

	ah.writeJSONResponse(w, response, http.StatusOK)
}

// ChangePassword handles password change requests
func (ah *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		ah.writeErrorResponse(w, "User not found in context", http.StatusInternalServerError, nil)
		return
	}

	var req models.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ah.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest, err)
		return
	}

	// Basic validation
	if req.CurrentPassword == "" || req.NewPassword == "" {
		ah.writeErrorResponse(w, "Current and new passwords are required", http.StatusBadRequest, nil)
		return
	}

	// Get database and change password
	domain := ah.getDomainFromRequest(r)
	db, err := ah.siteManager.GetDB(domain, "users")
	if err != nil {
		ah.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	repository := auth.NewUserRepository(db)

	// Verify current password
	if err := auth.VerifyPassword(req.CurrentPassword, user.PasswordHash); err != nil {
		ah.writeErrorResponse(w, "Current password is incorrect", http.StatusBadRequest, err)
		return
	}

	// Validate new password
	if !auth.IsValidPassword(req.NewPassword) {
		ah.writeErrorResponse(w, "New password does not meet requirements", http.StatusBadRequest, nil)
		return
	}

	// Hash new password
	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		ah.writeErrorResponse(w, "Failed to hash new password", http.StatusInternalServerError, err)
		return
	}

	// Update password
	if err := repository.UpdatePassword(user.ID, newPasswordHash); err != nil {
		ah.writeErrorResponse(w, "Failed to update password", http.StatusInternalServerError, err)
		return
	}

	// Clear all sessions to force re-login
	config := auth.NewAuthConfig()
	sessionManager := auth.NewSessionManager(db, config)
	sessionManager.DeleteAllUserSessions(user.ID)
	sessionManager.ClearSessionCookie(w)

	response := models.AuthResponse{
		Success: true,
		Message: "Password changed successfully. Please login again.",
	}

	ah.writeJSONResponse(w, response, http.StatusOK)
}

// RefreshSession extends the current session
func (ah *AuthHandler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	session, ok := auth.GetSessionFromContext(r.Context())
	if !ok {
		ah.writeErrorResponse(w, "Session not found in context", http.StatusInternalServerError, nil)
		return
	}

	domain := ah.getDomainFromRequest(r)
	db, err := ah.siteManager.GetDB(domain, "users")
	if err != nil {
		ah.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	config := auth.NewAuthConfig()
	sessionManager := auth.NewSessionManager(db, config)

	if err := sessionManager.RefreshSession(session.ID); err != nil {
		ah.writeErrorResponse(w, "Failed to refresh session", http.StatusInternalServerError, err)
		return
	}

	response := models.AuthResponse{
		Success: true,
		Message: "Session refreshed successfully",
	}

	ah.writeJSONResponse(w, response, http.StatusOK)
}

// GetUsers handles getting a list of users (admin only)
func (ah *AuthHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	domain := ah.getDomainFromRequest(r)
	db, err := ah.siteManager.GetDB(domain, "users")
	if err != nil {
		ah.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	repository := auth.NewUserRepository(db)
	users, err := repository.ListUsers(limit, offset)
	if err != nil {
		ah.writeErrorResponse(w, "Failed to retrieve users", http.StatusInternalServerError, err)
		return
	}

	// Convert to user info objects
	userInfos := make([]*models.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = &models.UserInfo{
			ID:          user.ID,
			Email:       user.Email,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		}
	}

	ah.writeJSONResponse(w, userInfos, http.StatusOK)
}

// GetUser handles getting a specific user by ID (admin only)
func (ah *AuthHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		ah.writeErrorResponse(w, "User ID is required", http.StatusBadRequest, nil)
		return
	}

	domain := ah.getDomainFromRequest(r)
	db, err := ah.siteManager.GetDB(domain, "users")
	if err != nil {
		ah.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError, err)
		return
	}

	repository := auth.NewUserRepository(db)
	user, err := repository.GetUserByID(userID)
	if err != nil {
		ah.writeErrorResponse(w, "User not found", http.StatusNotFound, err)
		return
	}

	userInfo := &models.UserInfo{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	ah.writeJSONResponse(w, userInfo, http.StatusOK)
}

// RequireAuth middleware that requires authentication for a site
func (ah *AuthHandler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domain := ah.getDomainFromRequest(r)

		session, err := ah.siteManager.GetSessionFromRequest(domain, r)
		if err != nil {
			ah.writeErrorResponse(w, "Unauthorized", http.StatusUnauthorized, err)
			return
		}

		user, _, err := ah.siteManager.ValidateSession(domain, session.Token)
		if err != nil {
			// Clear invalid session cookie
			db, _ := ah.siteManager.GetDB(domain, "users")
			config := auth.NewAuthConfig()
			sessionManager := auth.NewSessionManager(db, config)
			sessionManager.ClearSessionCookie(w)
			ah.writeErrorResponse(w, "Unauthorized", http.StatusUnauthorized, err)
			return
		}

		// Add user and session to request context
		ctx := auth.SetUserInContext(r.Context(), user)
		ctx = auth.SetSessionInContext(ctx, session)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRoles middleware that requires specific roles for a site
func (ah *AuthHandler) RequireRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return ah.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := auth.GetUserFromContext(r.Context())
			if !ok {
				ah.writeErrorResponse(w, "Unauthorized", http.StatusUnauthorized, nil)
				return
			}

			// Check if user has any of the required roles
			userRoles := user.GetRoles()
			hasRole := false
			for _, requiredRole := range roles {
				for _, userRole := range userRoles {
					if userRole == requiredRole {
						hasRole = true
						break
					}
				}
				if hasRole {
					break
				}
			}

			if !hasRole {
				ah.writeErrorResponse(w, "Forbidden", http.StatusForbidden, nil)
				return
			}

			next.ServeHTTP(w, r)
		}))
	}
}

// Helper functions

// writeJSONResponse writes a JSON response
func (ah *AuthHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response
func (ah *AuthHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int, err error) {
	// Log the error for debugging
	if err != nil {
		// You might want to add proper logging here
		w.Header().Set("X-Debug", err.Error())
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	w.Write([]byte(message))
}
