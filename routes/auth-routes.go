package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/wispberry-technologies/wispy-core/common"
	"github.com/wispberry-technologies/wispy-core/handlers"
)

// SetupAuthRoutes sets up authentication-related API routes
func SetupAuthRoutes(r chi.Router, siteManager *common.SiteManager) {
	authHandler := handlers.NewAuthHandler(siteManager)

	// Public authentication routes
	r.Route("/auth", func(r chi.Router) {
		// Authentication endpoints
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)

		// Protected authentication endpoints
		r.Group(func(r chi.Router) {
			r.Use(authHandler.RequireAuth)
			r.Post("/logout", authHandler.Logout)
			r.Post("/logout-all", authHandler.LogoutAll)
			r.Post("/refresh", authHandler.RefreshSession)
			r.Get("/me", authHandler.GetCurrentUser)
			r.Put("/profile", authHandler.UpdateProfile)
			r.Post("/change-password", authHandler.ChangePassword)
		})
	})

	// User management routes (admin only)
	r.Route("/users", func(r chi.Router) {
		r.Use(authHandler.RequireRoles("admin"))
		r.Get("/", authHandler.GetUsers)
		r.Get("/{userID}", authHandler.GetUser)
		// Additional admin routes can be added here:
		// r.Put("/{userID}", authHandler.UpdateUser)
		// r.Delete("/{userID}", authHandler.DeleteUser)
		// r.Post("/{userID}/lock", authHandler.LockUser)
		// r.Post("/{userID}/unlock", authHandler.UnlockUser)
	})
}
