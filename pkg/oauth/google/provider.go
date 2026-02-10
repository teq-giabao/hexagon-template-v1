package google

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"hexagon/auth"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const userInfoEndpoint = "https://www.googleapis.com/oauth2/v2/userinfo"

type Provider struct {
	config *oauth2.Config
}

type flexibleBool bool

func (b *flexibleBool) UnmarshalJSON(data []byte) error {
	s := strings.TrimSpace(string(data))
	if s == "" || s == "null" {
		*b = false
		return nil
	}
	if s == "true" || s == "false" {
		var v bool
		if err := json.Unmarshal(data, &v); err != nil {
			return err
		}
		*b = flexibleBool(v)
		return nil
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		unquoted, err := strconv.Unquote(s)
		if err != nil {
			return err
		}
		v, err := strconv.ParseBool(strings.TrimSpace(unquoted))
		if err != nil {
			return err
		}
		*b = flexibleBool(v)
		return nil
	}

	var num json.Number
	if err := json.Unmarshal(data, &num); err == nil {
		if i, err := num.Int64(); err == nil {
			*b = i != 0
			return nil
		}
	}

	return fmt.Errorf("invalid bool value: %s", s)
}

func NewProvider(clientID, clientSecret, redirectURL string) (*Provider, error) {
	clientID = strings.TrimSpace(clientID)
	clientSecret = strings.TrimSpace(clientSecret)
	redirectURL = strings.TrimSpace(redirectURL)

    if clientID == "" || clientSecret == "" || redirectURL == "" {
        return nil, errors.New("google oauth: missing required credentials")
    }

    p := &Provider{
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

    return p, nil
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoEndpoint, nil)
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
		Email         string       `json:"email"`
		Name          string       `json:"name"`
		VerifiedEmail flexibleBool `json:"verified_email"` // nolint: tagliatelle
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return auth.OAuthUser{}, err
	}

	return auth.OAuthUser{
		Email:         payload.Email,
		Name:          payload.Name,
		EmailVerified: bool(payload.VerifiedEmail),
	}, nil
}
