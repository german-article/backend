package usecases

import (
	"context"
	"github.com/DeryabinSergey/germanarticlebot/internal/domain/entities"
	"github.com/DeryabinSergey/germanarticlebot/internal/domain/services"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/logging"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/tracing"
)

// DetermineArticleUseCase handles the business logic for determining German articles
type DetermineArticleUseCase struct {
	aiService services.AIService
	logger    logging.Logger
	tracer    tracing.Tracer
}

// NewDetermineArticleUseCase creates a new use case instance
func NewDetermineArticleUseCase(
	aiService services.AIService,
	logger logging.Logger,
	tracer tracing.Tracer,
) *DetermineArticleUseCase {
	return &DetermineArticleUseCase{
		aiService: aiService,
		logger:    logger,
		tracer:    tracer,
	}
}

// Execute processes the article determination request
func (uc *DetermineArticleUseCase) Execute(ctx context.Context, request *entities.ArticleRequest) (*entities.ArticleResponse, error) {
	spanCtx, span := uc.tracer.Start(ctx, "Process Article Request")
	defer span.End()

	// Validate request
	if !request.IsValid() {
		uc.logger.Warning(spanCtx, map[string]interface{}{
			"message":  "Invalid article request",
			"word":     request.Word,
			"language": request.Language,
		})
		return entities.NewErrorResponse("Word cannot be empty"), nil
	}

	uc.logger.Info(spanCtx, map[string]interface{}{
		"message":  "Processing article request",
		"word":     request.Word,
		"language": request.Language,
	})

	// Call AI service to determine article
	response, err := uc.aiService.GenerateArticleInfo(spanCtx, request)
	if err != nil {
		return entities.NewErrorResponse("Failed to process request"), err
	}

	return response, nil
}
