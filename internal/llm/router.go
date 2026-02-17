package llm

import (
	"context"
	"fmt"
)

type Router interface {
	GetProvider() (Provider, error)
	SendMessage(ctx context.Context, messages []Message) (string, error)
}

type router struct {
	providers  []Provider
	defaultIdx int
}

func newRouter(providers []Provider, defaultIdx int) Router {
	return &router{
		providers:  providers,
		defaultIdx: defaultIdx,
	}
}

func (r *router) GetProvider() (Provider, error) {
	if r.defaultIdx >= 0 && r.defaultIdx < len(r.providers) {
		provider := r.providers[r.defaultIdx]
		if provider.IsEnabled() {
			return provider, nil
		}
	}

	for _, p := range r.providers {
		if p.IsEnabled() {
			return p, nil
		}
	}

	return nil, fmt.Errorf("no LLM provider enabled")
}

func (r *router) SendMessage(ctx context.Context, messages []Message) (string, error) {
	provider, err := r.GetProvider()
	if err != nil {
		return "", err
	}

	return provider.SendMessage(ctx, messages)
}
