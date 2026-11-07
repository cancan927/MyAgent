package llm

import "MyAgent/internal/domain"

type GenerateResponse struct {
	Message domain.Message
	Usage   TokenUsage
	Finish  FinishReason
}

type TokenUsage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// FinishReason 表示结束本次生成，而不是整个 Agent 任务结束
type FinishReason string

const (
	FinishReasonStop      FinishReason = "stop"
	FinishReasonToolCalls FinishReason = "tool_calls"
	FinishReasonLength    FinishReason = "length"
)
