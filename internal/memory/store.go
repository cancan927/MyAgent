package memory

import "context"

// Store 只负责保存、查询、删除记忆，不负责理解记忆的语义

type Store interface {
	Save(ctx context.Context, mem *Memory) error

	Search(ctx context.Context, userID string, query string, limit int) ([]Memory, error)

	Delete(ctx context.Context, userID string, id string) error
}
