package llm

import "context"

type LLM interface {
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)

	Stream(ctx context.Context, req *GenerateRequest) (<-chan StreamEvent, error)
}

type StreamEventType string

const (
	StreamEventTextDelta     StreamEventType = "text_delta"
	StreamEventToolCallDelta StreamEventType = "tool_call_delta"
	StreamEventCompleted     StreamEventType = "completed"
	StreamEventError         StreamEventType = "error"
)

type ToolCallDelta struct {
	Index     int
	ID        string
	Name      string
	Arguments string
}

type StreamEvent struct {
	Type     StreamEventType
	Delta    string
	ToolCall *ToolCallDelta
	Response *GenerateResponse
	Err      error
}
