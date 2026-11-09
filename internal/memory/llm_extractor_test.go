package memory

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"MyAgent/internal/domain"
	"MyAgent/internal/llm"
)

type extractorLLM struct {
	generate func(context.Context, *llm.GenerateRequest) (*llm.GenerateResponse, error)
}

func (f extractorLLM) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	return f.generate(ctx, req)
}
func (f extractorLLM) Stream(context.Context, *llm.GenerateRequest) (<-chan llm.StreamEvent, error) {
	panic("unexpected Stream call")
}

func TestLLMExtractorExtract(t *testing.T) {
	upstreamErr := errors.New("model unavailable")
	tests := []struct {
		name, content string
		responseNil   bool
		modelErr      error
		want          []MemoryCandidate
		wantErr       bool
	}{
		{name: "facts", content: `[{"type":"profile","context":"用户喜欢 Go"},{"type":"episodic","context":"用户昨天完成了项目"}]`, want: []MemoryCandidate{{Type: MemoryTypeProfile, Context: "用户喜欢 Go"}, {Type: MemoryTypeEpisodic, Context: "用户昨天完成了项目"}}},
		{name: "no facts", content: `[]`, want: []MemoryCandidate{}},
		{name: "invalid JSON", content: `not json`, wantErr: true},
		{name: "null", content: `null`, wantErr: true},
		{name: "invalid type", content: `[{"type":"unknown","context":"事实"}]`, wantErr: true},
		{name: "empty fact", content: `[{"type":"profile","context":"  "}]`, wantErr: true},
		{name: "wrong field", content: `[{"type":"profile","content":"事实"}]`, wantErr: true},
		{name: "nil response", responseNil: true, wantErr: true},
		{name: "model error", modelErr: upstreamErr, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := []domain.Message{{Role: domain.RoleUser, Content: "我喜欢 Go"}}
			client := extractorLLM{generate: func(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
				if ctx != t.Context() {
					t.Error("context was not preserved")
				}
				if len(req.Messages) != 2 || req.Messages[0].Role != domain.RoleSystem || !reflect.DeepEqual(req.Messages[1:], messages) {
					t.Errorf("unexpected messages: %+v", req.Messages)
				}
				if tt.modelErr != nil {
					return nil, tt.modelErr
				}
				if tt.responseNil {
					return nil, nil
				}
				return &llm.GenerateResponse{Message: domain.Message{Role: domain.RoleAssistant, Content: tt.content}}, nil
			}}
			got, err := NewLLMExtractor(client).Extract(t.Context(), messages)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.modelErr != nil && !errors.Is(err, tt.modelErr) {
				t.Errorf("error does not wrap model error: %v", err)
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestLLMExtractorSkipsGeneration(t *testing.T) {
	client := extractorLLM{generate: func(context.Context, *llm.GenerateRequest) (*llm.GenerateResponse, error) {
		t.Fatal("unexpected model call")
		return nil, nil
	}}
	e := NewLLMExtractor(client)
	got, err := e.Extract(t.Context(), nil)
	if err != nil || len(got) != 0 {
		t.Fatalf("empty input: %v, %v", got, err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = e.Extract(ctx, []domain.Message{{Role: domain.RoleUser, Content: "你好"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v", err)
	}
}
