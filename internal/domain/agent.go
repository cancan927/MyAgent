package domain

type Agent struct {
	ID           string
	Name         string
	Description  string
	SystemPrompt string

	Model    string
	MaxSteps int
}
