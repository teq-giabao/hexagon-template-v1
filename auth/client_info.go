package auth

import (
	"context"
	"strings"
)

type ClientInfo struct {
	UserAgent string
	IPAddress string
}

type clientInfoContextKey struct{}

func WithClientInfo(ctx context.Context, info ClientInfo) context.Context {
	return context.WithValue(ctx, clientInfoContextKey{}, normalizeClientInfo(info))
}

func clientInfoFromContext(ctx context.Context) ClientInfo {
	if ctx == nil {
		return ClientInfo{}
	}
	info, ok := ctx.Value(clientInfoContextKey{}).(ClientInfo)
	if !ok {
		return ClientInfo{}
	}
	return normalizeClientInfo(info)
}

func normalizeClientInfo(info ClientInfo) ClientInfo {
	return ClientInfo{
		UserAgent: trimWithMaxLen(info.UserAgent, 512),
		IPAddress: trimWithMaxLen(info.IPAddress, 64),
	}
}

func trimWithMaxLen(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if maxLen > 0 && len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}
