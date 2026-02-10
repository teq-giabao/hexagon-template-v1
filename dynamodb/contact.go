package dynamodb

import (
	"context"
	"fmt"
	"hexagon/contact"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
)

type ContactRepository struct {
	client *dynamodb.Client
	table  string
}

type contactItem struct {
	ID    string `dynamodbav:"id"`
	Name  string `dynamodbav:"name"`
	Phone string `dynamodbav:"phone"`
}

func NewContactRepository(client *dynamodb.Client, table string) *ContactRepository {
	return &ContactRepository{
		client: client,
		table:  table,
	}
}

func (r *ContactRepository) CreateContact(ctx context.Context, c contact.Contact) error {
	if err := validateTable(r.table); err != nil {
		return err
	}

	item := contactItem{
		ID:    uuid.NewString(),
		Name:  c.Name,
		Phone: c.Phone,
	}
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("dynamodb: marshal contact: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &r.table,
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("dynamodb: put contact: %w", err)
	}

	return nil
}

func (r *ContactRepository) AllContacts(ctx context.Context) ([]contact.Contact, error) {
	if err := validateTable(r.table); err != nil {
		return nil, err
	}

	var contacts []contact.Contact
	paginator := dynamodb.NewScanPaginator(r.client, &dynamodb.ScanInput{
		TableName: &r.table,
	})
	for paginator.HasMorePages() {
		out, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("dynamodb: scan contacts: %w", err)
		}

		var items []contactItem
		if err := attributevalue.UnmarshalListOfMaps(out.Items, &items); err != nil {
			return nil, fmt.Errorf("dynamodb: unmarshal contacts: %w", err)
		}
		for _, item := range items {
			contacts = append(contacts, contact.Contact{
				Name:  item.Name,
				Phone: item.Phone,
			})
		}
	}

	return contacts, nil
}
