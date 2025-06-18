package handlers

import (
	"encoding/json"
	"github.com/DeryabinSergey/germanarticlebot/internal/application/usecases"
	"github.com/DeryabinSergey/germanarticlebot/internal/domain/entities"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/logging"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/tracing"
	"net/http"
	"strings"
)

// ArticleHandler handles HTTP requests for article determination
type ArticleHandler struct {
	useCase *usecases.DetermineArticleUseCase
	logger  logging.Logger
	tracer  tracing.Tracer
}

// NewArticleHandler creates a new article handler
func NewArticleHandler(
	useCase *usecases.DetermineArticleUseCase,
	logger logging.Logger,
	tracer tracing.Tracer,
) *ArticleHandler {
	return &ArticleHandler{
		useCase: useCase,
		logger:  logger,
		tracer:  tracer,
	}
}

// HandleArticleRequest handles HTTP requests for article determination
func (h *ArticleHandler) HandleArticleRequest(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	h.setCORSHeaders(w)

	// Handle preflight OPTIONS request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	spanCtx, span := h.tracer.Start(r.Context(), "HTTP Handler")
	defer span.End()

	// Extract language from Accept-Language header
	language := h.extractLanguageFromHeader(r.Header.Get("Accept-Language"))
	if language == "" {
		h.logger.Warning(spanCtx, map[string]interface{}{
			"message": "No language specified, defaulting to 'en'",
		})
		language = "en"
	}

	var word string
	var err error

	switch r.Method {
	case http.MethodGet:
		if err = r.ParseForm(); err != nil {
			h.logger.Error(spanCtx, map[string]interface{}{
				"message": "Failed to parse form",
				"error":   err.Error(),
			})
			h.writeErrorResponse(w, "Invalid request format", http.StatusBadRequest)
			return
		}
		word = r.Form.Get("word")

	case http.MethodPost:
		var request struct {
			Word string `json:"word"`
		}
		if err = json.NewDecoder(r.Body).Decode(&request); err != nil {
			h.logger.Error(spanCtx, map[string]interface{}{
				"message": "Failed to decode JSON body",
				"error":   err.Error(),
			})
			h.writeErrorResponse(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		word = request.Word

	default:
		h.writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if word == "" {
		h.writeErrorResponse(w, "Word parameter is required", http.StatusBadRequest)
		return
	}

	// Create request entity
	articleRequest := entities.NewArticleRequest(word, language)

	// Execute use case
	response, err := h.useCase.Execute(spanCtx, articleRequest)
	if err != nil {
		h.logger.Error(spanCtx, map[string]interface{}{
			"message": "Use case execution failed",
			"error":   err.Error(),
			"word":    word,
		})
		h.writeErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Write response
	h.writeJSONResponse(w, response, http.StatusOK)
}

// setCORSHeaders sets CORS headers to allow all origins
func (h *ArticleHandler) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept-Language, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
}

func (h *ArticleHandler) extractLanguageFromHeader(acceptLanguage string) string {
	if acceptLanguage == "" {
		return "en"
	}

	// Simple language extraction - take the first language code
	languages := strings.Split(acceptLanguage, ",")
	if len(languages) > 0 {
		lang := strings.TrimSpace(languages[0])
		if strings.Contains(lang, "-") {
			lang = strings.Split(lang, "-")[0]
		}
		if strings.Contains(lang, ";") {
			lang = strings.Split(lang, ";")[0]
		}
		return lang
	}

	return "en"
}

func (h *ArticleHandler) writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *ArticleHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := entities.NewErrorResponse(message)
	h.writeJSONResponse(w, response, statusCode)
}