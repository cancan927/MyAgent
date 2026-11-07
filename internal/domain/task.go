package domain

import "time"

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Task represents a user's task.
type Task struct {
	ID        string
	AgentID   string
	Input     string
	Status    TaskStatus
	CreatedAt time.Time
}
