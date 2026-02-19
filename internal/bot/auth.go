package bot

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type AuthMiddleware struct {
	allowedUsers []int64
}

func NewAuthMiddleware(allowedUsers []int64) *AuthMiddleware {
	if len(allowedUsers) == 0 {
		log.Println("WARNING: Development mode - no allowed users configured")
	}
	return &AuthMiddleware{
		allowedUsers: allowedUsers,
	}
}

func (m *AuthMiddleware) Middleware(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if !m.isAuthorized(update) {
			chatID := m.getChatID(update)
			if chatID != 0 {
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: chatID,
					Text:   "Access denied. You are not authorized to use this bot.",
				})
			}
			return
		}
		next(ctx, b, update)
	}
}

func (m *AuthMiddleware) isAuthorized(update *models.Update) bool {
	if len(m.allowedUsers) == 0 {
		return true
	}

	userID := m.extractUserID(update)
	if userID == 0 {
		log.Printf("[AUTH] Unauthorized access attempt: missing user information")
		return false
	}

	for _, allowed := range m.allowedUsers {
		if userID == allowed {
			return true
		}
	}

	log.Printf("[AUTH] Unauthorized access attempt from user %d", userID)
	return false
}

func (m *AuthMiddleware) extractUserID(update *models.Update) int64 {
	if update.Message != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	if update.EditedMessage != nil {
		return update.EditedMessage.From.ID
	}
	return 0
}

func (m *AuthMiddleware) getChatID(update *models.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil {
		if update.CallbackQuery.Message.Message != nil {
			return update.CallbackQuery.Message.Message.Chat.ID
		}
		return update.CallbackQuery.From.ID
	}
	if update.EditedMessage != nil {
		return update.EditedMessage.Chat.ID
	}
	return 0
}
