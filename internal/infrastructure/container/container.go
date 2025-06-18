package container

import (
	"context"
	"fmt"
	"github.com/DeryabinSergey/germanarticlebot/internal/adapters/console"
	"github.com/DeryabinSergey/germanarticlebot/internal/adapters/http/handlers"
	"github.com/DeryabinSergey/germanarticlebot/internal/adapters/telegram"
	"github.com/DeryabinSergey/germanarticlebot/internal/application/usecases"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/ai"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/config"
	"github.com/DeryabinSergey/germanarticlebot/libs/logger"
	"github.com/DeryabinSergey/germanarticlebot/libs/tracer"
	"google.golang.org/genai"
)

// Container holds all application dependencies
type Container struct {
	Config         *config.Config
	Logger         *logger.Log
	Tracer         *tracer.Tracer
	GeminiClient   *genai.Client
	AIService      *ai.GeminiService
	UseCase        *usecases.DetermineArticleUseCase
	HTTPHandler    *handlers.ArticleHandler
	TelegramBot    *telegram.BotHandler
	ConsoleHandler *console.Handler
}

// NewContainer creates and initializes the dependency injection container
func NewContainer(ctx context.Context) (*Container, error) {
	cfg := config.LoadConfig()

	// Initialize logger
	l, err := logger.Init(ctx, cfg.ProjectID, cfg.ApplicationName, cfg.GCPEnabled, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize tracer
	tr, err := tracer.Init(ctx, cfg.ProjectID, cfg.ApplicationName, cfg.GCPEnabled)
	if err != nil {
		l.Critical(ctx, map[string]interface{}{
			"message": "failed to initialize tracer",
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	// Initialize Gemini client
	geminiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
		Backend:     genai.BackendVertexAI,
		Project:     cfg.ProjectID,
		Location:    "global",
	})
	if err != nil {
		l.Critical(ctx, map[string]interface{}{
			"message": "failed to create Gemini client",
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Initialize services
	aiService := ai.NewGeminiService(geminiClient, l, tr)
	useCase := usecases.NewDetermineArticleUseCase(aiService, l, tr)

	// Initialize handlers
	httpHandler := handlers.NewArticleHandler(useCase, l, tr)
	consoleHandler := console.NewConsoleHandler(useCase, l, tr)

	// Initialize Telegram bot (only if token is provided)
	var telegramBot *telegram.BotHandler
	if cfg.TelegramToken != "" {
		telegramBot, err = telegram.NewBotHandler(ctx, cfg.TelegramToken, useCase, l, tr)
		if err != nil {
			l.Error(ctx, map[string]interface{}{
				"message": "failed to initialize Telegram bot",
				"error":   err.Error(),
			})
			// Don't fail completely if Telegram bot fails to initialize
		}
	}

	return &Container{
		Config:         cfg,
		Logger:         l,
		Tracer:         tr,
		GeminiClient:   geminiClient,
		AIService:      aiService,
		UseCase:        useCase,
		HTTPHandler:    httpHandler,
		TelegramBot:    telegramBot,
		ConsoleHandler: consoleHandler,
	}, nil
}
