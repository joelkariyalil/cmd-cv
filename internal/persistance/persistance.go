package persistance

import "context"

type Persistance interface {
	GetRecord(ctx context.Context, uid string)
	PutRecord(ctx context.Context, uid string, message string)
	DeleteRecord(ctx context.Context, uid string)
}
