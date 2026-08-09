package persistance

import "context"

type Persistence interface {
	CreateRecord(ctx context.Context, uid string, message string, ttl int)
	ReadRecord(ctx context.Context, uid string)
	DeleteRecord(ctx context.Context, uid string)
}
