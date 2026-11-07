package memory

import (
	"MyAgent/internal/domain"
	"context"
)

type MockExtractor struct {
	memories []Memory
}

func NewMockExtractor(memories []Memory) *MockExtractor {
	return &MockExtractor{
		memories: memories,
	}
}

func (m *MockExtractor) Extract(ctx context.Context, messages []domain.Message) ([]Memory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make([]Memory, len(m.memories))
	copy(result, m.memories)

	return result, nil
}
