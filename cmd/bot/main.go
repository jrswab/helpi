package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	tgbot "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/jrswab/helpi/internal/bot"
	"github.com/jrswab/helpi/internal/config"
	"github.com/jrswab/helpi/internal/llm"
	"github.com/jrswab/helpi/internal/session"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Telegram.Token == "" {
		log.Fatal("Telegram bot token is required")
	}

	llmRouter, err := llm.NewRouter(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize LLM router: %v", err)
	}

	sessionManager, err := session.NewManager(cfg.Memory.Path, cfg.Memory.MaxMessages)
	if err != nil {
		log.Fatalf("Failed to initialize session manager: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	telegramBot, err := tgbot.New(cfg.Telegram.Token, tgbot.WithDefaultHandler(nil))
	if err != nil {
		log.Fatalf("Failed to create Telegram bot: %v", err)
	}

	handlers := bot.NewHandlers(llmRouter, sessionManager, cfg.AllowedUsers)

	telegramBot.RegisterHandler(tgbot.HandlerTypeMessageText, "/start", tgbot.MatchTypeExact, func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		handlers.StartHandler(ctx, b, update)
	})
	telegramBot.RegisterHandler(tgbot.HandlerTypeMessageText, "/help", tgbot.MatchTypeExact, func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		handlers.HelpHandler(ctx, b, update)
	})
	telegramBot.RegisterHandler(tgbot.HandlerTypeMessageText, "/myid", tgbot.MatchTypeExact, func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		handlers.MyIDHandler(ctx, b, update)
	})
	telegramBot.RegisterHandler(tgbot.HandlerTypeMessageText, "/model", tgbot.MatchTypeExact, func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		handlers.ModelHandler(ctx, b, update)
	})
	telegramBot.RegisterHandler(tgbot.HandlerTypeMessageText, "/clear", tgbot.MatchTypeExact, func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		handlers.ClearHandler(ctx, b, update)
	})
	telegramBot.RegisterHandler(tgbot.HandlerTypeMessageText, "", tgbot.MatchTypeContains, func(ctx context.Context, b *tgbot.Bot, update *models.Update) {
		handlers.TextMessageHandler(ctx, b, update)
	})

	log.Printf("Bot started with token: %s...", maskToken(cfg.Telegram.Token))
	log.Printf("Allowed users count: %d", len(cfg.AllowedUsers))
	if len(cfg.AllowedUsers) == 0 {
		log.Println("WARNING: Development mode - no allowed users configured")
	}

	log.Println("Starting polling...")

	go func() {
		telegramBot.Start(ctx)
	}()

	waitForSignal()
	log.Println("Shutting down bot...")
}

func maskToken(token string) string {
	if len(token) <= 10 {
		return "****"
	}
	return token[:5] + "..." + token[len(token)-5:]
}

func waitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
