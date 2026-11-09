package memory

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestInMemoryStoreSave(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []Memory
		wantIDs []string
	}{
		{
			name: "同一用户相同类型和内容只保存一次",
			inputs: []Memory{
				{ID: "mem-1", UserID: "user-1", Type: MemoryTypeProfile, Content: "用户喜欢自然风景"},
				{ID: "mem-2", UserID: "user-1", Type: MemoryTypeProfile, Content: "用户喜欢自然风景"},
			},
			wantIDs: []string{"mem-1"},
		},
		{
			name: "相同用户和内容但类型不同可以保存",
			inputs: []Memory{
				{ID: "mem-1", UserID: "user-1", Type: MemoryTypeProfile, Content: "用户喜欢自然风景"},
				{ID: "mem-2", UserID: "user-1", Type: MemoryTypeEpisodic, Content: "用户喜欢自然风景"},
			},
			wantIDs: []string{"mem-1", "mem-2"},
		},
		{
			name: "不同用户的相同类型和内容可以保存",
			inputs: []Memory{
				{ID: "mem-1", UserID: "user-1", Type: MemoryTypeProfile, Content: "用户喜欢自然风景"},
				{ID: "mem-2", UserID: "user-2", Type: MemoryTypeProfile, Content: "用户喜欢自然风景"},
			},
			wantIDs: []string{"mem-1", "mem-2"},
		},
		{
			name: "同一用户和类型的不同内容可以保存",
			inputs: []Memory{
				{ID: "mem-1", UserID: "user-1", Type: MemoryTypeProfile, Content: "用户喜欢自然风景"},
				{ID: "mem-2", UserID: "user-1", Type: MemoryTypeProfile, Content: "用户喜欢城市旅行"},
			},
			wantIDs: []string{"mem-1", "mem-2"},
		},
		{
			name: "大小写不同的内容不视为完全相同",
			inputs: []Memory{
				{ID: "mem-1", UserID: "user-1", Type: MemoryTypeProfile, Content: "Go"},
				{ID: "mem-2", UserID: "user-1", Type: MemoryTypeProfile, Content: "go"},
			},
			wantIDs: []string{"mem-1", "mem-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryStore()
			for i := range tt.inputs {
				if err := store.Save(t.Context(), &tt.inputs[i]); err != nil {
					t.Fatalf("Save() error = %v", err)
				}
			}

			if len(store.memories) != len(tt.wantIDs) {
				t.Fatalf("Save() stored %d memories, want %d: %#v", len(store.memories), len(tt.wantIDs), store.memories)
			}
			for _, wantID := range tt.wantIDs {
				if _, ok := store.memories[wantID]; !ok {
					t.Errorf("Save() did not retain memory %q", wantID)
				}
			}
		})
	}
}

func TestInMemoryStoreSaveDuplicateDoesNotModifyExistingMemory(t *testing.T) {
	store := NewInMemoryStore()
	original := &Memory{
		ID:      "mem-1",
		UserID:  "user-1",
		Type:    MemoryTypeProfile,
		Content: "用户喜欢自然风景",
	}
	if err := store.Save(t.Context(), original); err != nil {
		t.Fatalf("Save() original error = %v", err)
	}
	want := store.memories[original.ID]

	duplicate := &Memory{
		ID:      "mem-2",
		UserID:  original.UserID,
		Type:    original.Type,
		Content: original.Content,
	}
	if err := store.Save(t.Context(), duplicate); err != nil {
		t.Fatalf("Save() duplicate error = %v", err)
	}

	got := store.memories[original.ID]
	if got != want {
		t.Errorf("Save() modified existing memory: got %#v, want %#v", got, want)
	}
}

func TestInMemoryStoreSaveDeduplicatesConcurrentWrites(t *testing.T) {
	store := NewInMemoryStore()
	errCh := make(chan error, 32)
	var wg sync.WaitGroup

	for i := range 32 {
		wg.Go(func() {
			mem := &Memory{
				ID:      fmt.Sprintf("mem-%d", i),
				UserID:  "user-1",
				Type:    MemoryTypeProfile,
				Content: "用户喜欢自然风景",
			}
			if err := store.Save(t.Context(), mem); err != nil {
				errCh <- err
			}
		})
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("Save() error = %v", err)
	}
	if len(store.memories) != 1 {
		t.Fatalf("Save() stored %d duplicate memories, want 1", len(store.memories))
	}
}

func TestInMemoryStoreSearch(t *testing.T) {
	older := time.Date(2026, time.November, 1, 10, 0, 0, 0, time.UTC)
	newer := older.Add(time.Hour)

	tests := []struct {
		name     string
		memories map[string]Memory
		userID   string
		query    string
		limit    int
		wantIDs  []string
	}{
		{
			name: "匹配任意关键词并优先返回匹配数更多的记忆",
			memories: map[string]Memory{
				"both": {
					ID:        "both",
					UserID:    "user-1",
					Content:   "喜欢自然风景，也喜欢周末旅行",
					UpdatedAt: older,
				},
				"travel": {
					ID:        "travel",
					UserID:    "user-1",
					Content:   "经常安排周末旅行",
					UpdatedAt: newer,
				},
				"other-user": {
					ID:        "other-user",
					UserID:    "user-2",
					Content:   "自然风景 周末旅行",
					UpdatedAt: newer,
				},
			},
			userID:  "user-1",
			query:   "自然风景 周末旅行",
			limit:   10,
			wantIDs: []string{"both", "travel"},
		},
		{
			name: "匹配数相同时按更新时间和 ID 稳定排序",
			memories: map[string]Memory{
				"b": {ID: "b", UserID: "user-1", Content: "自然风景", UpdatedAt: newer},
				"a": {ID: "a", UserID: "user-1", Content: "周末旅行", UpdatedAt: newer},
				"c": {ID: "c", UserID: "user-1", Content: "自然风景", UpdatedAt: older},
			},
			userID:  "user-1",
			query:   "自然风景 周末旅行",
			limit:   2,
			wantIDs: []string{"a", "b"},
		},
		{
			name: "重复关键词只计算一次",
			memories: map[string]Memory{
				"both":   {ID: "both", UserID: "user-1", Content: "自然风景 周末旅行", UpdatedAt: older},
				"nature": {ID: "nature", UserID: "user-1", Content: "自然风景", UpdatedAt: newer},
			},
			userID:  "user-1",
			query:   "自然风景 自然风景 周末旅行",
			limit:   10,
			wantIDs: []string{"both", "nature"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 直接注入固定时间的测试数据，避免 Save 使用当前时间导致排序测试不稳定。
			store := &InMemoryStore{memories: tt.memories}

			got, err := store.Search(t.Context(), tt.userID, tt.query, tt.limit)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}

			if len(got) != len(tt.wantIDs) {
				t.Fatalf("Search() returned %d memories, want %d: %#v", len(got), len(tt.wantIDs), got)
			}
			for i, wantID := range tt.wantIDs {
				if got[i].ID != wantID {
					t.Errorf("Search()[%d].ID = %q, want %q", i, got[i].ID, wantID)
				}
			}
		})
	}
}
