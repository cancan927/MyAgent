package agent

import "errors"

var (
	ErrMaxStepsExceeded = errors.New("agent max steps exceeded")
	ErrEmptyResponse    = errors.New("llm returned empty response")
)
