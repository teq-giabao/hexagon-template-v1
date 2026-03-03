package resend

import (
	"context"
	"fmt"
	"strings"

	resendlib "github.com/resend/resend-go/v3"
)

type Provider struct {
	client    *resendlib.Client
	fromEmail string
	fromName  string
}

func NewProvider(apiKey, fromEmail, fromName string) (*Provider, error) {
	apiKey = strings.TrimSpace(apiKey)
	fromEmail = strings.TrimSpace(fromEmail)
	fromName = strings.TrimSpace(fromName)
	if apiKey == "" || fromEmail == "" {
		return nil, fmt.Errorf("invalid resend configuration")
	}

	return &Provider{
		client:    resendlib.NewClient(apiKey),
		fromEmail: fromEmail,
		fromName:  fromName,
	}, nil
}

func (p *Provider) SendResetPasswordEmail(ctx context.Context, toEmail, toName, resetURL string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	toEmail = strings.TrimSpace(toEmail)
	toName = strings.TrimSpace(toName)
	resetURL = strings.TrimSpace(resetURL)
	if toEmail == "" || resetURL == "" {
		return fmt.Errorf("invalid reset mail payload")
	}

	html := fmt.Sprintf(
		"<p>Hello %s,</p><p>Click <a href=\"%s\">here</a> to reset your password.</p><p>If you did not request this, ignore this email.</p>",
		displayName(toName),
		resetURL,
	)

	params := &resendlib.SendEmailRequest{
		From:    fromHeader(p.fromName, p.fromEmail),
		To:      []string{toEmail},
		Subject: "Reset your password",
		Html:    html,
	}

	_, err := p.client.Emails.Send(params)
	return err
}

func fromHeader(name, email string) string {
	if strings.TrimSpace(name) == "" {
		return email
	}
	return fmt.Sprintf("%s <%s>", strings.TrimSpace(name), email)
}

func displayName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "there"
	}
	return strings.TrimSpace(name)
}
