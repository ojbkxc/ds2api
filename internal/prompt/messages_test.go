package prompt

import (
	"strings"
	"testing"
)

func TestNormalizeContentNilReturnsEmpty(t *testing.T) {
	if got := NormalizeContent(nil); got != "" {
		t.Fatalf("expected empty string for nil content, got %q", got)
	}
}

func TestMessagesPrepareNilContentNoNullLiteral(t *testing.T) {
	messages := []map[string]any{
		{"role": "assistant", "content": nil},
		{"role": "user", "content": "ok"},
	}
	got := MessagesPrepare(messages)
	if got == "" {
		t.Fatalf("expected non-empty output")
	}
	if got == "null" {
		t.Fatalf("expected no null literal output, got %q", got)
	}
}

func TestMessagesPrepareUsesTurnSuffixes(t *testing.T) {
	messages := []map[string]any{
		{"role": "system", "content": "System rule"},
		{"role": "user", "content": "Question"},
		{"role": "assistant", "content": "Answer"},
	}
	got := MessagesPrepare(messages)
	if !strings.Contains(got, "System rule") {
		t.Fatalf("expected system instructions to be present, got %q", got)
	}
	if !strings.Contains(got, "Question") {
		t.Fatalf("expected user question to be present, got %q", got)
	}
	if !strings.Contains(got, "Answer") {
		t.Fatalf("expected assistant answer to be present, got %q", got)
	}
	if strings.Contains(got, "<think>") || strings.Contains(got, "</think>") {
		t.Fatalf("did not expect think tags in prompt, got %q", got)
	}
}

func TestMessagesPreparePrependsOutputIntegrityGuard(t *testing.T) {
	messages := []map[string]any{
		{"role": "system", "content": "System rule"},
		{"role": "user", "content": "Question"},
	}
	got := MessagesPrepare(messages)
	// Guard variants no longer have a fixed prefix. Check for any of the
	// guard keywords that appear in all variants.
	hasGuard := strings.Contains(got, "garbled") ||
		strings.Contains(got, "malformed") ||
		strings.Contains(got, "mangled") ||
		strings.Contains(got, "corrupted")
	if !hasGuard {
		t.Fatalf("expected output integrity guard to be prepended, got %q", got)
	}
	if !strings.Contains(got, "System rule") {
		t.Fatalf("expected system rule content to be present, got %q", got)
	}
	if !strings.Contains(got, "Question") {
		t.Fatalf("expected user question to be present, got %q", got)
	}
}

func TestNormalizeContentArrayFallsBackToContentWhenTextEmpty(t *testing.T) {
	got := NormalizeContent([]any{
		map[string]any{"type": "text", "text": "", "content": "from-content"},
	})
	if got != "from-content" {
		t.Fatalf("expected fallback to content when text is empty, got %q", got)
	}
}

func TestMessagesPrepareWithThinkingPreservesPromptShape(t *testing.T) {
	messages := []map[string]any{{"role": "user", "content": "Question"}}
	gotThinking := MessagesPrepareWithThinking(messages, true)
	gotPlain := MessagesPrepareWithThinking(messages, false)
	// Both should contain the DeepSeek native markup format
	if !strings.Contains(gotThinking, "<|User|>Question") {
		t.Fatalf("expected user question in deepseek format, got %q", gotThinking)
	}
	if !strings.Contains(gotPlain, "<|User|>Question") {
		t.Fatalf("expected user question in deepseek format, got %q", gotPlain)
	}
	// Guard should be present (skipGuard=false)
	hasGuard1 := strings.Contains(gotThinking, "garbled") ||
		strings.Contains(gotThinking, "malformed") ||
		strings.Contains(gotThinking, "mangled") ||
		strings.Contains(gotThinking, "corrupted")
	if !hasGuard1 {
		t.Fatalf("expected output integrity guard in thinking mode, got %q", gotThinking)
	}
	hasGuard2 := strings.Contains(gotPlain, "garbled") ||
		strings.Contains(gotPlain, "malformed") ||
		strings.Contains(gotPlain, "mangled") ||
		strings.Contains(gotPlain, "corrupted")
	if !hasGuard2 {
		t.Fatalf("expected output integrity guard in plain mode, got %q", gotPlain)
	}
}
