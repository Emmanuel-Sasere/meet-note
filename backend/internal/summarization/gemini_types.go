package summarization

// GeminiRequest is the main request structure
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent contains the parts of the request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart can be either text OR inline data (not both)
type GeminiPart struct {
	Text       string      `json:"text,omitempty"`
	InlineData *InlineData `json:"inline_data,omitempty"` // Changed to inline_data
}

// GeminiResponse is the API response structure
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

// GeminiCandidate contains the generated content
type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// InlineData contains base64 encoded data
type InlineData struct {
	MimeType string `json:"mime_type"` // Changed to mime_type
	Data     string `json:"data"`
}