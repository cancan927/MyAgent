package memory

import "time"

type MemoryType string

const (
	MemoryTypeProfile  MemoryType = "profile"
	MemoryTypeEpisodic MemoryType = "episodic"
)

type Memory struct {
	ID     string
	UserID string

	Type    MemoryType
	Content string

	CreatedAt time.Time
	UpdatedAt time.Time
}
