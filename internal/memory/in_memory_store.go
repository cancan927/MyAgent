package memory

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidMemory  = errors.New("invalid memory")
	ErrMemoryNotFound = errors.New("memory not found")
)

type InMemoryStore struct {
	mu       sync.RWMutex
	memories map[string]Memory
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		memories: make(map[string]Memory),
	}
}

func (s *InMemoryStore) Save(ctx context.Context, mem *Memory) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if mem == nil || mem.ID == "" || mem.UserID == "" || mem.Content == "" {
		return ErrInvalidMemory
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 去重必须与写入使用同一把写锁，避免并发请求同时通过检查后重复保存。
	for _, existing := range s.memories {
		if existing.UserID == mem.UserID &&
			existing.Type == mem.Type &&
			existing.Content == mem.Content {
			// 重复保存是幂等操作，不写入新记录，也不刷新原记录时间。
			return nil
		}
	}

	now := time.Now()

	if old, ok := s.memories[mem.ID]; ok {
		// 不是同一个 User 的记忆不放在一起
		if old.UserID != mem.UserID {
			return ErrInvalidMemory
		}
		mem.CreatedAt = old.CreatedAt
	} else if mem.CreatedAt.IsZero() {
		mem.CreatedAt = now
	}

	mem.UpdatedAt = now
	// 这里使用 Memory ID 作为 Key，如果两个用户保存了相同的 ID，就不能允许后一个用户覆盖前一个用户的数据
	// 生产环境中还可以直接使用（UserID，MemoryID）作为联合 Key
	s.memories[mem.ID] = *mem

	return nil

}

// Search 使用空白分隔的关键词检索指定用户的记忆，命中任意关键词即可返回。
func (s *InMemoryStore) Search(ctx context.Context, userID string, query string, limit int) ([]Memory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if userID == "" || limit <= 0 {
		return nil, nil
	}

	keywords := strings.Fields(strings.ToLower(query))
	if len(keywords) == 0 {
		return nil, nil
	}

	// 去除重复关键词，避免相同关键词被重复计分，影响相关度排序。
	uniqueKeywords := make([]string, 0, len(keywords))
	seen := make(map[string]struct{}, len(keywords))
	for _, keyword := range keywords {
		if _, ok := seen[keyword]; ok {
			continue
		}
		seen[keyword] = struct{}{}
		uniqueKeywords = append(uniqueKeywords, keyword)
	}

	type searchResult struct {
		memory     Memory
		matchCount int
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 只保留命中至少一个关键词的记忆，并记录命中数量作为相关度分数。
	candidates := make([]searchResult, 0)
	for _, mem := range s.memories {
		if mem.UserID != userID {
			continue
		}

		content := strings.ToLower(mem.Content)
		matchCount := 0

		for _, keyword := range uniqueKeywords {
			if strings.Contains(content, keyword) {
				matchCount++
			}
		}

		if matchCount > 0 {
			candidates = append(candidates, searchResult{
				memory:     mem,
				matchCount: matchCount,
			})
		}
	}

	// 相关度优先；分数相同时优先返回最近更新的记忆，最后使用 ID 保证结果稳定。
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].matchCount != candidates[j].matchCount {
			return candidates[i].matchCount > candidates[j].matchCount
		}
		if !candidates[i].memory.UpdatedAt.Equal(candidates[j].memory.UpdatedAt) {
			return candidates[i].memory.UpdatedAt.After(candidates[j].memory.UpdatedAt)
		}
		return candidates[i].memory.ID < candidates[j].memory.ID
	})

	// 必须在完成相关度排序后再截断，否则可能遗漏匹配度更高的记忆。
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}

	results := make([]Memory, len(candidates))
	for i, candidate := range candidates {
		results[i] = candidate.memory
	}
	return results, nil
}

func (s *InMemoryStore) Delete(ctx context.Context, userID string, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	mem, ok := s.memories[id]
	if !ok || mem.UserID != userID {
		return ErrMemoryNotFound
	}

	delete(s.memories, id)

	return nil
}
