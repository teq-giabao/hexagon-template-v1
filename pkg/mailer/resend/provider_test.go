package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	resendlib "github.com/resend/resend-go/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type capturedEmail struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

func newTestProvider(t *testing.T, handler func(*capturedEmail)) *Provider {
	rt := &captureTransport{
		t:       t,
		handler: handler,
	}

	client := resendlib.NewCustomClient(&http.Client{Transport: rt}, "api-key")
	client.BaseURL, _ = url.Parse("https://resend.local/")

	return &Provider{
		client:    client,
		fromEmail: "from@example.com",
		fromName:  "Sender",
	}
}

type captureTransport struct {
	t       *testing.T
	handler func(*capturedEmail)
}

func (c *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Path != "/emails" || req.Method != http.MethodPost {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("not found")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	}

	body, err := io.ReadAll(req.Body)
	require.NoError(c.t, err)

	var payload capturedEmail
	err = json.Unmarshal(body, &payload)
	require.NoError(c.t, err)

	c.handler(&payload)

	respBody := io.NopCloser(bytes.NewReader([]byte(`{"id":"email_123"}`)))

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       respBody,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    req,
	}, nil
}

func TestNewProvider_Validation(t *testing.T) {
	_, err := NewProvider("", "from@example.com", "Sender")
	assert.Error(t, err)

	_, err = NewProvider("key", "", "Sender")
	assert.Error(t, err)
}

func TestFromHeader(t *testing.T) {
	assert.Equal(t, "from@example.com", fromHeader("", "from@example.com"))
	assert.Equal(t, "Sender <from@example.com>", fromHeader("Sender", "from@example.com"))
}

func TestDisplayName(t *testing.T) {
	assert.Equal(t, "there", displayName(""))
	assert.Equal(t, "John", displayName("  John  "))
}

func TestProvider_SendResetPasswordEmail(t *testing.T) {
	provider := newTestProvider(t, func(payload *capturedEmail) {
		assert.Equal(t, "Sender <from@example.com>", payload.From)
		assert.Equal(t, []string{"to@example.com"}, payload.To)
		assert.Equal(t, "Reset your password", payload.Subject)
		assert.True(t, strings.Contains(payload.Html, "reset"))
	})

	err := provider.SendResetPasswordEmail(context.Background(), "to@example.com", "John", "https://reset")
	assert.NoError(t, err)
}

func TestProvider_SendVerifyEmail(t *testing.T) {
	provider := newTestProvider(t, func(payload *capturedEmail) {
		assert.Equal(t, "Sender <from@example.com>", payload.From)
		assert.Equal(t, []string{"to@example.com"}, payload.To)
		assert.Equal(t, "Verify your email", payload.Subject)
		assert.True(t, strings.Contains(payload.Html, "verify"))
	})

	err := provider.SendVerifyEmail(context.Background(), "to@example.com", "John", "https://verify")
	assert.NoError(t, err)
}

func TestProvider_SendEmail_InvalidPayload(t *testing.T) {
	provider := newTestProvider(t, func(payload *capturedEmail) {})

	err := provider.SendResetPasswordEmail(context.Background(), "", "John", "https://reset")
	assert.Error(t, err)

	err = provider.SendVerifyEmail(context.Background(), "to@example.com", "John", "")
	assert.Error(t, err)
}

func TestProvider_SendEmail_ContextCanceled(t *testing.T) {
	provider := newTestProvider(t, func(payload *capturedEmail) {})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := provider.SendResetPasswordEmail(ctx, "to@example.com", "John", "https://reset")
	assert.Error(t, err)
}
