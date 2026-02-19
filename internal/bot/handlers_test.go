package bot

import (
	"context"
	"errors"
	"strings"
	"testing"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jrswab/helpi/internal/llm"
)

type mockRouter struct {
	providerName string
	response     string
	err          error
}

func (m *mockRouter) GetProvider() (llm.Provider, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &mockProvider{name: m.providerName}, nil
}

func (m *mockRouter) SendMessage(ctx context.Context, messages []llm.Message) (string, error) {
	return m.response, m.err
}

type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) IsEnabled() bool {
	return true
}

func (m *mockProvider) SendMessage(ctx context.Context, messages []llm.Message) (string, error) {
	return "", nil
}

type mockSessionManager struct {
	messages []llm.Message
	err      error
}

func (m *mockSessionManager) Get(userID int64) ([]llm.Message, error) {
	return m.messages, m.err
}

func (m *mockSessionManager) Save(userID int64, messages []llm.Message) error {
	return m.err
}

func (m *mockSessionManager) Delete(userID int64) error {
	return m.err
}

type mockBot struct {
	lastMessageParams *tgbot.SendMessageParams
	lastChatAction    *tgbot.SendChatActionParams
}

func (m *mockBot) SendMessage(ctx context.Context, params *tgbot.SendMessageParams) (*models.Message, error) {
	m.lastMessageParams = params
	return nil, nil
}

func (m *mockBot) SendChatAction(ctx context.Context, params *tgbot.SendChatActionParams) (bool, error) {
	m.lastChatAction = params
	return true, nil
}

var _ BotSender = (*mockBot)(nil)

func makeUpdate(userID int64, chatID int64, text string) *models.Update {
	return &models.Update{
		Message: &models.Message{
			From: &models.User{ID: userID},
			Chat: models.Chat{ID: chatID},
			Text: text,
		},
	}
}

func TestStartHandler(t *testing.T) {
	router := &mockRouter{}
	sessionMgr := &mockSessionManager{}
	handlers := NewHandlers(router, sessionMgr, []int64{})

	bot := &mockBot{}
	update := makeUpdate(12345, 12345, "/start")

	handlers.StartHandler(context.Background(), bot, update)

	if bot.lastMessageParams == nil {
		t.Fatal("expected message to be sent")
	}

	expected := "Welcome to Helpi"
	if !strings.Contains(bot.lastMessageParams.Text, expected) {
		t.Errorf("expected message to contain %q, got %q", expected, bot.lastMessageParams.Text)
	}
}

func TestHelpHandler(t *testing.T) {
	router := &mockRouter{}
	sessionMgr := &mockSessionManager{}
	handlers := NewHandlers(router, sessionMgr, []int64{})

	bot := &mockBot{}
	update := makeUpdate(12345, 12345, "/help")

	handlers.HelpHandler(context.Background(), bot, update)

	if bot.lastMessageParams == nil {
		t.Fatal("expected message to be sent")
	}

	expected := "Available commands"
	if !strings.Contains(bot.lastMessageParams.Text, expected) {
		t.Errorf("expected message to contain %q, got %q", expected, bot.lastMessageParams.Text)
	}
}

func TestMyIDHandler(t *testing.T) {
	router := &mockRouter{}
	sessionMgr := &mockSessionManager{}
	handlers := NewHandlers(router, sessionMgr, []int64{})

	userID := int64(98765)
	bot := &mockBot{}
	update := makeUpdate(userID, userID, "/myid")

	handlers.MyIDHandler(context.Background(), bot, update)

	if bot.lastMessageParams == nil {
		t.Fatal("expected message to be sent")
	}

	expectedText := "98765"
	if !contains(bot.lastMessageParams.Text, expectedText) {
		t.Errorf("expected message to contain user ID %q, got %q", expectedText, bot.lastMessageParams.Text)
	}

	if bot.lastMessageParams.ParseMode != models.ParseModeMarkdown {
		t.Errorf("expected parse mode to be Markdown, got %v", bot.lastMessageParams.ParseMode)
	}
}

