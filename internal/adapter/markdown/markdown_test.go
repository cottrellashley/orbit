package markdown

import (
	"context"
	"testing"
)

func TestFallbackRenderer_Render(t *testing.T) {
	r := NewFallbackRenderer()
	result, err := r.Render(context.Background(), "# Hello World")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != "# Hello World" {
		t.Fatalf("source mismatch: %q", result.Source)
	}
	if !result.Fallback {
		t.Fatal("expected Fallback = true")
	}

	// Verify HTML escaping
	expected := "<pre><code># Hello World</code></pre>"
	if result.HTML != expected {
		t.Fatalf("HTML mismatch:\n  got:  %q\n  want: %q", result.HTML, expected)
	}
}

func TestFallbackRenderer_HTMLEscaping(t *testing.T) {
	r := NewFallbackRenderer()
	result, err := r.Render(context.Background(), `<script>alert("xss")</script>`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HTML == `<pre><code><script>alert("xss")</script></code></pre>` {
		t.Fatal("expected HTML to be escaped, but it wasn't")
	}
	// Verify the script tag is escaped
	expected := `<pre><code>&lt;script&gt;alert(&#34;xss&#34;)&lt;/script&gt;</code></pre>`
	if result.HTML != expected {
		t.Fatalf("HTML mismatch:\n  got:  %q\n  want: %q", result.HTML, expected)
	}
}

func TestFallbackRenderer_EmptySource(t *testing.T) {
	r := NewFallbackRenderer()
	result, err := r.Render(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HTML != "<pre><code></code></pre>" {
		t.Fatalf("unexpected HTML for empty source: %q", result.HTML)
	}
}
