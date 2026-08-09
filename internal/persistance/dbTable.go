package persistance

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamoDBAPI interface {
	Put(ctx context.Context, tableName string, item map[string]types.AttributeValue) error
	Get(ctx context.Context, tableName string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error)
	Delete(ctx context.Context, tableName string, key map[string]types.AttributeValue) error
}

type DbTable struct {
	dbClient  dynamoDBAPI
	tableName string
}

type record struct {
	uid         string    `dynamodbav:"uid"`
	messsge     string    `dynamodbav:"message"`
	createdTime time.Time `dynamodbav:"created_time"`
	Ttl         int       `dynamodbav:"ttl"`
}

func (r *DbTable) PutRecord(ctx context.Context, uid, message string) {

}

func (r *DbTable) GetRecord(ctx context.Context, uid string) {

}

func (r *DbTable) DeleteRecord(ctx context.Context, uid string) {

}
