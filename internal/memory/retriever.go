package memory

import "context"

// Retriever 负责查找可能相关的已有记忆，不负责决定是否更新

type Retriever interface {
	Search(ctx context.Context, UserID string, query string, limit int) ([]Memory, error)
}
