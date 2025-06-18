package console

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DeryabinSergey/germanarticlebot/internal/application/usecases"
	"github.com/DeryabinSergey/germanarticlebot/internal/domain/entities"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/logging"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/tracing"
)

// Handler handles console-based interactions for testing
type Handler struct {
	useCase *usecases.DetermineArticleUseCase
	logger  logging.Logger
	tracer  tracing.Tracer
}

// NewConsoleHandler creates a new console handler
func NewConsoleHandler(
	useCase *usecases.DetermineArticleUseCase,
	logger logging.Logger,
	tracer tracing.Tracer,
) *Handler {
	return &Handler{
		useCase: useCase,
		logger:  logger,
		tracer:  tracer,
	}
}

// ProcessRequest processes a console request and returns JSON response
func (h *Handler) ProcessRequest(ctx context.Context, word, language string) (string, error) {
	spanCtx, span := h.tracer.Start(ctx, "ConsoleHandler.ProcessRequest")
	defer span.End()

	h.logger.Info(spanCtx, map[string]interface{}{
		"message":  "Processing console request",
		"word":     word,
		"language": language,
	})

	// Create request entity
	request := entities.NewArticleRequest(word, language)

	// Execute a use case
	response, err := h.useCase.Execute(spanCtx, request)
	if err != nil {
		h.logger.Error(spanCtx, map[string]interface{}{
			"message": "Failed to process console request",
			"error":   err.Error(),
			"word":    word,
		})
		return "", err
	}

	// Convert to JSON
	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %w", err)
	}

	return string(jsonResponse), nil
}
