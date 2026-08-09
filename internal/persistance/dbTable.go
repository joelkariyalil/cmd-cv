package persistance

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamoDBAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

type Record struct {
	DbKey
	Message        string `dynamodbav:"message"`
	CreatedAt      int64  `dynamodbav:"created_at"`
	LastAccessedAt int64  `dynamodbav:"last_accessed_at"`
	ExpiresAt      int64  `dynamodbav:"expires_at"`
}

type DbTable struct {
	dbClient  dynamoDBAPI
	tableName string
}

func NewDbTable(dbClient dynamoDBAPI, tableName string) *DbTable {
	return &DbTable{
		dbClient:  dbClient,
		tableName: tableName,
	}
}

func (r *DbTable) CreateRecord(ctx context.Context, uid, message string, ttl time.Duration) error {
	currentTime := time.Now()
	rec := Record{
		DbKey:          recordKey(uid),
		Message:        message,
		CreatedAt:      currentTime.Unix(),
		LastAccessedAt: currentTime.Unix(),
		ExpiresAt:      currentTime.Add(ttl).Unix(),
	}

	item, err := attributevalue.MarshalMap(rec)
	if err != nil {
		return err
	}

	_, err = r.dbClient.PutItem(ctx, &dynamodb.PutItemInput{
		Item:      item,
		TableName: aws.String(r.tableName),
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *DbTable) ReadRecord(ctx context.Context, uid string) (*Record, error) {
	key := recordKey(uid)
	now := time.Now().Unix()

	update := expression.Set(expression.Name("last_accessed_at"), expression.Value(now))
	condition := expression.Name("expires_at").GreaterThan(expression.Value(now))

	expr, err := expression.NewBuilder().WithUpdate(update).WithCondition(condition).Build()
	if err != nil {
		return nil, err
	}

	keyItem, err := attributevalue.MarshalMap(key)
	if err != nil {
		return nil, err
	}

	output, err := r.dbClient.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.tableName),
		Key:                       keyItem,
		UpdateExpression:          expr.Update(),
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              types.ReturnValueAllNew,
	})
	if err != nil {
		return nil, err
	}

	var record Record
	if err := attributevalue.UnmarshalMap(output.Attributes, &record); err != nil {
		return nil, err
	}

	return &record, nil
}

func (r *DbTable) DeleteRecord(ctx context.Context, uid string) error {
	key, err := attributevalue.MarshalMap(recordKey(uid))
	if err != nil {
		return err
	}

	_, err = r.dbClient.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key:       key,
	})
	return err
}

func recordKey(uid string) DbKey {
	return DbKey{
		PK: uid,
	}
}
