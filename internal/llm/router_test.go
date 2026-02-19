package llm

import (
	"context"
	"errors"
	"testing"
)

type mockProvider struct {
	name     string
	enabled  bool
	response string
	err      error
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) IsEnabled() bool { return m.enabled }

func (m *mockProvider) SendMessage(ctx context.Context, messages []Message) (string, error) {
	return m.response, m.err
}

func TestGetProvider(t *testing.T) {
	tests := []struct {
		name           string
		providers      []Provider
		defaultIdx     int
		expectedName   string
		expectedErrMsg string
	}{
		{
			name: "default provider enabled",
			providers: []Provider{
				&mockProvider{name: "openai", enabled: false},
				&mockProvider{name: "anthropic", enabled: true},
			},
			defaultIdx:   1,
			expectedName: "anthropic",
		},
		{
			name: "default disabled fallback to first enabled",
			providers: []Provider{
				&mockProvider{name: "openai", enabled: true},
				&mockProvider{name: "anthropic", enabled: false},
			},
			defaultIdx:   1,
			expectedName: "openai",
		},
		{
			name: "no providers enabled",
			providers: []Provider{
				&mockProvider{name: "openai", enabled: false},
				&mockProvider{name: "anthropic", enabled: false},
			},
			defaultIdx:     0,
			expectedErrMsg: "no LLM provider enabled",
		},
		{
			name: "default disabled fallback first in list",
			providers: []Provider{
				&mockProvider{name: "openai", enabled: true},
				&mockProvider{name: "anthropic", enabled: true},
			},
			defaultIdx:   -1,
			expectedName: "openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRouter(tt.providers, tt.defaultIdx)
			provider, err := r.GetProvider()

			if tt.expectedErrMsg != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedErrMsg)
				} else if err.Error() != tt.expectedErrMsg {
					t.Errorf("expected error %q, got %q", tt.expectedErrMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if provider.Name() != tt.expectedName {
				t.Errorf("expected provider %q, got %q", tt.expectedName, provider.Name())
			}
		})
	}
}

func TestSendMessage(t *testing.T) {
	tests := []struct {
		name           string
		providers      []Provider
		defaultIdx     int
		messages       []Message
		expectedResp   string
		expectedErrMsg string
	}{
		{
			name: "valid provider returns response",
			providers: []Provider{
				&mockProvider{name: "openai", enabled: true, response: "hello world"},
			},
			defaultIdx:   0,
			messages:     []Message{{Role: "user", Content: "hi"}},
			expectedResp: "hello world",
		},
		{
			name: "no provider returns error",
			providers: []Provider{
				&mockProvider{name: "openai", enabled: false},
			},
			defaultIdx:     0,
			messages:       []Message{{Role: "user", Content: "hi"}},
			expectedErrMsg: "no LLM provider enabled",
		},
		{
			name: "provider returns error",
			providers: []Provider{
				&mockProvider{name: "openai", enabled: true, err: errors.New("provider error")},
			},
			defaultIdx:     0,
			messages:       []Message{{Role: "user", Content: "hi"}},
			expectedErrMsg: "provider error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRouter(tt.providers, tt.defaultIdx)
			resp, err := r.SendMessage(context.Background(), tt.messages)

			if tt.expectedErrMsg != "" {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.expectedErrMsg)
				} else if err.Error() != tt.expectedErrMsg {
					t.Errorf("expected error %q, got %q", tt.expectedErrMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp != tt.expectedResp {
				t.Errorf("expected response %q, got %q", tt.expectedResp, resp)
			}
		})
	}
}
