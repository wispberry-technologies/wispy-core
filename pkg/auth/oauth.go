package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
	"wispy-core/pkg/common"
	"wispy-core/pkg/models"

	"github.com/go-chi/chi/v5"
)

// GenerateRandomState returns a random string for OAuth state
func GenerateRandomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

type DiscordTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type DiscordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

func (u *DiscordUser) AvatarURL() string {
	if u.Avatar == "" {
		return ""
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", u.ID, u.Avatar)
}

type DiscordTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    *time.Time
	TokenType    string
}

// ExchangeDiscordCodeForToken exchanges code for Discord access token
func ExchangeDiscordCodeForToken(conf models.OAuth, code string) (*DiscordTokenResult, error) {
	data := url.Values{}
	data.Set("client_id", conf.ClientID)
	data.Set("client_secret", conf.ClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", conf.RedirectURI)
	data.Set("scope", "identify email")

	req, err := http.NewRequest("POST", "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord token error: %s", string(body))
	}
	var token DiscordTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &DiscordTokenResult{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    &expiresAt,
		TokenType:    token.TokenType,
	}, nil
}

// FetchDiscordUser fetches Discord user info
func FetchDiscordUser(accessToken string) (*DiscordUser, error) {
	req, err := http.NewRequest("GET", "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discord user error: %s", string(body))
	}
	var user DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// OAuthSqlDriver handles oauth_accounts table
// (Stub: implement DB logic as needed)
type OAuthSqlDriver struct {
	db *sql.DB
}

func NewOAuthSqlDriver(db *sql.DB) *OAuthSqlDriver {
	return &OAuthSqlDriver{db: db}
}

func (d *OAuthSqlDriver) GetOAuthAccount(provider, providerID string) (*OAuthAccount, error) {
	row := d.db.QueryRow(GetOAuthAccountSQL, provider, providerID)
	var acc OAuthAccount
	var expiresAt sql.NullTime
	if err := row.Scan(&acc.ID, &acc.UserID, &acc.Provider, &acc.ProviderID, &acc.Email, &acc.DisplayName, &acc.Avatar, &acc.AccessToken, &acc.RefreshToken, &expiresAt, &acc.CreatedAt, &acc.UpdatedAt); err != nil {
		return nil, err
	}
	if expiresAt.Valid {
		acc.ExpiresAt = &expiresAt.Time
	}
	return &acc, nil
}

func (d *OAuthSqlDriver) UpdateOAuthAccount(acc *OAuthAccount) error {
	_, err := d.db.Exec(UpdateOAuthAccountSQL, acc.Email, acc.DisplayName, acc.Avatar, acc.AccessToken, acc.RefreshToken, acc.ExpiresAt, acc.UpdatedAt, acc.ID)
	return err
}

func (d *OAuthSqlDriver) CreateOAuthAccount(acc *OAuthAccount) error {
	_, err := d.db.Exec(InsertOAuthAccountSQL, acc.ID, acc.UserID, acc.Provider, acc.ProviderID, acc.Email, acc.DisplayName, acc.Avatar, acc.AccessToken, acc.RefreshToken, acc.ExpiresAt, acc.CreatedAt, acc.UpdatedAt)
	return err
}

// Helper: get Discord OAuth config and site instance, with error handling
func GetDiscordOAuthConfig(r *http.Request) (conf models.OAuth, siteInstance *models.SiteInstance, err error) {
	provider := chi.URLParam(r, "provider")
	siteInstance, ok := r.Context().Value(SiteInstanceContextKey).(*models.SiteInstance)
	if !ok || siteInstance == nil {
		return conf, nil, fmt.Errorf("site context missing")
	}
	// Check if provider is allowed for this site
	allowed := false
	allowed = slices.Contains(siteInstance.Config.OAuthProviders, provider)
	if !allowed {
		return conf, siteInstance, fmt.Errorf(provider + " OAuth not allowed for this site")
	}
	// Get credentials from environment
	switch provider {
	case "discord":
		conf.ClientID = common.MustGetEnv("DISCORD_CLIENT_ID")
		conf.ClientSecret = common.MustGetEnv("DISCORD_CLIENT_SECRET")
		conf.RedirectURI = common.MustGetEnv("DISCORD_REDIRECT_URI")
		conf.Enabled = true
		return conf, siteInstance, nil
	default:
		return conf, siteInstance, fmt.Errorf("unsupported OAuth provider: %s", provider)
	}
}
