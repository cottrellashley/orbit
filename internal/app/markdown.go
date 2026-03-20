package app

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
)

// markdownRenderer is the consumer-defined interface for the markdown port.
type markdownRenderer interface {
	Render(ctx context.Context, source string) (*domain.RenderedMarkdown, error)
}

// MarkdownService provides app-facing markdown rendering. It delegates
// to a MarkdownRenderer port adapter — currently the fallback renderer,
// but a full GFM renderer can be swapped in later.
type MarkdownService struct {
	renderer markdownRenderer
}

// NewMarkdownService creates a MarkdownService.
func NewMarkdownService(r markdownRenderer) *MarkdownService {
	return &MarkdownService{renderer: r}
}

// Render converts markdown source to HTML using the configured renderer.
func (s *MarkdownService) Render(ctx context.Context, source string) (*domain.RenderedMarkdown, error) {
	return s.renderer.Render(ctx, source)
}

// IsFallback returns true if the configured renderer is the
// deterministic fallback. Useful for driver UI hints.
func (s *MarkdownService) IsFallback(ctx context.Context) bool {
	result, err := s.renderer.Render(ctx, "test")
	if err != nil {
		return true
	}
	return result.Fallback
}
