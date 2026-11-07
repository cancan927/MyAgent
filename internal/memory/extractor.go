package memory

import (
	"MyAgent/internal/domain"
	"context"
)

type Extractor interface {
	Extract(ctx context.Context, messages []domain.Message) ([]Memory, error)
}
