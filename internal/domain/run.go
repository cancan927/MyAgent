package domain

import "time"

// RunStatus represents an agent execution status.
type RunStatus string

const (
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusCancelled RunStatus = "cancelled"
)

// Run represents one execution of an agent.
type Run struct {
	ID      string
	TaskID  string
	AgentID string

	Status   RunStatus
	Step     int
	Messages []Message

	Result    string
	Error     string
	StartedAt time.Time
	EndedAt   *time.Time
}
