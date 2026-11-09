package llm

type StreamEventType string

const (
	StreamEventTextDelta     StreamEventType = "text_delta"
	StreamEventToolCallDelta StreamEventType = "tool_call_delta"
	StreamEventCompleted     StreamEventType = "completed"
	StreamEventError         StreamEventType = "error"
)

type StreamEvent struct {
	Type     StreamEventType
	Delta    string
	ToolCall *ToolCallDelta
	Response *GenerateResponse
	Err      error
}

type ToolCallDelta struct {
	Index     int
	ID        string
	Name      string
	Arguments string
}
