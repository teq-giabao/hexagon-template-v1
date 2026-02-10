package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"hexagon/errs"
	"hexagon/user"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type UserRepository struct {
	client *dynamodb.Client
	table  string
	now    func() time.Time
}

type userItem struct {
	Email    string `dynamodbav:"email"`
	Username string `dynamodbav:"username"`
	Password string `dynamodbav:"password"`
	ID       int64  `dynamodbav:"id"`
}

func NewUserRepository(client *dynamodb.Client, table string) *UserRepository {
	return &UserRepository{
		client: client,
		table:  table,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if err := validateTable(r.table); err != nil {
		return user.User{}, err
	}

	out, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &r.table,
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return user.User{}, fmt.Errorf("dynamodb: get user: %w", err)
	}
	if len(out.Item) == 0 {
		return user.User{}, user.ErrInvalidEmail
	}

	var item userItem
	if err := attributevalue.UnmarshalMap(out.Item, &item); err != nil {
		return user.User{}, fmt.Errorf("dynamodb: unmarshal user: %w", err)
	}

	return user.User{
		ID:       item.ID,
		Username: item.Username,
		Email:    item.Email,
		Password: item.Password,
	}, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, u user.User) error {
	_, err := r.createUser(ctx, u)
	return err
}

func (r *UserRepository) CreateUserTx(
	ctx context.Context,
	u user.User,
	fn func(created user.User) error,
) (user.User, error) {
	created, err := r.createUser(ctx, u)
	if err != nil {
		return user.User{}, err
	}

	if fn != nil {
		if err := fn(created); err != nil {
			_ = r.deleteByEmail(ctx, created.Email)
			return user.User{}, err
		}
	}

	return created, nil
}

func (r *UserRepository) AllUsers(ctx context.Context) ([]user.User, error) {
	if err := validateTable(r.table); err != nil {
		return nil, err
	}

	var users []user.User
	paginator := dynamodb.NewScanPaginator(r.client, &dynamodb.ScanInput{
		TableName: &r.table,
	})
	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("dynamodb: scan users: %w", err)
		}

		var items []userItem
		if err := attributevalue.UnmarshalListOfMaps(out.Items, &items); err != nil {
			return nil, fmt.Errorf("dynamodb: unmarshal users: %w", err)
		}

		for _, item := range items {
			users = append(users, user.User{
				ID:       item.ID,
				Username: item.Username,
				Email:    item.Email,
				Password: item.Password,
			})
		}
	}

	return users, nil
}

func (r *UserRepository) createUser(ctx context.Context, u user.User) (user.User, error) {
	if err := validateTable(r.table); err != nil {
		return user.User{}, err
	}

	id := u.ID
	if id == 0 {
		id = r.now().UnixNano()
	}

	item := userItem{
		ID:       id,
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
	}
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return user.User{}, fmt.Errorf("dynamodb: marshal user: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           &r.table,
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(email)"),
	})
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return user.User{}, errs.Errorf(errs.ECONFLICT, "user already exists")
		}
		return user.User{}, fmt.Errorf("dynamodb: put user: %w", err)
	}

	return user.User{
		ID:       id,
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) deleteByEmail(ctx context.Context, email string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &r.table,
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return fmt.Errorf("dynamodb: delete user rollback: %w", err)
	}
	return nil
}
