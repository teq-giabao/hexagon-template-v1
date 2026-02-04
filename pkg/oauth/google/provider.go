package google

import (
	"context"
	"encoding/json"
	"errors"
	"hexagon/auth"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

type Provider struct {
	config *oauth2.Config
}

func NewProvider(clientID, clientSecret, redirectURL string) *Provider {
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" || strings.TrimSpace(redirectURL) == "" {
		return nil
	}
	return &Provider{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"openid",
				"email",
				"profile",
			},
		},
	}
}

func (p *Provider) AuthCodeURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *Provider) Exchange(ctx context.Context, code string) (auth.OAuthUser, error) {
	if p == nil || p.config == nil {
		return auth.OAuthUser{}, errors.New("google oauth not configured")
	}

	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return auth.OAuthUser{}, err
	}

	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return auth.OAuthUser{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return auth.OAuthUser{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return auth.OAuthUser{}, errors.New("failed to fetch user info")
	}

	var payload struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		VerifiedEmail bool   `json:"verified_email"` // nolint: tagliatelle
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return auth.OAuthUser{}, err
	}

	return auth.OAuthUser{
		Email:         payload.Email,
		Name:          payload.Name,
		EmailVerified: payload.VerifiedEmail,
	}, nil
}
