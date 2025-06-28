package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// validator instance
var validate = validator.New()

func init() {
	// Register validation for comparing password fields
	validate.RegisterStructValidation(passwordMatchValidation, RegisterRequest{})
}

// passwordMatchValidation validates that password and confirm password match
func passwordMatchValidation(sl validator.StructLevel) {
	req := sl.Current().Interface().(RegisterRequest)
	if req.Password != req.ConfirmPassword {
		sl.ReportError(req.ConfirmPassword, "ConfirmPassword", "confirm_password", "eqfield", "Password")
	}
}

// parseFormRequest parses and validates input from form data
// Per project standards, form data (application/x-www-form-urlencoded) MUST be used
// JSON is only allowed in specific cases where deeply nested data is needed
func parseFormRequest(r *http.Request, formResult interface{}, jsonResult interface{}) error {
	contentType := r.Header.Get("Content-Type")

	// Per project standards, we must use application/x-www-form-urlencoded
	if contentType == "" || contentType == "application/x-www-form-urlencoded" {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("failed to parse form data: %w", err)
		}

		// Map form values to the struct
		switch v := formResult.(type) {
		case *LoginRequest:
			v.Email = r.FormValue("email")
			v.Username = r.FormValue("username")
			v.Password = r.FormValue("password")

			// Validate the populated struct
			return validate.Struct(v)

		case *RegisterRequest:
			v.Email = r.FormValue("email")
			v.Username = r.FormValue("username")
			v.Password = r.FormValue("password")
			v.ConfirmPassword = r.FormValue("confirm_password")

			// Validate the populated struct
			return validate.Struct(v)

		default:
			return fmt.Errorf("unsupported form result type")
		}
	} else if contentType == "application/json" {
		// JSON is only allowed for specific cases with deeply nested structures
		// We'll check if this is coming from an API endpoint requiring JSON

		// Only accept JSON for specific endpoints that need complex data structures
		if !isDeeplyNestedDataEndpoint(r) {
			return fmt.Errorf("Content-Type application/json not allowed for this endpoint; use application/x-www-form-urlencoded")
		}

		if err := json.NewDecoder(r.Body).Decode(jsonResult); err != nil {
			return fmt.Errorf("failed to parse JSON data: %w", err)
		}

		// Validate the populated struct
		return validate.Struct(jsonResult)
	}

	return fmt.Errorf("unsupported Content-Type: %s", contentType)
}

// isDeeplyNestedDataEndpoint checks if the current request is for an endpoint
// that specifically requires JSON due to complex nested data structures
// Per project standards, this should be very limited
func isDeeplyNestedDataEndpoint(r *http.Request) bool {
	// List of endpoints that are allowed to accept JSON
	// This should be kept to an absolute minimum per project standards
	jsonAllowedPaths := []string{
		"/api/auth/complex-data", // Example endpoint needing complex data
	}

	for _, path := range jsonAllowedPaths {
		if r.URL.Path == path {
			return true
		}
	}

	return false
}

// validateLoginRequest validates the login request
func validateLoginRequest(req LoginRequest) error {
	// At least one of email or username must be provided
	if req.Email == "" && req.Username == "" {
		return fmt.Errorf("email or username is required")
	}

	// Validate password requirements
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	// Validate email format if provided
	if req.Email != "" {
		if err := validate.Var(req.Email, "email"); err != nil {
			return fmt.Errorf("invalid email format")
		}
	}

	// Validate password length
	if err := validate.Var(req.Password, "min=6"); err != nil {
		return fmt.Errorf("password must be at least 6 characters")
	}

	return nil
}

// validateRegisterRequest validates the register request
func validateRegisterRequest(req RegisterRequest) error {
	// Required fields
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	// Validate email
	if err := validate.Var(req.Email, "email"); err != nil {
		return fmt.Errorf("invalid email format")
	}

	// Validate username length
	if err := validate.Var(req.Username, "min=3,max=50"); err != nil {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}

	// Validate password
	if err := validate.Var(req.Password, "min=6"); err != nil {
		return fmt.Errorf("password must be at least 6 characters")
	}

	// Check password confirmation
	if req.Password != req.ConfirmPassword {
		return fmt.Errorf("passwords do not match")
	}

	return nil
}
