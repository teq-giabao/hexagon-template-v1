package s3

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUploader_MissingBucket(t *testing.T) {
	_, err := NewUploader(context.Background(), Config{})
	assert.ErrorIs(t, err, ErrMissingBucket)
}

func TestUploader_Upload_WithEndpoint(t *testing.T) {
	var gotPath string

	var gotContentType string

	var gotContentLength int64

	var bodyBytes []byte

	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotPath = req.URL.Path
		gotContentType = req.Header.Get("Content-Type")
		gotContentLength = req.ContentLength
		bodyBytes, _ = io.ReadAll(req.Body)

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	awsCfg := aws.Config{
		Region:      "us-east-1",
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("key", "secret", "")),
		HTTPClient:  &http.Client{Transport: transport},
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String("https://s3.local")
		o.UsePathStyle = true
	})

	uploader := &Uploader{
		client:  client,
		bucket:  "test-bucket",
		baseURL: "https://s3.local/test-bucket",
		prefix:  "uploads",
	}

	payload := []byte("hello")
	url, err := uploader.Upload(context.Background(), "image.jpg", bytes.NewReader(payload), int64(len(payload)), "image/jpeg")
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(url, "https://s3.local/test-bucket/"))
	assert.True(t, strings.HasSuffix(url, "/uploads/image.jpg"))
	assert.Equal(t, "/test-bucket/uploads/image.jpg", gotPath)
	assert.Equal(t, "image/jpeg", gotContentType)
	assert.Equal(t, int64(len(payload)), gotContentLength)
	assert.Equal(t, payload, bodyBytes)
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
