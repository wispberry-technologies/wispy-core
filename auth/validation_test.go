package auth

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestValidateLoginRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     LoginRequest
		wantErr bool
	}{
		{
			name: "Valid email login",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "Valid username login",
			req: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "Missing both email and username",
			req: LoginRequest{
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "Missing password",
			req: LoginRequest{
				Email: "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "Invalid email format",
			req: LoginRequest{
				Email:    "not-an-email",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "Password too short",
			req: LoginRequest{
				Email:    "test@example.com",
				Password: "short",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoginRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLoginRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRegisterRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     RegisterRequest
		wantErr bool
	}{
		{
			name: "Valid registration",
			req: RegisterRequest{
				Email:           "test@example.com",
				Username:        "testuser",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: false,
		},
		{
			name: "Missing email",
			req: RegisterRequest{
				Username:        "testuser",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: true,
		},
		{
			name: "Missing username",
			req: RegisterRequest{
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: true,
		},
		{
			name: "Missing password",
			req: RegisterRequest{
				Email:           "test@example.com",
				Username:        "testuser",
				ConfirmPassword: "password123",
			},
			wantErr: true,
		},
		{
			name: "Password mismatch",
			req: RegisterRequest{
				Email:           "test@example.com",
				Username:        "testuser",
				Password:        "password123",
				ConfirmPassword: "differentpassword",
			},
			wantErr: true,
		},
		{
			name: "Invalid email format",
			req: RegisterRequest{
				Email:           "not-an-email",
				Username:        "testuser",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: true,
		},
		{
			name: "Username too short",
			req: RegisterRequest{
				Email:           "test@example.com",
				Username:        "tu", // Too short
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegisterRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRegisterRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseFormRequest(t *testing.T) {
	// Test form data parsing
	formData := url.Values{}
	formData.Set("email", "form@example.com")
	formData.Set("username", "formuser")
	formData.Set("password", "formpass")

	req, _ := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var loginReq LoginRequest
	var jsonReq LoginRequest

	err := parseFormRequest(req, &loginReq, &jsonReq)
	if err != nil {
		t.Fatalf("Failed to parse form request: %v", err)
	}

	if loginReq.Email != "form@example.com" {
		t.Errorf("Expected email 'form@example.com', got %s", loginReq.Email)
	}
	if loginReq.Username != "formuser" {
		t.Errorf("Expected username 'formuser', got %s", loginReq.Username)
	}
	if loginReq.Password != "formpass" {
		t.Errorf("Expected password 'formpass', got %s", loginReq.Password)
	}
}

func TestFormatValidationErrors(t *testing.T) {
	// Create a request with validation errors
	loginReq := LoginRequest{
		Email:    "not-an-email",
		Password: "short",
	}

	err := validate.Struct(loginReq)
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}

	// Format the error
	errorMsg := formatValidationErrors(err)
	if errorMsg == "" {
		t.Error("Expected non-empty error message")
	}

	// Check that the message contains useful information
	if !strings.Contains(errorMsg, "email") && !strings.Contains(errorMsg, "password") {
		t.Errorf("Error message '%s' doesn't contain expected content", errorMsg)
	}
}
