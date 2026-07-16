package promptcompat

import (
	"strings"
	"testing"
)

func TestBuildOpenAIFinalPrompt_HandlerPathIncludesToolRoundtripSemantics(t *testing.T) {
	messages := []any{
		map[string]any{"role": "user", "content": "查北京天气"},
		map[string]any{
			"role": "assistant",
			"tool_calls": []any{
				map[string]any{
					"id": "call_1",
					"function": map[string]any{
						"name":      "get_weather",
						"arguments": "{\"city\":\"beijing\"}",
					},
				},
			},
		},
		map[string]any{
			"role":         "tool",
			"tool_call_id": "call_1",
			"name":         "get_weather",
			"content":      map[string]any{"temp": 18, "condition": "sunny"},
		},
	}
	tools := []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "get_weather",
				"description": "Get weather",
				"parameters": map[string]any{
					"type": "object",
				},
			},
		},
	}

	finalPrompt, toolNames := buildOpenAIFinalPrompt(messages, tools, "", false)
	if len(toolNames) != 1 || toolNames[0] != "get_weather" {
		t.Fatalf("unexpected tool names: %#v", toolNames)
	}
	if !strings.Contains(finalPrompt, `"condition":"sunny"`) {
		t.Fatalf("handler finalPrompt should preserve tool output content: %q", finalPrompt)
	}
	if !strings.Contains(finalPrompt, "<|DSML|tool_calls>") {
		t.Fatalf("handler finalPrompt should preserve assistant tool history: %q", finalPrompt)
	}
	if !strings.Contains(finalPrompt, `<|DSML|invoke name="get_weather">`) {
		t.Fatalf("handler finalPrompt should include tool name history: %q", finalPrompt)
	}
}

func TestBuildOpenAIFinalPrompt_VercelPreparePathKeepsFinalAnswerInstruction(t *testing.T) {
	messages := []any{
		map[string]any{"role": "system", "content": "You are helpful"},
		map[string]any{"role": "user", "content": "请调用工具"},
	}
	tools := []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "search",
				"description": "search docs",
				"parameters": map[string]any{
					"type": "object",
				},
			},
		},
	}

	finalPrompt, _ := buildOpenAIFinalPrompt(messages, tools, "", false)
	if !strings.Contains(finalPrompt, "Remember: The ONLY valid way to use tools is the <|DSML|tool_calls>...</|DSML|tool_calls> block at the end of your response.") {
		t.Fatalf("vercel prepare finalPrompt missing final tool-call anchor instruction: %q", finalPrompt)
	}
	if !strings.Contains(finalPrompt, "TOOL CALL FORMAT") {
		t.Fatalf("vercel prepare finalPrompt missing xml format instruction: %q", finalPrompt)
	}
	if !strings.Contains(finalPrompt, "Do NOT wrap in markdown fences") {
		t.Fatalf("vercel prepare finalPrompt missing no-fence xml instruction: %q", finalPrompt)
	}
	if strings.Contains(finalPrompt, "```json") {
		t.Fatalf("vercel prepare finalPrompt should not require fenced tool calls: %q", finalPrompt)
	}
}

func TestBuildOpenAIPromptWithToolInstructionsOnlyOmitsSchemas(t *testing.T) {
	messages := []any{
		map[string]any{"role": "system", "content": "You are helpful"},
		map[string]any{"role": "user", "content": "请调用工具"},
	}
	tools := []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "search",
				"description": "search docs",
				"parameters": map[string]any{
					"type": "object",
				},
			},
		},
	}

	finalPrompt, toolNames := BuildOpenAIPromptWithToolInstructionsOnly(messages, tools, "", DefaultToolChoicePolicy(), false)
	if len(toolNames) != 1 || toolNames[0] != "search" {
		t.Fatalf("unexpected tool names: %#v", toolNames)
	}
	if strings.Contains(finalPrompt, "You have access to these tools") || strings.Contains(finalPrompt, "Description: search docs") || strings.Contains(finalPrompt, "Parameters:") {
		t.Fatalf("tool descriptions should be externalized, got: %q", finalPrompt)
	}
	if !strings.Contains(finalPrompt, "tools.txt") && !strings.Contains(finalPrompt, "tools_") {
		t.Fatalf("expected instructions-only prompt to point model at tools file, got: %q", finalPrompt)
	}
	if !strings.Contains(finalPrompt, "TOOL CALL FORMAT") || !strings.Contains(finalPrompt, "Remember: The ONLY valid way to use tools") {
		t.Fatalf("expected tool format instructions to remain in live prompt, got: %q", finalPrompt)
	}
}

func TestBuildOpenAIToolsContextTranscriptContainsOnlyDescriptions(t *testing.T) {
	tools := []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "search",
				"description": "search docs",
				"parameters": map[string]any{
					"type": "object",
				},
			},
		},
	}

	transcript, toolNames := BuildOpenAIToolsContextTranscript(tools, DefaultToolChoicePolicy())
	if len(toolNames) != 1 || toolNames[0] != "search" {
		t.Fatalf("unexpected tool names: %#v", toolNames)
	}
	for _, want := range []string{"# ", "search", "search docs", `"type":"object"`} {
		if !strings.Contains(transcript, want) {
			t.Fatalf("expected tools transcript to contain %q, got: %q", want, transcript)
		}
	}
	if strings.Contains(transcript, "TOOL CALL FORMAT") || strings.Contains(transcript, "<|DSML|tool_calls>") {
		t.Fatalf("tools transcript should not duplicate format instructions, got: %q", transcript)
	}
}

