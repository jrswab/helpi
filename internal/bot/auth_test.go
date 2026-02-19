package bot

import (
	"context"
	"testing"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func TestNewAuthMiddleware(t *testing.T) {
	t.Run("creates middleware with users", func(t *testing.T) {
		users := []int64{123, 456}
		m := NewAuthMiddleware(users)
		if m == nil {
			t.Fatal("expected non-nil middleware")
		}
		if len(m.allowedUsers) != 2 {
			t.Errorf("expected 2 users, got %d", len(m.allowedUsers))
		}
	})

	t.Run("creates middleware with empty users", func(t *testing.T) {
		m := NewAuthMiddleware([]int64{})
		if m == nil {
			t.Fatal("expected non-nil middleware")
		}
		if len(m.allowedUsers) != 0 {
			t.Errorf("expected 0 users, got %d", len(m.allowedUsers))
		}
	})
}

func TestAuthMiddleware_Middleware(t *testing.T) {
	t.Run("authorized user calls next handler", func(t *testing.T) {
		allowedUsers := []int64{12345}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{
			Message: &models.Message{
				From: &models.User{ID: 12345},
				Chat: models.Chat{ID: 12345},
				Text: "test",
			},
		}

		wrapped(context.Background(), nil, update)

		if !nextCalled {
			t.Error("expected next handler to be called")
		}
	})

	t.Run("unauthorized user does not call next handler", func(t *testing.T) {
		allowedUsers := []int64{12345}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{}

		wrapped(context.Background(), nil, update)

		if nextCalled {
			t.Error("expected next handler to not be called for unauthorized user")
		}
	})

	t.Run("empty allowedUsers allows all (dev mode)", func(t *testing.T) {
		m := NewAuthMiddleware([]int64{})

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{
			Message: &models.Message{
				From: &models.User{ID: 99999},
				Chat: models.Chat{ID: 99999},
				Text: "test",
			},
		}

		wrapped(context.Background(), nil, update)

		if !nextCalled {
			t.Error("expected next handler to be called in dev mode")
		}
	})

	t.Run("extractUserID from Message", func(t *testing.T) {
		allowedUsers := []int64{54321}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{
			Message: &models.Message{
				From: &models.User{ID: 54321},
				Chat: models.Chat{ID: 54321},
				Text: "test",
			},
		}

		wrapped(context.Background(), nil, update)

		if !nextCalled {
			t.Error("expected next handler to be called for authorized user from Message")
		}
	})

	t.Run("extractUserID from CallbackQuery", func(t *testing.T) {
		allowedUsers := []int64{11111}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{
			CallbackQuery: &models.CallbackQuery{
				From: models.User{ID: 11111},
				ID:   "callback_id",
			},
		}

		wrapped(context.Background(), nil, update)

		if !nextCalled {
			t.Error("expected next handler to be called for authorized user from CallbackQuery")
		}
	})

	t.Run("extractUserID from EditedMessage", func(t *testing.T) {
		allowedUsers := []int64{22222}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{
			EditedMessage: &models.Message{
				From: &models.User{ID: 22222},
				Chat: models.Chat{ID: 22222},
				Text: "edited",
			},
		}

		wrapped(context.Background(), nil, update)

		if !nextCalled {
			t.Error("expected next handler to be called for authorized user from EditedMessage")
		}
	})

	t.Run("getChatID from Message", func(t *testing.T) {
		allowedUsers := []int64{33333}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{
			Message: &models.Message{
				From: &models.User{ID: 33333},
				Chat: models.Chat{ID: 33333},
				Text: "test",
			},
		}

		wrapped(context.Background(), nil, update)

		if !nextCalled {
			t.Error("expected next handler to be called for authorized user with Message chat ID")
		}
	})

	t.Run("getChatID from CallbackQuery with Message", func(t *testing.T) {
		allowedUsers := []int64{44444}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{
			CallbackQuery: &models.CallbackQuery{
				From: models.User{ID: 44444},
				ID:   "callback_id",
				Message: models.MaybeInaccessibleMessage{
					Type: models.MaybeInaccessibleMessageTypeMessage,
					Message: &models.Message{
						ID:   1,
						Chat: models.Chat{ID: 44444},
					},
				},
			},
		}

		wrapped(context.Background(), nil, update)

		if !nextCalled {
			t.Error("expected next handler to be called for authorized user with CallbackQuery.Message")
		}
	})

	t.Run("getChatID from CallbackQuery without Message uses From ID", func(t *testing.T) {
		allowedUsers := []int64{55555}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{
			CallbackQuery: &models.CallbackQuery{
				From: models.User{ID: 55555},
				ID:   "callback_id",
			},
		}

		wrapped(context.Background(), nil, update)

		if !nextCalled {
			t.Error("expected next handler to be called for authorized user with CallbackQuery without Message")
		}
	})

	t.Run("getChatID from EditedMessage", func(t *testing.T) {
		allowedUsers := []int64{66666}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{
			EditedMessage: &models.Message{
				From: &models.User{ID: 66666},
				Chat: models.Chat{ID: 66666},
				Text: "edited",
			},
		}

		wrapped(context.Background(), nil, update)

		if !nextCalled {
			t.Error("expected next handler to be called for authorized user with EditedMessage chat ID")
		}
	})

	t.Run("unauthorized user from CallbackQuery", func(t *testing.T) {
		allowedUsers := []int64{11111}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{}

		wrapped(context.Background(), nil, update)

		if nextCalled {
			t.Error("expected next handler to not be called for unauthorized user from CallbackQuery")
		}
	})

	t.Run("unauthorized user from EditedMessage", func(t *testing.T) {
		allowedUsers := []int64{22222}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{}

		wrapped(context.Background(), nil, update)

		if nextCalled {
			t.Error("expected next handler to not be called for unauthorized user from EditedMessage")
		}
	})

	t.Run("empty update returns false", func(t *testing.T) {
		allowedUsers := []int64{12345}
		m := NewAuthMiddleware(allowedUsers)

		nextCalled := false
		next := func(ctx context.Context, b *bot.Bot, update *models.Update) {
			nextCalled = true
		}

		wrapped := m.Middleware(next)

		update := &models.Update{}

		wrapped(context.Background(), nil, update)

		if nextCalled {
			t.Error("expected next handler to not be called for empty update")
		}
	})
}
