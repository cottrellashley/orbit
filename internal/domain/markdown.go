package domain

// RenderedMarkdown holds the result of rendering a markdown source string.
type RenderedMarkdown struct {
	// Source is the original markdown text that was rendered.
	Source string
	// HTML is the rendered HTML output.
	HTML string
	// Fallback indicates the output was produced by the deterministic
	// fallback renderer rather than a full-featured implementation.
	Fallback bool
}
