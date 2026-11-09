package memory

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

// Manager 负责记忆的写入决策
// 判重 · 冲突消解 · 合并 · 打分 · 过滤

type Manager struct {
	store Store

	retriever Retriever
	resolver  Resolver
}

func NewManager(store Store, retriever Retriever, resolver Resolver) *Manager {
	return &Manager{
		store:     store,
		retriever: retriever,
		resolver:  resolver,
	}

}

// Process 逐条处理候选记忆，根据 Resolver 的决策执行忽略、新增或更新。
// 任一步骤失败都会立即返回错误；此前已经保存的记忆不会自动回滚。
func (m *Manager) Process(ctx context.Context, userID string, candidates []MemoryCandidate) error {
	// 记忆必须关联用户，避免写入没有归属的数据。
	if userID == "" {
		return ErrInvalidMemory
	}

	for _, candidate := range candidates {
		// 跳过空白内容，避免无意义的检索和写入。
		if strings.TrimSpace(candidate.Context) == "" {
			continue
		}

		// 在当前用户的记忆范围内，检索最多 10 条相关记忆，供后续判重和冲突处理。
		existing, err := m.retriever.Search(ctx, userID, candidate.Context, 10)
		if err != nil {
			return err
		}

		// 将候选事实与已有记忆交给 Resolver，决定写入动作及最终保存的内容。
		resolution, err := m.resolver.Resolve(ctx, candidate, existing)
		if err != nil {
			return err
		}

		switch resolution.Action {
		case ActionIgnore:
			// 无需写入时直接处理下一条候选。
			continue

		case ActionAdd:
			// 为新记忆生成独立 ID，保留候选类型，并采用 Resolver 整理后的内容。
			mem := &Memory{
				ID:      uuid.NewString(),
				UserID:  userID,
				Type:    candidate.Type,
				Content: resolution.Content,
			}
			if err := m.store.Save(ctx, mem); err != nil {
				return err
			}

		case ActionUpdate:
			// 仅允许更新本次检索结果中的目标，确保它是 Resolver 评估过的记忆。
			// 用户隔离依赖 Retriever 按上面的 userID 限定检索范围。
			var target *Memory
			for i := range existing {
				if existing[i].ID == resolution.TargetID {
					target = &existing[i]
					break
				}
			}

			if target == nil {
				// 拒绝不存在于检索结果中的目标，避免按任意 ID 更新记忆。
				return errors.New("invalid update target")
			}

			// 在原记忆上替换内容，保留其 ID、用户归属和类型，再交给 Store 保存。
			target.Content = resolution.Content

			if err := m.store.Save(ctx, target); err != nil {
				return err
			}

		default:
			// 未知动作视为决策错误，停止处理，避免静默忽略异常结果。
			return errors.New("unknown resolution action")
		}

	}

	return nil

}