func TestBuildOpenAIFinalPromptPrependsOutputIntegrityGuard(t *testing.T) {
	messages := []any{
		map[string]any{"role": "system", "content": "You are helpful"},
		map[string]any{"role": "user", "content": "请调用工具"},
	}
	tools := []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "search",
				"description": "search docs",
				"parameters": map[string]any{
					"type": "object",
				},
			},
		},
	}

	finalPrompt, _ := buildOpenAIFinalPrompt(messages, tools, "", false)
	guardIdx := strings.Index(finalPrompt, "garbled")
	if guardIdx < 0 {
		guardIdx = strings.Index(finalPrompt, "malformed")
	}
	if guardIdx < 0 {
		guardIdx = strings.Index(finalPrompt, "mangled")
	}
	if guardIdx < 0 {
		guardIdx = strings.Index(finalPrompt, "corrupted")
	}
	toolIdx := strings.Index(finalPrompt, "TOOL CALL FORMAT")
	if guardIdx < 0 {
		t.Fatalf("expected output integrity guard in final prompt, got: %q", finalPrompt)
	}
	if toolIdx < 0 {
		t.Fatalf("expected tool instructions in final prompt, got: %q", finalPrompt)
	}
	if guardIdx > toolIdx {
		t.Fatalf("expected output integrity guard to precede tool instructions, got: %q", finalPrompt)
	}
}

func TestBuildOpenAIFinalPromptReadLikeToolIncludesCacheGuard(t *testing.T) {
	messages := []any{
		map[string]any{"role": "user", "content": "请读取文件"},
	}
	tools := []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "read_file",
				"description": "Read a file",
				"parameters": map[string]any{
					"type": "object",
				},
			},
		},
	}

	finalPrompt, _ := buildOpenAIFinalPrompt(messages, tools, "", false)
	if !strings.Contains(finalPrompt, "Read-tool cache guard") && !strings.Contains(finalPrompt, "File reading optimization") && !strings.Contains(finalPrompt, "Cache-aware reading") && !strings.Contains(finalPrompt, "Read operation guard") && !strings.Contains(finalPrompt, "Content retrieval safeguard") {
		t.Fatalf("read-like tool prompt missing cache guard: %q", finalPrompt)
	}
	if !strings.Contains(finalPrompt, "no file body") && !strings.Contains(finalPrompt, "not provided") && !strings.Contains(finalPrompt, "unavailable") && !strings.Contains(finalPrompt, "empty") {
		t.Fatalf("read-like tool prompt missing no-body handling: %q", finalPrompt)
	}
	if !strings.Contains(finalPrompt, "Do not repeatedly") && !strings.Contains(finalPrompt, "Do not loop") && !strings.Contains(finalPrompt, "Avoid repeated") && !strings.Contains(finalPrompt, "Skip redundant") {
		t.Fatalf("read-like tool prompt missing loop guard: %q", finalPrompt)
	}
}

func TestBuildOpenAIFinalPromptNonReadToolOmitsCacheGuard(t *testing.T) {
	messages := []any{
		map[string]any{"role": "user", "content": "搜索一下"},
	}
	tools := []any{
		map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        "search",
				"description": "Search docs",
				"parameters": map[string]any{
					"type": "object",
				},
			},
		},
	}

	finalPrompt, _ := buildOpenAIFinalPrompt(messages, tools, "", false)
	if strings.Contains(finalPrompt, "Read-tool cache guard") {
		t.Fatalf("non-read tool prompt should not include read cache guard: %q", finalPrompt)
	}
}

func TestBuildOpenAIFinalPromptWithThinkingKeepsPromptUnchanged(t *testing.T) {
	messages := []any{
		map[string]any{"role": "user", "content": "继续回答上一个问题"},
	}

	finalPromptThinking, _ := buildOpenAIFinalPrompt(messages, nil, "", true)
	finalPromptPlain, _ := buildOpenAIFinalPrompt(messages, nil, "", false)
	// Strip the output integrity guard (randomized per call) from both prompts
	// before comparing. The guard is injected as the first system message.
	stripGuard := func(s string) string {
		// Guard is between <|System|> and <|end▁of▁instructions|> in the first
		// system block. Remove the first system message entirely.
		const marker = "<|end▁of▁instructions|>"
		idx := strings.Index(s, marker)
		if idx < 0 {
			return s
		}
		// Skip the guard: reconstruct without the first system message
		remaining := s[idx+len(marker):]
		return "<|begin▁of▁sentence|><|System|>" + marker + remaining
	}
	thinkingStripped := stripGuard(finalPromptThinking)
	plainStripped := stripGuard(finalPromptPlain)
	if thinkingStripped != plainStripped {
		t.Fatalf("expected thinking flag not to change prompt (aside from guard), thinking=%q plain=%q", thinkingStripped, plainStripped)
	}
}
