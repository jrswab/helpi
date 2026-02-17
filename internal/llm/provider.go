package llm

import "context"

type Provider interface {
	Name() string
	SendMessage(ctx context.Context, messages []Message) (string, error)
	IsEnabled() bool
}
