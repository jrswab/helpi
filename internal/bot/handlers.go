package bot

import (
	"context"
	"fmt"
	"log"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jrswab/helpi/internal/llm"
	"github.com/jrswab/helpi/internal/session"
)

type BotSender interface {
	SendMessage(ctx context.Context, params *tgbot.SendMessageParams) (*models.Message, error)
	SendChatAction(ctx context.Context, params *tgbot.SendChatActionParams) (bool, error)
}

type botAdapter struct {
	*tgbot.Bot
}

func (a *botAdapter) SendMessage(ctx context.Context, params *tgbot.SendMessageParams) (*models.Message, error) {
	return a.Bot.SendMessage(ctx, params)
}

func (a *botAdapter) SendChatAction(ctx context.Context, params *tgbot.SendChatActionParams) (bool, error) {
	return a.Bot.SendChatAction(ctx, params)
}

type Handlers struct {
	router         llm.Router
	sessionManager session.Manager
	allowedUsers   []int64
}

func NewHandlers(router llm.Router, sessionManager session.Manager, allowedUsers []int64) *Handlers {
	return &Handlers{
		router:         router,
		sessionManager: sessionManager,
		allowedUsers:   allowedUsers,
	}
}

func (h *Handlers) StartHandler(ctx context.Context, b any, update *models.Update) {
	var sender BotSender
	switch v := b.(type) {
	case *tgbot.Bot:
		sender = &botAdapter{Bot: v}
	case BotSender:
		sender = v
	}
	if sender == nil {
		return
	}
	if !h.checkAuth(update) {
		return
	}
	sender.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Welcome to Helpi! I'm here to help you interact with AI models.\n\nAvailable commands:\n/start - Show this welcome message\n/help - Get detailed help\n/myid - Get your Telegram ID\n/model - Show current model info\n/clear - Clear your conversation history\n\nJust send me a message and I'll respond using the configured AI provider.",
	})
}

func (h *Handlers) HelpHandler(ctx context.Context, b any, update *models.Update) {
	var sender BotSender
	switch v := b.(type) {
	case *tgbot.Bot:
		sender = &botAdapter{Bot: v}
	case BotSender:
		sender = v
	}
	if sender == nil {
		return
	}
	if !h.checkAuth(update) {
		return
	}
	sender.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: `Available commands:

/start - Welcome message
/help - Show this help message
/myid - Get your Telegram user ID
/model - Display current active provider and all available providers
/clear - Clear your conversation history

How it works:
- Send me any message and I'll forward it to the AI
- Your conversation history is preserved between messages
- Use /clear to start a fresh conversation`,
	})
}

func (h *Handlers) MyIDHandler(ctx context.Context, b any, update *models.Update) {
	var sender BotSender
	switch v := b.(type) {
	case *tgbot.Bot:
		sender = &botAdapter{Bot: v}
	case BotSender:
		sender = v
	}
	if sender == nil {
		return
	}
	if !h.checkAuth(update) {
		return
	}
	sender.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      fmt.Sprintf("Your Telegram ID: `%d`", update.Message.From.ID),
		ParseMode: models.ParseModeMarkdown,
	})
}

func (h *Handlers) ModelHandler(ctx context.Context, b any, update *models.Update) {
	var sender BotSender
	switch v := b.(type) {
	case *tgbot.Bot:
		sender = &botAdapter{Bot: v}
	case BotSender:
		sender = v
	}
	if sender == nil {
		return
	}
	if !h.checkAuth(update) {
		return
	}
	provider, err := h.router.GetProvider()
	if err != nil {
		sender.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Error: No LLM provider enabled",
		})
		return
	}
	sender.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Active provider: %s", provider.Name()),
	})
}

func (h *Handlers) ClearHandler(ctx context.Context, b any, update *models.Update) {
	var sender BotSender
	switch v := b.(type) {
	case *tgbot.Bot:
		sender = &botAdapter{Bot: v}
	case BotSender:
		sender = v
	}
	if sender == nil {
		return
	}
	if !h.checkAuth(update) {
		return
	}
	userID := update.Message.From.ID
	err := h.sessionManager.Delete(userID)
	if err != nil {
		sender.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Error clearing session: %v", err),
		})
		return
	}
	sender.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Conversation history cleared.",
	})
}

func (h *Handlers) TextMessageHandler(ctx context.Context, b any, update *models.Update) {
	var sender BotSender
	switch v := b.(type) {
	case *tgbot.Bot:
		sender = &botAdapter{Bot: v}
	case BotSender:
		sender = v
	}
	if sender == nil {
		return
	}
	if !h.checkAuth(update) {
		return
	}

	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID

	sender.SendChatAction(ctx, &tgbot.SendChatActionParams{
		ChatID: chatID,
		Action: models.ChatActionTyping,
	})

	messages, err := h.sessionManager.Get(userID)
	if err != nil {
		sender.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   "Error loading conversation history",
		})
		return
	}

	messages = append(messages, llm.Message{
		Role:    "user",
		Content: update.Message.Text,
	})

	response, err := h.router.SendMessage(ctx, messages)
	if err != nil {
		errMsg := "Error communicating with AI"
		if contains(err.Error(), "no LLM provider enabled") {
			errMsg = "No LLM provider enabled. Please check configuration."
		} else if contains(err.Error(), "timeout") || contains(err.Error(), "context deadline") {
			errMsg = "Request timed out. Please try again."
		} else if contains(err.Error(), "context canceled") {
			return
		}
		sender.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   errMsg,
		})
		return
	}

	if response == "" {
		sender.SendMessage(ctx, &tgbot.SendMessageParams{
			ChatID: chatID,
			Text:   "Empty response from AI",
		})
		return
	}

	messages = append(messages, llm.Message{
		Role:    "assistant",
		Content: response,
	})

	if err := h.sessionManager.Save(userID, messages); err != nil {
		log.Printf("Failed to save session for user %d: %v", userID, err)
	}

	sender.SendMessage(ctx, &tgbot.SendMessageParams{
		ChatID: chatID,
		Text:   response,
	})
}

func (h *Handlers) checkAuth(update *models.Update) bool {
	if len(h.allowedUsers) == 0 {
		return true
	}

	var userID int64
	if update.Message != nil {
		userID = update.Message.From.ID
	} else if update.CallbackQuery != nil {
		userID = update.CallbackQuery.From.ID
	} else if update.EditedMessage != nil {
		userID = update.EditedMessage.From.ID
	}

	if userID == 0 {
		log.Printf("[%s] Unauthorized access attempt: missing user info", timestamp())
		return false
	}

	for _, allowed := range h.allowedUsers {
		if userID == allowed {
			return true
		}
	}

	log.Printf("[%s] Unauthorized access attempt from user %d", timestamp(), userID)
	return false
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func timestamp() string {
	return "2024-01-01 00:00:00"
}
