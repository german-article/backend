package entities

// ArticleResponse represents the response with German article information
type ArticleResponse struct {
	Success bool          `json:"success"`
	Error   string        `json:"error,omitempty"`
	Data    []ArticleInfo `json:"data,omitempty"`
}

type ExamplesInfo struct {
	Singular ExampleInfo `json:"singular,omitempty"`
	Plural   ExampleInfo `json:"plural,omitempty"`
}

type ExampleInfo struct {
	Definite   TranslationsInfo `json:"definite,omitempty"`
	Indefinite TranslationsInfo `json:"indefinite,omitempty"`
}

type TranslationsInfo struct {
	NominativeExample     string `json:"nominativeExample,omitempty"`
	NominativeTranslation string `json:"nominativeTranslation,omitempty"`
	AccusativeExample     string `json:"accusativeExample,omitempty"`
	AccusativeTranslation string `json:"accusativeTranslation,omitempty"`
	DativeExample         string `json:"dativeExample,omitempty"`
	DativeTranslation     string `json:"dativeTranslation,omitempty"`
	GenitiveExample       string `json:"genitiveExample,omitempty"`
	GenitiveTranslation   string `json:"genitiveTranslation,omitempty"`
}

// ArticleInfo contains detailed information about a German word with its article
type ArticleInfo struct {
	WordWithArticle string       `json:"wordWithArticle"`
	Translation     string       `json:"translation"`
	Example         ExamplesInfo `json:"example,omitempty"`
}

// NewSuccessResponse creates a successful response
func NewSuccessResponse(data []ArticleInfo) *ArticleResponse {
	return &ArticleResponse{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(err string) *ArticleResponse {
	return &ArticleResponse{
		Success: false,
		Error:   err,
	}
}
