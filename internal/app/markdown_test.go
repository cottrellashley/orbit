package app

import (
	"context"
	"testing"
)

func TestMarkdownService_Render(t *testing.T) {
	renderer := &mockMarkdownRenderer{fallback: true}
	svc := NewMarkdownService(renderer)

	result, err := svc.Render(context.Background(), "# Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != "# Hello" {
		t.Fatalf("expected source '# Hello', got %q", result.Source)
	}
	if result.HTML == "" {
		t.Fatal("expected non-empty HTML")
	}
	if !result.Fallback {
		t.Fatal("expected fallback to be true")
	}
}

func TestMarkdownService_Render_NonFallback(t *testing.T) {
	renderer := &mockMarkdownRenderer{fallback: false}
	svc := NewMarkdownService(renderer)

	result, err := svc.Render(context.Background(), "some text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Fallback {
		t.Fatal("expected fallback to be false")
	}
}

func TestMarkdownService_IsFallback(t *testing.T) {
	tests := []struct {
		name     string
		fallback bool
		want     bool
	}{
		{"fallback renderer", true, true},
		{"full renderer", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := &mockMarkdownRenderer{fallback: tt.fallback}
			svc := NewMarkdownService(renderer)
			got := svc.IsFallback(context.Background())
			if got != tt.want {
				t.Fatalf("IsFallback() = %v, want %v", got, tt.want)
			}
		})
	}
}
