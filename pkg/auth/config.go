package auth

// Config represents authentication configuration
type Config struct {
	Enabled   bool
	PublicURL string
	LoginURL  string
	LogoutURL string
	Required  bool
}
