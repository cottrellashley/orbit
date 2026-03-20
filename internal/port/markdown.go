package port

import (
	"context"

	"github.com/cottrellashley/orbit/internal/domain"
)

// MarkdownRenderer converts raw markdown text into HTML. Adapters may
// wrap a full CommonMark/GFM library or provide a deterministic
// fallback that wraps the source in a <pre> block.
type MarkdownRenderer interface {
	// Render converts the markdown source to HTML.
	// The Fallback field of the result indicates whether the output was
	// produced by the deterministic fallback renderer.
	Render(ctx context.Context, source string) (*domain.RenderedMarkdown, error)
}
