package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GoogleOAuthProvider implements the OAuthProvider interface for Google
type GoogleOAuthProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
	scopes       []string
}

// NewGoogleOAuthProvider creates a new Google OAuth provider
func NewGoogleOAuthProvider() *GoogleOAuthProvider {
	return &GoogleOAuthProvider{
		scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}
}

// Name implements OAuthProvider.Name
func (p *GoogleOAuthProvider) Name() string {
	return "google"
}

// DisplayName implements OAuthProvider.DisplayName
func (p *GoogleOAuthProvider) DisplayName() string {
	return "Google"
}

// GetAuthURL implements OAuthProvider.GetAuthURL
func (p *GoogleOAuthProvider) GetAuthURL(state string, redirectURI string) string {
	params := url.Values{}
	params.Add("client_id", p.clientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("state", state)
	params.Add("scope", strings.Join(p.scopes, " "))
	params.Add("access_type", "offline")
	params.Add("prompt", "consent")

	return "https://accounts.google.com/o/oauth2/auth?" + params.Encode()
}

// ExchangeCode implements OAuthProvider.ExchangeCode
func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code string, redirectURI string) (*OAuthToken, error) {
	params := url.Values{}
	params.Add("client_id", p.clientID)
	params.Add("client_secret", p.clientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", redirectURI)
	params.Add("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to exchange code: %s (%d)", body, resp.StatusCode)
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		IDToken      string `json:"id_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	token := &OAuthToken{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		TokenType:    tokenResponse.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
		Scope:        tokenResponse.Scope,
	}

	return token, nil
}

// GetUserInfo implements OAuthProvider.GetUserInfo
func (p *GoogleOAuthProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://www.googleapis.com/oauth2/v2/userinfo",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s (%d)", body, resp.StatusCode)
	}

	var googleUserInfo struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUserInfo); err != nil {
		return nil, err
	}

	// Convert to generic user info
	rawData := make(map[string]interface{})
	respBody, _ := json.Marshal(googleUserInfo)
	_ = json.Unmarshal(respBody, &rawData)

	userInfo := &OAuthUserInfo{
		ID:            googleUserInfo.ID,
		Email:         googleUserInfo.Email,
		VerifiedEmail: googleUserInfo.VerifiedEmail,
		Name:          googleUserInfo.Name,
		GivenName:     googleUserInfo.GivenName,
		FamilyName:    googleUserInfo.FamilyName,
		Picture:       googleUserInfo.Picture,
		Locale:        googleUserInfo.Locale,
		RawData:       rawData,
	}

	return userInfo, nil
}

// Configure implements OAuthProvider.Configure
func (p *GoogleOAuthProvider) Configure(config map[string]string) error {
	p.clientID = config["client_id"]
	p.clientSecret = config["client_secret"]

	if p.clientID == "" {
		return errors.New("client_id is required")
	}

	if p.clientSecret == "" {
		return errors.New("client_secret is required")
	}

	if scopes, ok := config["scopes"]; ok && scopes != "" {
		p.scopes = strings.Split(scopes, ",")
	}

	return nil
}

// DiscordOAuthProvider implements the OAuthProvider interface for Discord
type DiscordOAuthProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
	scopes       []string
}

// NewDiscordOAuthProvider creates a new Discord OAuth provider
func NewDiscordOAuthProvider() *DiscordOAuthProvider {
	return &DiscordOAuthProvider{
		scopes: []string{
			"identify",
			"email",
		},
	}
}

// Name implements OAuthProvider.Name
func (p *DiscordOAuthProvider) Name() string {
	return "discord"
}

// DisplayName implements OAuthProvider.DisplayName
func (p *DiscordOAuthProvider) DisplayName() string {
	return "Discord"
}

// GetAuthURL implements OAuthProvider.GetAuthURL
func (p *DiscordOAuthProvider) GetAuthURL(state string, redirectURI string) string {
	params := url.Values{}
	params.Add("client_id", p.clientID)
	params.Add("redirect_uri", redirectURI)
	params.Add("response_type", "code")
	params.Add("state", state)
	params.Add("scope", strings.Join(p.scopes, " "))

	return "https://discord.com/oauth2/authorize?" + params.Encode()
}

// ExchangeCode implements OAuthProvider.ExchangeCode
func (p *DiscordOAuthProvider) ExchangeCode(ctx context.Context, code string, redirectURI string) (*OAuthToken, error) {
	params := url.Values{}
	params.Add("client_id", p.clientID)
	params.Add("client_secret", p.clientSecret)
	params.Add("code", code)
	params.Add("redirect_uri", redirectURI)
	params.Add("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://discord.com/api/oauth2/token", strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to exchange code: %s (%d)", body, resp.StatusCode)
	}

	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return nil, err
	}

	token := &OAuthToken{
		AccessToken:  tokenResponse.AccessToken,
		RefreshToken: tokenResponse.RefreshToken,
		TokenType:    tokenResponse.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
		Scope:        tokenResponse.Scope,
	}

	return token, nil
}

// GetUserInfo implements OAuthProvider.GetUserInfo
func (p *DiscordOAuthProvider) GetUserInfo(ctx context.Context, token *OAuthToken) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		"https://discord.com/api/v10/users/@me",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("%s %s", token.TokenType, token.AccessToken))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get user info: %s (%d)", body, resp.StatusCode)
	}

	var discordUserInfo struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		Discriminator string `json:"discriminator"`
		Avatar        string `json:"avatar"`
		Email         string `json:"email"`
		Verified      bool   `json:"verified"`
		Locale        string `json:"locale"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&discordUserInfo); err != nil {
		return nil, err
	}

	// For Discord, the profile picture requires special formatting
	var picture string
	if discordUserInfo.Avatar != "" {
		if strings.HasPrefix(discordUserInfo.Avatar, "a_") {
			// Animated avatar (GIF)
			picture = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.gif", discordUserInfo.ID, discordUserInfo.Avatar)
		} else {
			// Static avatar (PNG)
			picture = fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", discordUserInfo.ID, discordUserInfo.Avatar)
		}
	}

	// Convert to generic user info
	rawData := make(map[string]interface{})
	respBody, _ := json.Marshal(discordUserInfo)
	_ = json.Unmarshal(respBody, &rawData)

	// Format the username to include the discriminator if available (for older Discord accounts)
	name := discordUserInfo.Username
	if discordUserInfo.Discriminator != "" && discordUserInfo.Discriminator != "0" {
		name = fmt.Sprintf("%s#%s", discordUserInfo.Username, discordUserInfo.Discriminator)
	}

	userInfo := &OAuthUserInfo{
		ID:            discordUserInfo.ID,
		Email:         discordUserInfo.Email,
		VerifiedEmail: discordUserInfo.Verified,
		Name:          name,
		Picture:       picture,
		Locale:        discordUserInfo.Locale,
		RawData:       rawData,
	}

	return userInfo, nil
}

// Configure implements OAuthProvider.Configure
func (p *DiscordOAuthProvider) Configure(config map[string]string) error {
	p.clientID = config["client_id"]
	p.clientSecret = config["client_secret"]

	if p.clientID == "" {
		return errors.New("client_id is required")
	}

	if p.clientSecret == "" {
		return errors.New("client_secret is required")
	}

	if scopes, ok := config["scopes"]; ok && scopes != "" {
		p.scopes = strings.Split(scopes, ",")
	}

	return nil
}
