package memory

import (
	"testing"
	"time"
)

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
