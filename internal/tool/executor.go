package tool

import (
	"context"
	"encoding/json"
	"fmt"
)

type Call struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

type Result struct {
	ToolCallID string
	Name       string
	Content    string
	IsError    bool
}

type Executor struct {
	registry  *Registry
	validator Validator
}

func NewExecutor(registry *Registry, validator Validator) *Executor {
	return &Executor{
		registry:  registry,
		validator: validator,
	}
}

func (e *Executor) Execute(ctx context.Context, call Call) Result {
	result := Result{
		ToolCallID: call.ID,
		Name:       call.Name,
	}

	t, err := e.registry.Get(call.Name)
	if err != nil {
		result.Content = err.Error()
		result.IsError = true
		return result
	}

	def := t.Definition()

	if err := e.validator.Validate(def.Parameters, call.Arguments); err != nil {
		result.Content = err.Error()
		result.IsError = true
		return result
	}

	content, err := t.Execute(ctx, call.Arguments)
	if err != nil {
		result.Content = fmt.Sprintf(
			"tool execution failed: %v",
			err,
		)
		result.IsError = true
		return result
	}

	result.Content = content
	return result
}

// ExecuteAll 一次执行多个工具调用（串行）
func (e *Executor) ExecuteAll(ctx context.Context, calls []Call) []Result {
	results := make([]Result, 0, len(calls))

	for _, call := range calls {
		if err := ctx.Err(); err != nil {
			break
		}

		result := e.Execute(ctx, call)
		results = append(results, result)
	}

	return results
}
