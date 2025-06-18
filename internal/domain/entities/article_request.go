package entities

// ArticleRequest represents a request to determine German article
type ArticleRequest struct {
	Word     string
	Language string
}

// NewArticleRequest creates a new article request
func NewArticleRequest(word, language string) *ArticleRequest {
	if language == "" {
		language = "en" // default to English
	}
	return &ArticleRequest{
		Word:     word,
		Language: language,
	}
}

// IsValid checks if the request is valid
func (r *ArticleRequest) IsValid() bool {
	return r.Word != ""
}