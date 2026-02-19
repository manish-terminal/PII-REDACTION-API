package store

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/asoasis/pii-redaction-api/internal/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBStore struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBStore(ctx context.Context, region, tableName string) (*DynamoDBStore, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	return &DynamoDBStore{
		client:    client,
		tableName: tableName,
	}, nil
}

func (s *DynamoDBStore) StoreToken(ctx context.Context, mapping model.TokenMapping) error {
	ttl := strconv.FormatInt(mapping.ExpiresAt.Unix(), 10)
	_, err := s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.tableName),
		Item: map[string]types.AttributeValue{
			"token":       &types.AttributeValueMemberS{Value: mapping.Token},
			"original":    &types.AttributeValueMemberS{Value: mapping.Value},
			"entity_type": &types.AttributeValueMemberS{Value: mapping.EntityType},
			"expires_at":  &types.AttributeValueMemberN{Value: ttl},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to store token in DynamoDB: %w", err)
	}
	return nil
}

func (s *DynamoDBStore) GetToken(ctx context.Context, token string) (*model.TokenMapping, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.tableName),
		Key: map[string]types.AttributeValue{
			"token": &types.AttributeValueMemberS{Value: token},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get token from DynamoDB: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("token not found")
	}

	expiresAtUnix, _ := strconv.ParseInt(result.Item["expires_at"].(*types.AttributeValueMemberN).Value, 10, 64)
	expiresAt := time.Unix(expiresAtUnix, 0)

	if time.Now().After(expiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	return &model.TokenMapping{
		Token:      result.Item["token"].(*types.AttributeValueMemberS).Value,
		Value:      result.Item["original"].(*types.AttributeValueMemberS).Value,
		EntityType: result.Item["entity_type"].(*types.AttributeValueMemberS).Value,
		ExpiresAt:  expiresAt,
	}, nil
}
