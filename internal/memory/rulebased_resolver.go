package memory

import (
	"context"
	"strings"
)

// 这里的 RuleBasedResolver 只实现完全重复检测，其他情况全部新增

type RuleBasedResolver struct{}

func NewRuleBasedResolver() *RuleBasedResolver {
	return &RuleBasedResolver{}
}

func (r *RuleBasedResolver) Resolve(ctx context.Context, candidate MemoryCandidate, existing []Memory) (Resolution, error) {
	if err := ctx.Err(); err != nil {
		return Resolution{}, err
	}

	content := strings.TrimSpace(candidate.Context)

	for _, mem := range existing {
		if mem.Type == candidate.Type && strings.EqualFold(strings.TrimSpace(mem.Content), content) {
			return Resolution{Action: ActionIgnore}, nil
		}
	}

	return Resolution{
		Action:  ActionAdd,
		Content: content,
	}, nil

}
