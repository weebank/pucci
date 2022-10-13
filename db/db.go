package db

import "context"

// Database Service
type DatabaseService interface {
	Connect() (context.Context, context.CancelFunc)
	Disconnect(ctx context.Context, cancel context.CancelFunc)
	Create(ctx context.Context, database, table, id string, doc any) (string, error)
	Read(ctx context.Context, database, table string, filter map[string]interface{}, to any) error
	Update(ctx context.Context, database, table string, filter map[string]interface{}, doc any) error
	Delete(ctx context.Context, database, table, id string) error
}