func TestModelHandler_WithProvider(t *testing.T) {
	router := &mockRouter{providerName: "OpenAI"}
	sessionMgr := &mockSessionManager{}
	handlers := NewHandlers(router, sessionMgr, []int64{})

	bot := &mockBot{}
	update := makeUpdate(12345, 12345, "/model")

	handlers.ModelHandler(context.Background(), bot, update)

	if bot.lastMessageParams == nil {
		t.Fatal("expected message to be sent")
	}

	expected := "Active provider:"
	if !strings.Contains(bot.lastMessageParams.Text, expected) {
		t.Errorf("expected message to contain %q, got %q", expected, bot.lastMessageParams.Text)
	}
}

func TestModelHandler_NoProvider(t *testing.T) {
	router := &mockRouter{err: errors.New("no LLM provider enabled")}
	sessionMgr := &mockSessionManager{}
	handlers := NewHandlers(router, sessionMgr, []int64{})

	bot := &mockBot{}
	update := makeUpdate(12345, 12345, "/model")

	handlers.ModelHandler(context.Background(), bot, update)

	if bot.lastMessageParams == nil {
		t.Fatal("expected message to be sent")
	}

	expected := "Error: No LLM provider enabled"
	if bot.lastMessageParams.Text != expected {
		t.Errorf("expected %q, got %q", expected, bot.lastMessageParams.Text)
	}
}

func TestClearHandler_Success(t *testing.T) {
	router := &mockRouter{}
	sessionMgr := &mockSessionManager{}
	handlers := NewHandlers(router, sessionMgr, []int64{})

	bot := &mockBot{}
	update := makeUpdate(12345, 12345, "/clear")

	handlers.ClearHandler(context.Background(), bot, update)

	if bot.lastMessageParams == nil {
		t.Fatal("expected message to be sent")
	}

	expected := "Conversation history cleared."
	if bot.lastMessageParams.Text != expected {
		t.Errorf("expected %q, got %q", expected, bot.lastMessageParams.Text)
	}
}

func TestClearHandler_Error(t *testing.T) {
	router := &mockRouter{}
	sessionMgr := &mockSessionManager{err: errors.New("delete failed")}
	handlers := NewHandlers(router, sessionMgr, []int64{})

	bot := &mockBot{}
	update := makeUpdate(12345, 12345, "/clear")

	handlers.ClearHandler(context.Background(), bot, update)

	if bot.lastMessageParams == nil {
		t.Fatal("expected message to be sent")
	}

	expected := "Error clearing session: delete failed"
	if bot.lastMessageParams.Text != expected {
		t.Errorf("expected %q, got %q", expected, bot.lastMessageParams.Text)
	}
}

func TestTextMessageHandler_Success(t *testing.T) {
	router := &mockRouter{response: "Hello from AI"}
	sessionMgr := &mockSessionManager{}
	handlers := NewHandlers(router, sessionMgr, []int64{})

	bot := &mockBot{}
	update := makeUpdate(12345, 12345, "Hello")

	handlers.TextMessageHandler(context.Background(), bot, update)

	if bot.lastMessageParams == nil {
		t.Fatal("expected message to be sent")
	}

	expected := "Hello from AI"
	if bot.lastMessageParams.Text != expected {
		t.Errorf("expected %q, got %q", expected, bot.lastMessageParams.Text)
	}
}

func TestTextMessageHandler_NoProviderError(t *testing.T) {
	router := &mockRouter{err: errors.New("no LLM provider enabled")}
	sessionMgr := &mockSessionManager{}
	handlers := NewHandlers(router, sessionMgr, []int64{})

	bot := &mockBot{}
	update := makeUpdate(12345, 12345, "Hello")

	handlers.TextMessageHandler(context.Background(), bot, update)

	if bot.lastMessageParams == nil {
		t.Fatal("expected message to be sent")
	}

	expected := "No LLM provider enabled. Please check configuration."
	if bot.lastMessageParams.Text != expected {
		t.Errorf("expected %q, got %q", expected, bot.lastMessageParams.Text)
	}
}
