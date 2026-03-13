package auth

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithClientInfo_NormalizesAndStores(t *testing.T) {
	ctx := context.Background()
	longAgent := strings.Repeat("a", 600)
	longIP := strings.Repeat("1", 100)

	ctx = WithClientInfo(ctx, ClientInfo{
		UserAgent: "  " + longAgent + "  ",
		IPAddress: "  " + longIP + "  ",
	})

	info := clientInfoFromContext(ctx)
	assert.Len(t, info.UserAgent, 512)
	assert.Len(t, info.IPAddress, 64)
	assert.Equal(t, strings.Repeat("a", 512), info.UserAgent)
	assert.Equal(t, strings.Repeat("1", 64), info.IPAddress)
}

func TestClientInfoFromContext_NilContext(t *testing.T) {
	info := clientInfoFromContext(context.TODO())
	assert.Equal(t, ClientInfo{}, info)
}

func TestClientInfoFromContext_MissingValue(t *testing.T) {
	info := clientInfoFromContext(context.Background())
	assert.Equal(t, ClientInfo{}, info)
}
