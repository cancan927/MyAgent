package llm

import (
	"MyAgent/internal/domain"
	"encoding/json"
)

type GenerateRequest struct {
	Model       string
	Messages    []domain.Message
	Tools       []ToolDefinition
	Temperature *float64
	MaxTokens   int
}

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}
