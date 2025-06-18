package telegram

import (
	"context"
	"fmt"
	"github.com/DeryabinSergey/germanarticlebot/internal/application/usecases"
	"github.com/DeryabinSergey/germanarticlebot/internal/domain/entities"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/logging"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/tracing"
	tele "gopkg.in/telebot.v3"
	"strings"
)

// BotHandler handles Telegram bot interactions
type BotHandler struct {
	ctx     context.Context
	bot     *tele.Bot
	useCase *usecases.DetermineArticleUseCase
	logger  logging.Logger
	tracer  tracing.Tracer
}

// NewBotHandler creates a new Telegram bot handler
func NewBotHandler(
	ctx context.Context,
	token string,
	useCase *usecases.DetermineArticleUseCase,
	logger logging.Logger,
	tracer tracing.Tracer,
) (*BotHandler, error) {
	bot, err := tele.NewBot(tele.Settings{
		Token:       token,
		Synchronous: true,
		Poller:      &tele.Webhook{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	handler := &BotHandler{
		ctx:     ctx,
		bot:     bot,
		useCase: useCase,
		logger:  logger,
		tracer:  tracer,
	}

	bot.Use(SetContextMiddleware(handler))
	// Handle /start command
	bot.Handle("/start", handler.handleStart)
	// Handle text messages
	bot.Handle(tele.OnText, handler.handleText)
	return handler, nil
}

func (h *BotHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func SetContextMiddleware(h *BotHandler) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			c.Set("invokeCtx", h.ctx)
			return next(c)
		}
	}
}

// handleStart handles the /start command
func (h *BotHandler) handleStart(c tele.Context) error {
	ctx := c.Get("invokeCtx").(context.Context)
	_, span := h.tracer.Start(ctx, "Telegram Start Command")
	defer span.End()

	welcomeMessage := `🇩🇪 Willkommen! Welcome! Добро пожаловать!

I'm your German Article Bot! Send me any German noun, and I'll help you determine the correct article (der, die, das) along with usage examples.

Just type a German word and I'll provide:
• The correct article
• Translation
• Examples in different grammatical cases

Try sending me a word like "Haus" or "Katze"!`

	return c.Send(welcomeMessage)
}

// handleText handles regular text messages
func (h *BotHandler) handleText(c tele.Context) error {
	ctx := c.Get("invokeCtx").(context.Context)
	spanCtx, span := h.tracer.Start(ctx, "Telegram Text Message")
	defer span.End()

	word := strings.TrimSpace(c.Text())
	if word == "" {
		return c.Send("Please send me a German word to analyze.")
	}

	// Determine user language (simplified - could be enhanced)
	language := h.getUserLanguage(c.Sender())
	// Create request entity
	request := entities.NewArticleRequest(word, language)

	// Execute a use case
	response, err := h.useCase.Execute(spanCtx, request)
	if err != nil {
		return c.Send("Sorry, I encountered an error while processing your request. Please try again.")
	}

	// Format and send response
	message := h.formatResponse(response)
	return c.Send(message, tele.ModeHTML)
}

// getUserLanguage determines user's preferred language
func (h *BotHandler) getUserLanguage(user *tele.User) string {
	if user.LanguageCode != "" {
		return user.LanguageCode
	}

	return "en"
}

// formatResponse formats the article response for Telegram
func (h *BotHandler) formatResponse(response *entities.ArticleResponse) string {
	if !response.Success {
		return fmt.Sprintf("❌ <b>Error:</b> %s", response.Error)
	}

	if len(response.Data) == 0 {
		return "❌ No information found for this word."
	}

	var result strings.Builder
	wr := func(result *strings.Builder, info entities.ExampleInfo) {
		var hasData bool
		if info.Definite != (entities.TranslationsInfo{}) && info.Definite.NominativeExample != "" && info.Definite.NominativeTranslation != "" {
			result.WriteString(fmt.Sprintf("• <b>Nominative Definite:</b> %s / <i>%s</i>\n", info.Definite.NominativeExample, info.Definite.NominativeTranslation))
			hasData = true
		}
		if info.Indefinite != (entities.TranslationsInfo{}) && info.Indefinite.NominativeExample != "" && info.Indefinite.NominativeTranslation != "" {
			result.WriteString(fmt.Sprintf("• <b>Nominative Indefinite:</b> %s / <i>%s</i>\n", info.Indefinite.NominativeExample, info.Indefinite.NominativeTranslation))
			hasData = true
		}
		if hasData {
			result.WriteString("\n")
			hasData = false
		}

		if info.Definite != (entities.TranslationsInfo{}) && info.Definite.AccusativeExample != "" && info.Definite.AccusativeTranslation != "" {
			result.WriteString(fmt.Sprintf("• <b>Accusative Definite:</b> %s / <i>%s</i>\n", info.Definite.AccusativeExample, info.Definite.AccusativeTranslation))
			hasData = true
		}
		if info.Indefinite != (entities.TranslationsInfo{}) && info.Indefinite.AccusativeExample != "" && info.Indefinite.AccusativeTranslation != "" {
			result.WriteString(fmt.Sprintf("• <b>Accusative Indefinite:</b> %s / <i>%s</i>\n", info.Indefinite.AccusativeExample, info.Indefinite.AccusativeTranslation))
			hasData = true
		}
		if hasData {
			result.WriteString("\n")
			hasData = false
		}

		if info.Definite != (entities.TranslationsInfo{}) && info.Definite.DativeExample != "" && info.Definite.DativeTranslation != "" {
			result.WriteString(fmt.Sprintf("• <b>Dative Definite:</b> %s / <i>%s</i>\n", info.Definite.DativeExample, info.Definite.DativeTranslation))
			hasData = true
		}
		if info.Indefinite != (entities.TranslationsInfo{}) && info.Indefinite.DativeExample != "" && info.Indefinite.DativeTranslation != "" {
			result.WriteString(fmt.Sprintf("• <b>Dative Indefinite:</b> %s / <i>%s</i>\n", info.Indefinite.DativeExample, info.Indefinite.DativeTranslation))
			hasData = true
		}
		if hasData {
			result.WriteString("\n")
			hasData = false
		}

		if info.Definite != (entities.TranslationsInfo{}) && info.Definite.GenitiveExample != "" && info.Definite.GenitiveTranslation != "" {
			result.WriteString(fmt.Sprintf("• <b>Genitive Definite:</b> %s / <i>%s</i>\n", info.Definite.GenitiveExample, info.Definite.GenitiveTranslation))
			hasData = true
		}
		if info.Indefinite != (entities.TranslationsInfo{}) && info.Indefinite.GenitiveExample != "" && info.Indefinite.GenitiveTranslation != "" {
			result.WriteString(fmt.Sprintf("• <b>Genitive Indefinite:</b> %s / <i>%s</i>\n", info.Indefinite.GenitiveExample, info.Indefinite.GenitiveTranslation))
			hasData = true
		}
	}
	for i, info := range response.Data {
		if i > 0 {
			result.WriteString("\n\n" + strings.Repeat("─", 10) + "\n\n")
		}

		result.WriteString(fmt.Sprintf("🇩🇪 <b>%s</b>\n", info.WordWithArticle))
		result.WriteString(fmt.Sprintf("📖 <i>%s</i>\n\n", info.Translation))

		var hasData bool
		if (info.Example.Singular != entities.ExampleInfo{}) {
			result.WriteString("📝 <b>Singular Examples:</b>\n")
			wr(&result, info.Example.Singular)
			hasData = true
		}

		if (info.Example.Plural != entities.ExampleInfo{}) {
			if hasData {
				result.WriteString("\n")
			}
			result.WriteString("📝 <b>Plural Examples:</b>\n")
			wr(&result, info.Example.Plural)
			hasData = true
		}
	}

	return result.String()
}

// GetBot returns the underlying bot instance
func (h *BotHandler) GetBot() *tele.Bot {
	return h.bot
}
