package llm

import (
	"context"
	"time"

	"github.com/life-journaling/core/internal/domain"
)

// GeminiProvider implements ILLMProvider using Google's Gemini API.
// V1: This is a stub that returns raw text as-is. Full AI parsing is deferred to V2.
type GeminiProvider struct {
	apiKey string
}

// NewGeminiProvider creates a new GeminiProvider.
func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{apiKey: apiKey}
}

// ExtractMemoryData parses raw email text into structured memory data.
// V1 stub: Returns the raw text as content with neutral sentiment.
// V2 will call the Gemini API for intelligent extraction.
func (p *GeminiProvider) ExtractMemoryData(ctx context.Context, rawText string) (domain.Memory, error) {
	return domain.Memory{
		EntryDate: time.Now().UTC(),
		Content:   rawText,
		Sentiment: "neutral",
	}, nil
}
