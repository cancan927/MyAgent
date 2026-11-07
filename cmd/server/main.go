package main

import (
	"MyAgent/internal/memory"
	"context"
	"fmt"
)

func main() {

	ctx := context.Background()

	store := memory.NewInMemoryStore()

	_ = store.Save(ctx, &memory.Memory{
		ID:      "mem_001",
		UserID:  "user_001",
		Type:    memory.MemoryTypeProfile,
		Content: "用户喜欢自然风景。",
	})

	_ = store.Save(ctx, &memory.Memory{
		ID:      "mem_002",
		UserID:  "user_001",
		Type:    memory.MemoryTypeProfile,
		Content: "用户是一名 Go 后端工程师。",
	})

	_ = store.Save(ctx, &memory.Memory{
		ID:      "mem_003",
		UserID:  "user_002",
		Type:    memory.MemoryTypeProfile,
		Content: "用户喜欢城市旅游。",
	})

	results, err := store.Search(
		ctx,
		"user_001",
		"自然风景",
		5,
	)
	if err != nil {
		panic(err)
	}

	for _, mem := range results {
		fmt.Println(mem.Content)
	}
}
