package memory

import "context"

type Resolver interface {
	Resolve(ctx context.Context, candidate MemoryCandidate, existing []Memory) (Resolution, error)
}

type Action string

const (
	ActionAdd    Action = "add"
	ActionUpdate Action = "update"
	ActionIgnore Action = "ignore"
)

type Resolution struct {
	Action Action

	// 更新已有记忆时使用
	TargetID string

	// 最终需要保存的内容
	Content string
}
