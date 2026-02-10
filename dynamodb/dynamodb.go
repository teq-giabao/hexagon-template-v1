package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Options struct {
	Region       string
	Endpoint     string
	AccessKey    string
	SecretKey    string
	SessionToken string
}

func NewClient(ctx context.Context, opts Options) (*dynamodb.Client, error) {
	region := strings.TrimSpace(opts.Region)
	if region == "" {
		return nil, errors.New("dynamodb: region is required")
	}

	loadOpts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion(region),
	}

	if opts.AccessKey != "" || opts.SecretKey != "" || opts.SessionToken != "" {
		if opts.AccessKey == "" || opts.SecretKey == "" {
			return nil, errors.New("dynamodb: access key and secret key must be set together")
		}
		loadOpts = append(loadOpts, awscfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(opts.AccessKey, opts.SecretKey, opts.SessionToken),
		))
	}

	cfg, err := awscfg.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("dynamodb: load aws config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if opts.Endpoint != "" {
			o.BaseEndpoint = aws.String(opts.Endpoint)
		}
	})

	return client, nil
}

func validateTable(table string) error {
	if strings.TrimSpace(table) == "" {
		return errors.New("dynamodb: table name is required")
	}
	return nil
}
