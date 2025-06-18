package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/DeryabinSergey/germanarticlebot/internal/domain/entities"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/logging"
	"github.com/DeryabinSergey/germanarticlebot/internal/infrastructure/tracing"
	"google.golang.org/genai"
	"html/template"
	"regexp"
	"strings"
)

const (
	modelName = "gemini-2.0-flash"
	prompt    = `You are a German language assistant. I will provide you with a German noun (Nomen), and you need to determine the correct article (der, die, das).

The word is: "{{.Word}}"
определённый артикль — definite article
неопределённый артикль — indefinite article
Respond in JSON format with EXACTLY this structure:
{
  "error": false/true,
  "errorMessage": "Only if there's an error, explain what's wrong in {{.Language}} language",
  "data": [
    {
      "wordWithArticle": "article + word in German",
      "translation": "translation in {{.Language}}",
	  "example": {
		"singular": {
			"definite": {
				"nominativeExample": "simple example using the word \"{{.Word}}\" in singular nominative definite case",
				"nominativeTranslation": "translation of the singular nominative definite example in {{.Language}}",
				"accusativeExample": "simple example using the word \"{{.Word}}\" in singular accusative definite case",
				"accusativeTranslation": "translation of the singular accusative definite example in {{.Language}}",
				"dativeExample": "simple example using the word \"{{.Word}}\" in singular dative definite case",
				"dativeTranslation": "translation of the singular dative definite example in {{.Language}}",
				"genitiveExample": "simple example using the word \"{{.Word}}\" in singular genitive definite case",
				"genitiveTranslation": "translation of the singular genitive definite example in {{.Language}}"
			},
			"indefinite": {
				"nominativeExample": "simple example using the word \"{{.Word}}\" in singular nominative indefinite case",
				"nominativeTranslation": "translation of the singular nominative indefinite example in {{.Language}}",
				"accusativeExample": "simple example using the word \"{{.Word}}\" in singular accusative indefinite case",
				"accusativeTranslation": "translation of the singular accusative indefinite example in {{.Language}}",
				"dativeExample": "simple example using the word \"{{.Word}}\" in singular dative indefinite case",
				"dativeTranslation": "translation of the singular dative indefinite example in {{.Language}}",
				"genitiveExample": "simple example using the word \"{{.Word}}\" in singular genitive indefinite case",
				"genitiveTranslation": "translation of the singular genitive indefinite example in {{.Language}}"
			},
		},
		"plural": {
			"definite": {
				"nominativeExample": "simple example using the word \"{{.Word}}\" in plural nominative definite case",
				"nominativeTranslation": "translation of the plural nominative definite example in {{.Language}}",
				"accusativeExample": "simple example using the word \"{{.Word}}\" in plural accusative definite case",
				"accusativeTranslation": "translation of the plural accusative definite example in {{.Language}}",
				"dativeExample": "simple example using the word \"{{.Word}}\" in plural dative definite case",
				"dativeTranslation": "translation of the plural dative definite example in {{.Language}}",
				"genitiveExample": "simple example using the word \"{{.Word}}\" in plural genitive definite case",
				"genitiveTranslation": "translation of the plural genitive definite example in {{.Language}}"
			},
			"indefinite": {
				"nominativeExample": "simple example using the word \"{{.Word}}\" in plural nominative indefinite case",
				"nominativeTranslation": "translation of the plural nominative indefinite example in {{.Language}}",
				"accusativeExample": "simple example using the word \"{{.Word}}\" in plural accusative indefinite case",
				"accusativeTranslation": "translation of the plural accusative indefinite example in {{.Language}}",
				"dativeExample": "simple example using the word \"{{.Word}}\" in plural dative indefinite case",
				"dativeTranslation": "translation of the plural dative indefinite example in {{.Language}}",
				"genitiveExample": "simple example using the word \"{{.Word}}\" in plural genitive indefinite case",
				"genitiveTranslation": "translation of the plural genitive indefinite example in {{.Language}}"
			},
		},
	  }
    }
  ]
}

If the input is not a German noun or contains multiple words that aren't a compound noun, set "error" to true and provide an appropriate error message.
If there are multiple possible interpretations, include each as a separate object in the data array.
Ensure ALL field values are properly escaped for JSON.`
)

