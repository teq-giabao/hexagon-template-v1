package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var ErrMissingBucket = errors.New("s3: bucket is required")

type Config struct {
	Region          string
	Bucket          string
	BaseURL         string
	Prefix          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type Uploader struct {
	client  *s3.Client
	bucket  string
	baseURL string
	prefix  string
}

func NewUploader(ctx context.Context, cfg Config) (*Uploader, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, ErrMissingBucket
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{}
	if strings.TrimSpace(cfg.Region) != "" {
		loadOptions = append(loadOptions, awsconfig.WithRegion(cfg.Region))
	}

	if strings.TrimSpace(cfg.AccessKeyID) != "" && strings.TrimSpace(cfg.SecretAccessKey) != "" {
		provider := credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			cfg.SessionToken,
		)
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(provider))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("s3: load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if strings.TrimSpace(cfg.Endpoint) != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		}
	})

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		if strings.TrimSpace(cfg.Endpoint) != "" {
			baseURL = strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/") + "/" + cfg.Bucket
		} else {
			region := strings.TrimSpace(cfg.Region)
			if region == "" {
				region = awsCfg.Region
			}

			baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.Bucket, region)
		}
	}

	return &Uploader{
		client:  client,
		bucket:  cfg.Bucket,
		baseURL: baseURL,
		prefix:  strings.Trim(strings.TrimSpace(cfg.Prefix), "/"),
	}, nil
}

func (u *Uploader) Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error) {
	key = strings.Trim(strings.TrimSpace(key), "/")
	if u.prefix != "" {
		key = path.Join(u.prefix, key)
	}

	input := &s3.PutObjectInput{
		Bucket:      &u.bucket,
		Key:         &key,
		Body:        body,
		ContentType: aws.String(contentType),
	}
	if size > 0 {
		input.ContentLength = aws.Int64(size)
	}

	_, err := u.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("s3: put object: %w", err)
	}

	return u.baseURL + "/" + key, nil
}
