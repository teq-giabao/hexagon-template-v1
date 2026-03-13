package google

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type mockTransport struct {
	t *testing.T
}

func (t *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	switch {
	case req.URL.Path == "/token":
		return mockResponse(req, http.StatusOK, `{"access_token":"token-123","token_type":"bearer","expires_in":3600}`), nil
	case req.URL.Host == "www.googleapis.com" && req.URL.Path == "/oauth2/v2/userinfo":
		return mockResponse(req, http.StatusOK, `{"id":"gid-1","email":"user@example.com","name":"User","verified_email":"true"}`), nil
	default:
		return mockResponse(req, http.StatusNotFound, `{"error":"not_found"}`), nil
	}
}

func mockResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    req,
	}
}

func TestFlexibleBool_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"string-true", "\"true\"", true},
		{"string-false", "\"false\"", false},
		{"number-1", "1", true},
		{"number-0", "0", false},
		{"null", "null", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b flexibleBool

			err := json.Unmarshal([]byte(tc.in), &b)
			require.NoError(t, err)
			assert.Equal(t, tc.want, bool(b))
		})
	}
}

func TestFlexibleBool_UnmarshalJSON_Invalid(t *testing.T) {
	var b flexibleBool

	err := json.Unmarshal([]byte("\"notabool\""), &b)
	assert.Error(t, err)
}

func TestNewProvider_Validation(t *testing.T) {
	_, err := NewProvider("", "secret", "http://localhost/callback")
	assert.Error(t, err)

	_, err = NewProvider("id", "", "http://localhost/callback")
	assert.Error(t, err)

	_, err = NewProvider("id", "secret", "")
	assert.Error(t, err)
}

func TestProvider_Exchange(t *testing.T) {
	cfg := &oauth2.Config{
		ClientID:     "id",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://example.com/auth",
			TokenURL: "https://oauth.local/token",
		},
		Scopes: []string{"openid", "email", "profile"},
	}

	p := &Provider{config: cfg}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, &http.Client{
		Transport: &mockTransport{t: t},
		Timeout:   time.Second,
	})

	user, err := p.Exchange(ctx, "code-123")
	require.NoError(t, err)
	assert.Equal(t, "gid-1", user.ProviderUserID)
	assert.Equal(t, "user@example.com", user.Email)
	assert.Equal(t, "User", user.Name)
	assert.True(t, user.EmailVerified)
}

func TestProvider_Exchange_NotConfigured(t *testing.T) {
	p := &Provider{}
	_, err := p.Exchange(context.Background(), "code")
	assert.Error(t, err)
}

func TestProvider_AuthCodeURL(t *testing.T) {
	p, err := NewProvider("id", "secret", "http://localhost/callback")
	require.NoError(t, err)

	url := p.AuthCodeURL("state-1")
	assert.True(t, strings.Contains(url, "state=state-1"))
}
