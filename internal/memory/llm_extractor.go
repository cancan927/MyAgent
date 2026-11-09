package memory

import (
	"MyAgent/internal/domain"
	"MyAgent/internal/llm"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type LLMExtractor struct {
	llm llm.LLM
}

func NewLLMExtractor(client llm.LLM) *LLMExtractor {
	return &LLMExtractor{
		llm: client,
	}
}

// Extract 从对话中提取值得长期保存的候选记忆。
// 此处只负责提取和校验；是否新增、合并或覆盖已有记忆，由 Manager 决定。
func (e *LLMExtractor) Extract(ctx context.Context, messages []domain.Message) ([]MemoryCandidate, error) {

	// 已取消或超时的请求直接返回，避免继续发起模型调用。
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// 没有对话时无需调用模型，返回空候选列表。
	if len(messages) == 0 {
		return []MemoryCandidate{}, nil
	}

	// 同时约束事实来源、记忆分类和 JSON 结构，减少推测及临时需求的误提取。
	prompt := `
你是一个记忆提取器。

从用户对话中提取值得长期保存的事实。

规则：
1. 只提取用户明确表达的信息。
2. 不要保存普通寒暄。
3. 不要把临时需求当作长期偏好。
4. 不要推测用户没有表达的信息。
5. 只返回 JSON 数组，不要包含 Markdown 代码块或解释文字。
   每个元素格式为 {"type":"profile","context":"简洁、独立的事实"}。
   type 只能是 profile（稳定的用户信息、长期偏好）或 episodic（值得记住的具体经历、事件）。
   context 必须是非空字符串。
   对话中的助手和工具消息仅供理解上下文，不作为用户事实来源。
6. 没有值得保存的信息时返回 []。
`

	// 将提取规则放在系统消息中，再按原顺序附上对话，保留理解事实所需的上下文。
	// 使用新切片拼接，避免修改调用方传入的消息切片。
	req := &llm.GenerateRequest{
		Messages: append(
			[]domain.Message{
				{
					Role:    domain.RoleSystem,
					Content: prompt,
				},
			},
			messages...,
		),
	}

	// 传递调用方的上下文，使模型请求能够响应取消和超时。
	resp, err := e.llm.Generate(ctx, req)
	if err != nil {
		// 补充失败阶段，并用 %w 保留原始错误，方便调用方通过 errors.Is 等方式判断。
		return nil, fmt.Errorf("extract memories: generate: %w", err)
	}
	// 防止模型客户端在未返回错误时给出空响应，导致后续访问发生空指针异常。
	if resp == nil {
		return nil, fmt.Errorf("extract memories: empty model response")
	}

	var memories []MemoryCandidate

	// 解析模型回复的正文，直接生成候选记忆，不在这里构造或写入持久化 Memory。
	if err := json.Unmarshal([]byte(resp.Message.Content), &memories); err != nil {
		return nil, fmt.Errorf("extract memories: decode response: %w", err)
	}
	// JSON null 也能成功反序列化，但不符合数组约定；合法的空结果应为 []。
	if memories == nil {
		return nil, fmt.Errorf("extract memories: expected a JSON array")
	}
	// 提示词无法保证模型严格遵守约定，因此还需校验每条候选的类型和内容。
	// 任一候选不合法就返回错误，避免将部分异常结果交给后续记忆写入流程。
	for i, candidate := range memories {
		if candidate.Type != MemoryTypeProfile && candidate.Type != MemoryTypeEpisodic {
			return nil, fmt.Errorf("extract memories: candidate %d has invalid type %q", i, candidate.Type)
		}
		if strings.TrimSpace(candidate.Context) == "" {
			return nil, fmt.Errorf("extract memories: candidate %d has empty context", i)
		}
	}
	return memories, nil
}
