package services

import (
	"context"
	"github.com/DeryabinSergey/germanarticlebot/internal/domain/entities"
)

// ArticleService defines the interface for article determination service
type ArticleService interface {
	DetermineArticle(ctx context.Context, request *entities.ArticleRequest) (*entities.ArticleResponse, error)
}