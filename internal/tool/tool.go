package tool

import (
	"context"
	"encoding/json"
)

type Definition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type Tool interface {
	Definition() Definition
	Execute(ctx context.Context, arguments json.RawMessage) (string, error)
}
