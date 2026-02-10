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

	if opts.Endpoint != "" {
		loadOpts = append(loadOpts, awscfg.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, resolvedRegion string, options ...interface{}) (aws.Endpoint, error) {
				if service == dynamodb.ServiceID {
					return aws.Endpoint{
						URL:           opts.Endpoint,
						SigningRegion: region,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		))
	}

	cfg, err := awscfg.LoadDefaultConfig(ctx, loadOpts...)
	if err != nil {
		return nil, fmt.Errorf("dynamodb: load aws config: %w", err)
	}

	return dynamodb.NewFromConfig(cfg), nil
}

func validateTable(table string) error {
	if strings.TrimSpace(table) == "" {
		return errors.New("dynamodb: table name is required")
	}
	return nil
}
