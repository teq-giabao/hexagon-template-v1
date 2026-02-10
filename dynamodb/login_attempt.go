package dynamodb

import (
	"context"
	"fmt"
	"hexagon/auth"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type LoginAttemptRepository struct {
	client *dynamodb.Client
	table  string
}

func NewLoginAttemptRepository(client *dynamodb.Client, table string) *LoginAttemptRepository {
	return &LoginAttemptRepository{
		client: client,
		table:  table,
	}
}

func (r *LoginAttemptRepository) Get(ctx context.Context, email string) (auth.LoginAttempt, error) {
	if err := validateTable(r.table); err != nil {
		return auth.LoginAttempt{}, err
	}

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.table,
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return auth.LoginAttempt{}, fmt.Errorf("dynamodb: get login attempt: %w", err)
	}
	if len(out.Item) == 0 {
		return auth.LoginAttempt{}, nil
	}

	var attempt auth.LoginAttempt
	if v, ok := out.Item["failed_count"].(*types.AttributeValueMemberN); ok {
		if n, err := strconv.Atoi(v.Value); err == nil {
			attempt.FailedCount = n
		}
	}

	if v, ok := out.Item["jailed_until"].(*types.AttributeValueMemberS); ok && v.Value != "" {
		parsed, err := time.Parse(time.RFC3339Nano, v.Value)
		if err != nil {
			return auth.LoginAttempt{}, fmt.Errorf("dynamodb: parse jailed_until: %w", err)
		}
		attempt.JailedUntil = parsed.UTC()
	}

	return attempt, nil
}

func (r *LoginAttemptRepository) Save(ctx context.Context, email string, attempt auth.LoginAttempt) error {
	if err := validateTable(r.table); err != nil {
		return err
	}

	item := map[string]types.AttributeValue{
		"email":        &types.AttributeValueMemberS{Value: email},
		"failed_count": &types.AttributeValueMemberN{Value: strconv.Itoa(attempt.FailedCount)},
	}

	if !attempt.JailedUntil.IsZero() {
		item["jailed_until"] = &types.AttributeValueMemberS{
			Value: attempt.JailedUntil.UTC().Format(time.RFC3339Nano),
		}
	}

	_, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &r.table,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("dynamodb: put login attempt: %w", err)
	}

	return nil
}

func (r *LoginAttemptRepository) Reset(ctx context.Context, email string) error {
	if err := validateTable(r.table); err != nil {
		return err
	}

	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.table,
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return fmt.Errorf("dynamodb: delete login attempt: %w", err)
	}

	return nil
}
