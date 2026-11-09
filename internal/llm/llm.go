package llm

import "context"

type LLM interface {
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)

	Stream(ctx context.Context, req *GenerateRequest) (<-chan StreamEvent, error)
}
