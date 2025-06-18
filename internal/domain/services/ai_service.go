package services

import (
	"context"
	"github.com/DeryabinSergey/germanarticlebot/internal/domain/entities"
)

// AIService defines the interface for AI-powered article determination
type AIService interface {
	GenerateArticleInfo(ctx context.Context, request *entities.ArticleRequest) (*entities.ArticleResponse, error)
}