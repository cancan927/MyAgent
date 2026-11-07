package domain

import "encoding/json"

// ToolCall represents a model-generated tool request.
type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

// ToolResult represents a tool execution result.
type ToolResult struct {
	ToolCallID string
	Content    string
	IsError    bool
}
