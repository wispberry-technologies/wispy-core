package api_v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"wispy-core/internal/cache"
	"wispy-core/pkg/auth"
	"wispy-core/pkg/common"
	"wispy-core/pkg/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

func AuthRouter(r chi.Router) {
	// Auth routes
	// ----------------------------------------------------------------
	r.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		// Check if the site has registration enabled
		siteInstance, ok := r.Context().Value(auth.SiteInstanceContextKey).(*models.SiteInstance)
		if !ok || siteInstance == nil {
			common.PlainTextError(w, http.StatusInternalServerError, "Site context missing")
			return
		}

		// Check if registration is enabled for this site
		if siteInstance.AuthConfig != nil && !siteInstance.AuthConfig.RegistrationEnabled {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error": map[string]interface{}{
					"message": "Registration is not enabled for this site",
				},
			})
			return
		}

		// Parse request data from JSON or form
		type Req struct {
			Email       string `json:"email" form:"email" validate:"required,email"`
			Password    string `json:"password" form:"password" validate:"required,min=8"`
			FirstName   string `json:"first_name" form:"first_name" validate:"required"`
			LastName    string `json:"last_name" form:"last_name" validate:"required"`
			DisplayName string `json:"display_name" form:"display_name"`
		}
		var req Req

		contentType := r.Header.Get("Content-Type")
		if contentType == "application/json" {
			// Parse JSON request
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&req); err != nil {
				common.PlainTextError(w, http.StatusBadRequest, "Invalid JSON data", err.Error())
				return
			}
		} else {
			// Parse form data
			if err := r.ParseForm(); err != nil {
				common.PlainTextError(w, http.StatusBadRequest, "Invalid form data", err.Error())
				return
			}

			req = Req{
				Email:       r.FormValue("email"),
				Password:    r.FormValue("password"),
				FirstName:   r.FormValue("first_name"),
				LastName:    r.FormValue("last_name"),
				DisplayName: r.FormValue("display_name"),
			}
		}

		// Check required fields based on site configuration
		if siteInstance.AuthConfig != nil && len(siteInstance.AuthConfig.RequiredFields) > 0 {
			missingFields := []string{}

			// Custom validation based on site configuration
			for _, field := range siteInstance.AuthConfig.RequiredFields {
				switch field {
				case "email":
					if req.Email == "" {
						missingFields = append(missingFields, "email")
					}
				case "password":
					if req.Password == "" {
						missingFields = append(missingFields, "password")
					}
				case "first_name":
					if req.FirstName == "" {
						missingFields = append(missingFields, "first_name")
					}
				case "last_name":
					if req.LastName == "" {
						missingFields = append(missingFields, "last_name")
					}
				case "display_name":
					if req.DisplayName == "" {
						missingFields = append(missingFields, "display_name")
					}
				}
			}

			if len(missingFields) > 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"message": "Required fields missing: " + fmt.Sprintf("%v", missingFields),
						"fields":  missingFields,
					},
				})
				return
			}
		} else {
			// Default validation if no specific fields are configured
			validate := validator.New()
			if err := validate.Struct(req); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"error": map[string]interface{}{
						"message": "Validation failed: " + err.Error(),
					},
				})
				return
			}
		}

		db, err := cache.GetConnection(siteInstance.DBCache, siteInstance.Domain, "users")
		if err != nil {
			common.PlainTextError(w, http.StatusInternalServerError, "DB error", err.Error())
			return
		}

		// Get default roles if configured
		defaultRoles := []string{}
		if siteInstance.AuthConfig != nil {
			defaultRoles = siteInstance.AuthConfig.DefaultRoles
		}

		// Register the user with default roles
		user, err := auth.Register(db, siteInstance.Domain, req.Email, req.Password, req.FirstName, req.LastName, req.DisplayName)
		if err != nil {
			// Return JSON error response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error": map[string]interface{}{
					"message": err.Error(),
				},
			})
			return
		}

		// Add default roles if any
		if len(defaultRoles) > 0 {
			userDriver := auth.NewUserSqlDriver(db)
			for _, role := range defaultRoles {
				if err := userDriver.AddRoleToUser(user.ID, role); err != nil {
					common.Error("Failed to add default role %s to user %s: %v", role, user.ID, err)
				}
			}
		}

		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Registration successful",
			"user": map[string]interface{}{
				"id":    user.ID,
				"email": user.Email,
			},
		})
	})
	//  Login route
	// ----------------------------------------------------------------
	r.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			common.PlainTextError(w, http.StatusBadRequest, "Invalid form data", err.Error())
			return
		}
		type Req struct {
			Email    string `validate:"required,email"`
			Password string `validate:"required"`
		}
		req := Req{
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			common.PlainTextError(w, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}
		siteInstance, ok := r.Context().Value(auth.SiteInstanceContextKey).(*models.SiteInstance)
		if !ok || siteInstance == nil {
			common.PlainTextError(w, http.StatusInternalServerError, "Site context missing")
			return
		}
		db, err := cache.GetConnection(siteInstance.DBCache, siteInstance.Domain, "users")
		if err != nil {
			common.PlainTextError(w, http.StatusInternalServerError, "DB error", err.Error())
			return
		}
		maxAttempts := siteInstance.AuthConfig.MaxFailedLoginAttempts
		lockDuration := siteInstance.AuthConfig.FailedLoginAttemptLockDuration
		var session *auth.Session
		_, session, err = auth.Login(db, siteInstance.Domain, req.Email, req.Password, r.RemoteAddr, r.UserAgent(), maxAttempts, lockDuration)
		if err != nil {
			common.PlainTextError(w, http.StatusUnauthorized, err.Error())
			return
		}
		sessionDriver := auth.NewSessionSqlDriver(db)
		sessionDriver.SetSessionCookie(w, session.Token)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Logged in"))
	})
	//  Logout route
	r.Post("/logout", func(w http.ResponseWriter, r *http.Request) {
		siteInstance, ok := r.Context().Value(auth.SiteInstanceContextKey).(*models.SiteInstance)
		if !ok || siteInstance == nil {
			common.PlainTextError(w, http.StatusInternalServerError, "Site context missing")
			return
		}
		db, err := cache.GetConnection(siteInstance.DBCache, siteInstance.Domain, "users")
		if err != nil {
			common.PlainTextError(w, http.StatusInternalServerError, "DB error", err.Error())
			return
		}
		sessionDriver := auth.NewSessionSqlDriver(db)
		sessionDriver.ClearSessionCookie(w, r)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Logged out"))
	})
	r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
		session, _ := r.Context().Value(auth.SessionContextKey).(*auth.Session)
		site, _ := r.Context().Value(auth.SiteInstanceContextKey).(*models.SiteInstance)
		fmt.Println("Session:", session, "Site:", site)

		user, ok := r.Context().Value(auth.UserContextKey).(*auth.User)
		if !ok || user == nil {
			common.PlainTextError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(user.Email))
	})
	//  Refresh
	// ----------------------------------------------------------------
	r.Post("/refresh", func(w http.ResponseWriter, r *http.Request) {
		siteInstance, ok := r.Context().Value(auth.SiteInstanceContextKey).(*models.SiteInstance)
		if !ok || siteInstance == nil {
			common.PlainTextError(w, http.StatusInternalServerError, "Site context missing")
			return
		}
		db, err := cache.GetConnection(siteInstance.DBCache, siteInstance.Domain, "users")
		if err != nil {
			common.PlainTextError(w, http.StatusInternalServerError, "DB error", err.Error())
			return
		}
		session, ok := r.Context().Value(auth.SessionContextKey).(*auth.Session)
		if !ok || session == nil {
			common.PlainTextError(w, http.StatusUnauthorized, "No session")
			return
		}
		sessionDriver := auth.NewSessionSqlDriver(db)
		err = sessionDriver.RefreshSession(session.ID)
		if err != nil {
			common.PlainTextError(w, http.StatusInternalServerError, "Failed to refresh session", err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Session refreshed"))
	})
	//  Change password
	// ----------------------------------------------------------------
	r.Post("/change-password", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			common.PlainTextError(w, http.StatusBadRequest, "Invalid form data", err.Error())
			return
		}
		type Req struct {
			CurrentPassword string `validate:"required"`
			NewPassword     string `validate:"required,min=8"`
		}
		req := Req{
			CurrentPassword: r.FormValue("current_password"),
			NewPassword:     r.FormValue("new_password"),
		}
		validate := validator.New()
		if err := validate.Struct(req); err != nil {
			common.PlainTextError(w, http.StatusBadRequest, "Validation failed", err.Error())
			return
		}
		user, ok := r.Context().Value(auth.UserContextKey).(*auth.User)
		if !ok || user == nil {
			common.PlainTextError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		siteInstance, ok := r.Context().Value(auth.SiteInstanceContextKey).(*models.SiteInstance)
		if !ok || siteInstance == nil {
			common.PlainTextError(w, http.StatusInternalServerError, "Site context missing")
			return
		}
		db, err := cache.GetConnection(siteInstance.DBCache, siteInstance.Domain, "users")
		if err != nil {
			common.PlainTextError(w, http.StatusInternalServerError, "DB error", err.Error())
			return
		}
		// Validate current password
		if err := auth.VerifyPassword(req.CurrentPassword, user.PasswordHash); err != nil {
			common.PlainTextError(w, http.StatusUnauthorized, "Current password incorrect")
			return
		}
		// Hash new password
		hash, err := auth.HashPassword(req.NewPassword)
		if err != nil {
			common.PlainTextError(w, http.StatusInternalServerError, "Failed to hash password", err.Error())
			return
		}
		userDriver := auth.NewUserSqlDriver(db)
		err = userDriver.UpdatePassword(user.ID, hash)
		if err != nil {
			common.PlainTextError(w, http.StatusInternalServerError, "Failed to update password", err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Password changed"))
	})

	// OAuth Provider (Discord only)
	r.Get("/oauth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		// provider := chi.URLParam(r, "provider")
		conf, _, err := auth.GetDiscordOAuthConfig(r)
		if err != nil {
			common.PlainTextError(w, http.StatusBadRequest, err.Error())
			return
		}
		// Generate state and store in secure cookie
		state := auth.GenerateRandomState()
		http.SetCookie(w, &http.Cookie{
			Name:     "discord_oauth_state",
			Value:    state,
			Path:     "/",
			HttpOnly: common.IsProduction(),
			Secure:   common.IsProduction(),
			MaxAge:   300, // 5 minutes
			SameSite: http.SameSiteLaxMode,
		})
		params := "?response_type=code&client_id=" + conf.ClientID +
			"&scope=identify%20email" +
			"&redirect_uri=" + url.QueryEscape(conf.RedirectURI) +
			"&state=" + url.QueryEscape(state)
		http.Redirect(w, r, "https://discord.com/oauth2/authorize"+params, http.StatusFound)
	})

	// OAuth Callback (Discord only)
	r.Get("/oauth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		// provider := chi.URLParam(r, "provider")
		conf, siteInstance, err := auth.GetDiscordOAuthConfig(r)
		if err != nil {
			common.PlainTextError(w, http.StatusBadRequest, err.Error())
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			common.PlainTextError(w, http.StatusBadRequest, "Missing code from Discord")
			return
		}
		// Validate state from cookie
		stateParam := r.URL.Query().Get("state")
		stateCookie, err := r.Cookie("discord_oauth_state")
		if err != nil || stateCookie.Value == "" || stateParam == "" || stateCookie.Value != stateParam {
			common.PlainTextError(w, http.StatusBadRequest, "Invalid or missing OAuth state")
			return
		}
		// Clear the state cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "discord_oauth_state",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: common.IsProduction(),
			Secure:   common.IsProduction(),
			SameSite: http.SameSiteLaxMode,
		})
		// Exchange code for access token
		tokenResp, err := auth.ExchangeDiscordCodeForToken(conf, code)
		if err != nil {
			common.PlainTextError(w, http.StatusBadGateway, "Failed to get Discord token", err.Error())
			return
		}
		discordUser, err := auth.FetchDiscordUser(tokenResp.AccessToken)
		if err != nil {
			common.PlainTextError(w, http.StatusBadGateway, "Failed to fetch Discord user", err.Error())
			return
		}
		db, err := cache.GetConnection(siteInstance.DBCache, siteInstance.Domain, "users")
		if err != nil {
			common.PlainTextError(w, http.StatusInternalServerError, "DB error", err.Error())
			return
		}
		oauthDriver := auth.NewOAuthSqlDriver(db)
		userDriver := auth.NewUserSqlDriver(db)
		oauthAcc, err := oauthDriver.GetOAuthAccount("discord", discordUser.ID)
		var user *auth.User
		if err == nil && oauthAcc != nil {
			// Existing OAuth account, get user
			user, err = userDriver.GetUserByID(oauthAcc.UserID)
			if err != nil || user == nil {
				common.PlainTextError(w, http.StatusUnauthorized, "User not found for OAuth account")
				return
			}
			// Update tokens
			oauthAcc.AccessToken = tokenResp.AccessToken
			oauthAcc.RefreshToken = tokenResp.RefreshToken
			oauthAcc.ExpiresAt = tokenResp.ExpiresAt
			oauthAcc.DisplayName = discordUser.Username
			oauthAcc.Avatar = discordUser.AvatarURL()
			oauthAcc.Email = discordUser.Email
			err = oauthDriver.UpdateOAuthAccount(oauthAcc)
			if err != nil {
				common.PlainTextError(w, http.StatusInternalServerError, "Failed to update OAuth account", err.Error())
				return
			}
		} else {
			// New OAuth account, create user and account
			user, err = userDriver.GetUserByEmail(discordUser.Email)
			if err != nil || user == nil {
				// Create new user
				user = auth.NewUser(discordUser.Email, "", "", discordUser.Username)
				user.Avatar = discordUser.AvatarURL()
				err = userDriver.CreateUser(user)
				if err != nil {
					common.PlainTextError(w, http.StatusInternalServerError, "Failed to create user", err.Error())
					return
				}
			}
			oauthAcc = auth.NewOAuthAccount(user.ID, "discord", discordUser.ID, discordUser.Email, discordUser.Username, discordUser.AvatarURL())
			oauthAcc.AccessToken = tokenResp.AccessToken
			oauthAcc.RefreshToken = tokenResp.RefreshToken
			oauthAcc.ExpiresAt = tokenResp.ExpiresAt
			err = oauthDriver.CreateOAuthAccount(oauthAcc)
			if err != nil {
				common.PlainTextError(w, http.StatusInternalServerError, "Failed to create OAuth account", err.Error())
				return
			}
		}
		// Create session
		sessionDriver := auth.NewSessionSqlDriver(db)
		session, err := sessionDriver.CreateSession(user.ID, r.RemoteAddr, r.UserAgent())
		if err != nil {
			common.PlainTextError(w, http.StatusInternalServerError, "Failed to create session", err.Error())
			return
		}
		sessionDriver.SetSessionCookie(w, session.Token)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("Logged in with Discord as " + user.DisplayName))
	})

}
