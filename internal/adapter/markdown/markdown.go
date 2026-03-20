// Package markdown implements the port.MarkdownRenderer interface.
// The current implementation provides a deterministic fallback renderer
// that wraps the source in an HTML <pre> block. This is intentional —
// a full CommonMark/GFM renderer can be swapped in later by providing
// a new adapter without changing any app or domain code.
package markdown

import (
	"context"
	"html"

	"github.com/cottrellashley/orbit/internal/domain"
)

// FallbackRenderer is a deterministic markdown-to-HTML renderer that
// escapes the source and wraps it in a <pre><code> block. It always
// sets Fallback = true on the result.
type FallbackRenderer struct{}

// NewFallbackRenderer creates a FallbackRenderer.
func NewFallbackRenderer() *FallbackRenderer {
	return &FallbackRenderer{}
}

// Render converts the markdown source to a safe HTML representation.
// The output is always a <pre><code>...</code></pre> block with the
// source HTML-escaped.
func (f *FallbackRenderer) Render(_ context.Context, source string) (*domain.RenderedMarkdown, error) {
	escaped := html.EscapeString(source)
	return &domain.RenderedMarkdown{
		Source:   source,
		HTML:     "<pre><code>" + escaped + "</code></pre>",
		Fallback: true,
	}, nil
}
