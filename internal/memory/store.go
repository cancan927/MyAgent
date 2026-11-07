package memory

import "context"

type Store interface {
	Save(ctx context.Context, mem *Memory) error

	Search(ctx context.Context, userID string, query string, limit int) ([]Memory, error)

	Delete(ctx context.Context, userID string, id string) error
}