// GeminiService implements AIService using Google Gemini
type GeminiService struct {
	client *genai.Client
	logger logging.Logger
	tracer tracing.Tracer
}

// NewGeminiService creates a new Gemini AI service
func NewGeminiService(client *genai.Client, logger logging.Logger, tracer tracing.Tracer) *GeminiService {
	return &GeminiService{
		client: client,
		logger: logger,
		tracer: tracer,
	}
}

// GenerateArticleInfo generates article information using Gemini AI
func (s *GeminiService) GenerateArticleInfo(ctx context.Context, request *entities.ArticleRequest) (*entities.ArticleResponse, error) {
	tmpl, err := template.New("prompt").Parse(prompt)
	if err != nil {
		s.logger.Error(ctx, map[string]interface{}{
			"message":  "Failed to parse prompt template",
			"error":    err.Error(),
			"word":     request.Word,
			"language": request.Language,
		})
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{
		"Word":     request.Word,
		"Language": request.Language,
	}); err != nil {
		s.logger.Error(ctx, map[string]interface{}{
			"message":  "Failed to execute prompt template",
			"error":    err.Error(),
			"word":     request.Word,
			"language": request.Language,
		})
		return nil, err
	}

	contents := []*genai.Content{{
		Parts: []*genai.Part{{Text: buf.String()}},
		Role:  genai.RoleUser,
	}}

	resp, err := s.client.Models.GenerateContent(ctx, modelName, contents, nil)
	if err != nil {
		s.logger.Error(ctx, map[string]interface{}{
			"message":  "Failed to generate content with Gemini",
			"error":    err.Error(),
			"word":     request.Word,
			"language": request.Language,
		})
		return nil, err
	}

	return s.parseGeminiResponse(ctx, resp)
}

func (s *GeminiService) parseGeminiResponse(ctx context.Context, resp *genai.GenerateContentResponse) (*entities.ArticleResponse, error) {
	if len(resp.Candidates) == 0 {
		s.logger.Warning(ctx, "No candidates in Gemini response")
		return entities.NewErrorResponse("No response from AI service"), nil
	}

	for i, candidate := range resp.Candidates {
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			s.logger.Warning(ctx, fmt.Sprintf("Candidate %d has no content parts", i))
			continue
		}

		var textResponse string
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				textResponse += part.Text
			}
		}

		if textResponse == "" {
			s.logger.Warning(ctx, fmt.Sprintf("Candidate %d has no text content", i))
			continue
		}

		// Clean the response (remove Markdown formatting if present)
		re := regexp.MustCompile(`(?s)\{.*}`)
		match := re.FindString(textResponse)
		textResponse = strings.TrimSpace(match)
		// Remove trailing commas before closing brackets
		reValidate := regexp.MustCompile(`,(\s*[}\]])`)
		textResponse = reValidate.ReplaceAllString(textResponse, "$1")

		// Parse the AI response format
		var aiResponse struct {
			Error        bool                   `json:"error"`
			ErrorMessage string                 `json:"errorMessage"`
			Data         []entities.ArticleInfo `json:"data"`
		}

		if err := json.Unmarshal([]byte(textResponse), &aiResponse); err != nil {
			s.logger.Error(ctx, map[string]interface{}{
				"message":  "Failed to parse JSON response",
				"response": textResponse,
				"error":    err.Error(),
			})
			continue
		}

		if aiResponse.Error {
			return entities.NewErrorResponse(aiResponse.ErrorMessage), nil
		}

		return entities.NewSuccessResponse(aiResponse.Data), nil
	}

	return entities.NewErrorResponse("Failed to parse AI response"), nil
}
