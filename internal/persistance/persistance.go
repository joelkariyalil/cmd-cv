package persistance

import (
	"context"
	"time"
)

type Repo interface {
	CreateRecord(ctx context.Context, uid, message string, ttl time.Duration) error
	ReadRecord(ctx context.Context, uid string) (*Record, error)
	DeleteRecord(ctx context.Context, uid string) error
}
																															