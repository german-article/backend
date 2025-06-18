package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/container"
	"gopkg.in/telebot.v3"
	"io"
	"net/http"
	"strings"
)

// Invoke is the main entry point for Google Cloud Functions
func Invoke(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Check if the request context is valid
	if ctx.Err() != nil {
		http.Error(w, "Request context is invalid", http.StatusInternalServerError)
		return
	}

	// Initialize container if not already done
	appContainer, err := container.NewContainer(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to initialize application: %v", err), http.StatusInternalServerError)
		return
	}
	if appContainer == nil {
		http.Error(w, fmt.Sprintf("Failed to initialize application: %v", err), http.StatusInternalServerError)
		return
	}
	defer func(ctx context.Context) {
		_ = appContainer.Tracer.Close(ctx)
		_ = appContainer.Logger.Close(ctx)
	}(ctx)

	spanCtx, span := appContainer.Tracer.Start(ctx, "Application Invoke")
	defer span.End()
	r = r.WithContext(spanCtx)

	// Route based on path and content type
	path := r.URL.Path
	contentType := r.Header.Get("Content-Type")

	switch {
	case path == "/" && r.Method == http.MethodPost && strings.Contains(contentType, "application/json"):
		// Check if it's a Telegram webhook
		body, err := io.ReadAll(r.Body)
		if err != nil {
			appContainer.Logger.Error(spanCtx, map[string]interface{}{
				"message": "Failed to read request body",
				"error":   err.Error(),
			})
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Try to parse as Telegram update
		var telegramUpdate telebot.Update
		if err := json.Unmarshal(body, &telegramUpdate); err == nil && telegramUpdate.ID > 0 && appContainer.TelegramBot != nil {
			appContainer.TelegramBot.SetContext(spanCtx)
			appContainer.TelegramBot.GetBot().ProcessUpdate(telegramUpdate)
			w.WriteHeader(http.StatusOK)
			return
		}

		// If not Telegram, treat as API request
		// Reset body reader
		r.Body = io.NopCloser(strings.NewReader(string(body)))
		appContainer.HTTPHandler.HandleArticleRequest(w, r)

	case path == "/" || path == "/article":
		// Handle API requests
		appContainer.HTTPHandler.HandleArticleRequest(w, r)

	case path == "/health":
		// Health check endpoint
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})

	default:
		http.Error(w, "Not found", http.StatusNotFound)
	}
}
