package tool

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrToolNotFound = errors.New("tool not found")
	ErrToolExists   = errors.New("tool already exists")
)

type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(t Tool) error {
	if t == nil {
		return errors.New("tool is nil")
	}

	def := t.Definition()
	if def.Name == "" {
		return errors.New("tool name is empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[def.Name]; exists {
		return fmt.Errorf("%w: %s", ErrToolExists, def.Name)
	}

	r.tools[def.Name] = t
	return nil
}

func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}

	return t, nil
}

func (r *Registry) Definitions() []Definition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	definitions := make([]Definition, 0, len(r.tools))

	for _, t := range r.tools {
		definitions = append(
			definitions,
			t.Definition(),
		)
	}

	return definitions
}
