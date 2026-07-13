package util

import (
	"ds2api/internal/claudeconv"
	"ds2api/internal/config"
	"ds2api/internal/prompt"
	"regexp"
	"strings"
)

var markdownImagePattern = regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)


const ClaudeDefaultModel = "claude-sonnet-4-6"

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

func MessagesPrepare(messages []map[string]any) string {
	if len(messages) == 0 {
		return ""
	}
	if messages[0]["role"] != "system" || !strings.Contains(prompt.NormalizeContent(messages[0]["content"]), "Output integrity guard") {
		messages = append([]map[string]any{{"role": "system", "content": "Output integrity guard"}}, messages...)
	}

	type block struct {
		Role string
		Text string
	}
	processed := make([]block, 0, len(messages))
	for _, m := range messages {
		role, _ := m["role"].(string)
		text := prompt.NormalizeContent(m["content"])
		switch role {
		case "user", "tool":
			text = prompt.SanitizeUserInput(text)
		}
		processed = append(processed, block{Role: role, Text: text})
	}
	if len(processed) == 0 {
		return ""
	}
	merged := make([]block, 0, len(processed))
	for _, msg := range processed {
		if len(merged) > 0 && merged[len(merged)-1].Role == msg.Role {
			merged[len(merged)-1].Text += "\n\n" + msg.Text
			continue
		}
		merged = append(merged, msg)
	}

	var parts []string
	parts = append(parts, "<|begin▁of▁sentence|>")
	for _, m := range merged {
		roleName := m.Role
		switch roleName {
		case "system":
			roleName = "<|System|>"
		case "user":
			roleName = "<|User|>"
		case "assistant":
			roleName = "<|Assistant|>"
		case "tool":
			roleName = "<|Tool|>"
		default:
			roleName = "<|" + strings.ToUpper(roleName) + "|>"
		}
		if m.Role == "tool" {
			parts = append(parts, roleName+m.Text+"<|end▁of▁toolresults|>")
		} else if m.Role == "assistant" {
			parts = append(parts, roleName+m.Text+"<|end▁of▁sentence|>")
		} else {
			parts = append(parts, roleName+m.Text)
		}
	}
	if len(merged) > 0 && merged[len(merged)-1].Role != "assistant" {
		parts = append(parts, "<|Assistant|>")
	}
	out := strings.Join(parts, "")
	return markdownImagePattern.ReplaceAllString(out, `[${1}](${2})`)
}

func normalizeContent(v any) string {
	return prompt.NormalizeContent(v)
}

func ConvertClaudeToDeepSeek(claudeReq map[string]any, store *config.Store) map[string]any {
	return claudeconv.ConvertClaudeToDeepSeek(claudeReq, store, ClaudeDefaultModel)
}

// EstimateTokens provides a rough token count approximation.
// For ASCII text (English, code, etc.) we use ~4 chars per token.
// For non-ASCII text (Chinese, Japanese, Korean, etc.) we use ~1.3 chars per token,
// which better reflects typical BPE tokenizer behavior for CJK scripts.
func EstimateTokens(text string) int {
	if text == "" {
		return 0
	}
	asciiChars := 0
	nonASCIIChars := 0
	for _, r := range text {
		if r < 128 {
			asciiChars++
		} else {
			nonASCIIChars++
		}
	}
	// ASCII: ~4 chars per token; non-ASCII (CJK): ~1.3 chars per token
	n := asciiChars/4 + (nonASCIIChars*10+7)/13
	if n < 1 {
		return 1
	}
	return n
}
