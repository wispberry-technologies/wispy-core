package auth

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// formatValidationErrors converts validator errors into user-friendly message
func formatValidationErrors(err error) string {
	if err == nil {
		return ""
	}

	// Try to extract validation errors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			// Map common validation errors to user-friendly messages
			switch e.Field() {
			case "Email":
				switch e.Tag() {
				case "required":
					return "Email address is required"
				case "email":
					return "Please enter a valid email address"
				}
			case "Username":
				switch e.Tag() {
				case "required":
					return "Username is required"
				case "min":
					return "Username must be at least 3 characters"
				case "max":
					return "Username cannot exceed 50 characters"
				}
			case "Password":
				switch e.Tag() {
				case "required":
					return "Password is required"
				case "min":
					return "Password must be at least 6 characters"
				}
			case "ConfirmPassword":
				switch e.Tag() {
				case "required":
					return "Please confirm your password"
				case "eqfield":
					return "Passwords do not match"
				}
			}

			// Default message for other validation errors
			return fmt.Sprintf("Invalid value for %s", e.Field())
		}
	}

	// If not validation error or no specific message was found
	return "Invalid input data"
}
