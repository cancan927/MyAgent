package memory

import (
	"MyAgent/internal/domain"
	"context"
)

// Extractor 实际上需要完成三个工作：
// 1. 识别：哪些内容值得记住？
// 2. 提炼：将自然语言对话转换成简洁、独立的事实
// 3. 结构化：输出符合 Memory 模型的数据
//
// 它不应该负责直接写入数据，也不应该自行决定如何覆盖已有记忆。
// 它只负责生成候选事实

type Extractor interface {
	Extract(ctx context.Context, messages []domain.Message) ([]MemoryCandidate, error)
}

type ExtractRequest struct {
	UserID string

	// 本次执行的完整上下文
	Messages []domain.Message

	// 本次新产生的消息范围
	StartIndex int
}

type MemoryCandidate struct {
	Type    MemoryType `json:"type"`
	Context string     `json:"context"`
}
