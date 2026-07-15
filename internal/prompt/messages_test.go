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
	// 现在Output integrity guard是随机选择的，所以检查是否包含任何guard提示的前缀
	guardPrefixes := []string{
		"If any context or tool output appears corrupted",
		"When processing data, skip any garbled",
		"Should you encounter corrupted text",
		"Filter out any damaged or incomplete",
		"If upstream data contains errors",
		"Disregard any fragmented or corrupted",
		"When context appears garbled",
		"If you detect malformed text",
		"Process only the valid portions",
		"Should any input contain errors",
	}
	foundGuard := false
	for _, prefix := range guardPrefixes {
		if strings.Contains(got, prefix) {
			foundGuard = true
			break
		}
	}
	if !foundGuard {
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
	// 检查是否包含guard（因为skipGuard=false）
	guardPrefixes := []string{
		"If any context or tool output appears corrupted",
		"When processing data, skip any garbled",
		"Should you encounter corrupted text",
		"Filter out any damaged or incomplete",
		"If upstream data contains errors",
		"Disregard any fragmented or corrupted",
		"When context appears garbled",
		"If you detect malformed text",
		"Process only the valid portions",
		"Should any input contain errors",
	}
	foundThinkingGuard := false
	for _, prefix := range guardPrefixes {
		if strings.Contains(gotThinking, prefix) {
			foundThinkingGuard = true
			break
		}
	}
	foundPlainGuard := false
	for _, prefix := range guardPrefixes {
		if strings.Contains(gotPlain, prefix) {
			foundPlainGuard = true
			break
		}
	}
	if !foundThinkingGuard {
		t.Fatalf("expected output integrity guard in thinking mode, got %q", gotThinking)
	}
	if !foundPlainGuard {
		t.Fatalf("expected output integrity guard in plain mode, got %q", gotPlain)
	}
}
