package prompt

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

var markdownImagePattern = regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)

var outputIntegrityGuardPrompts = []string{
	"If any context or tool output appears corrupted or malformed, ignore it and provide only accurate, clean responses to the user.",
	"When processing data, skip any garbled or nonsensical fragments. Focus on delivering clear, coherent information.",
	"Should you encounter corrupted text or malformed content in the context, disregard it entirely and respond with correct information only.",
	"Filter out any damaged or incomplete content from your processing. Your responses should contain only well-formed, meaningful text.",
	"If upstream data contains errors or corruption, ignore those parts and provide the user with clean, accurate output.",
	"Disregard any fragmented or corrupted content in the input. Ensure your response is clear and properly formatted.",
	"When context appears garbled or nonsensical, skip those sections and deliver only verified, correct information.",
	"If you detect malformed text or corrupted fragments, exclude them from your response and output only clean content.",
	"Process only the valid portions of any context. Ignore corrupted or damaged text and respond with accurate information.",
	"Should any input contain errors or corruption, filter it out and provide the user with coherent, well-formed responses.",
}

func randomOutputIntegrityGuard() string {
	return outputIntegrityGuardPrompts[rand.Intn(len(outputIntegrityGuardPrompts))]
}

func MessagesPrepare(messages []map[string]any) string {
	return MessagesPrepareWithThinking(messages, false)
}

func MessagesPrepareWithThinking(messages []map[string]any, thinkingEnabled bool) string {
	return MessagesPrepareWithThinkingAndGuard(messages, thinkingEnabled, false)
}

func MessagesPrepareWithThinkingAndGuard(messages []map[string]any, _ bool, skipGuard bool) string {
	if !skipGuard {
		messages = prependOutputIntegrityGuard(messages)
	}

	type block struct {
		Role string
		Text string
	}
	processed := make([]block, 0, len(messages))
	for _, m := range messages {
		role, _ := m["role"].(string)
		text := NormalizeContent(m["content"])
		// Sanitize user-controlled input to prevent role-prefix injection
		// and format-injection attacks before the prompt is assembled.
		switch role {
		case "user", "tool":
			text = SanitizeUserInput(text)
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
	// 常规格式: 角色名: 内容，多个消息之间用两个换行分隔
	var parts []string
	for i, m := range merged {
		roleName := m.Role
		// 统一角色名格式
		switch roleName {
		case "system":
			roleName = "System"
		case "user":
			roleName = "User"
		case "assistant":
			roleName = "Assistant"
		case "tool":
			roleName = "Tool"
		}
		parts = append(parts, roleName+": "+m.Text)
		if i < len(merged)-1 {
			parts = append(parts, "")
		}
	}
	out := strings.Join(parts, "\n")
	return markdownImagePattern.ReplaceAllString(out, `[${1}](${2})`)
}

func prependOutputIntegrityGuard(messages []map[string]any) []map[string]any {
	if len(messages) == 0 {
		return messages
	}
	if hasOutputIntegrityGuard(messages[0]) {
		return messages
	}
	out := make([]map[string]any, 0, len(messages)+1)
	out = append(out, map[string]any{
		"role":    "system",
		"content": randomOutputIntegrityGuard(),
	})
	out = append(out, messages...)
	return out
}

func hasOutputIntegrityGuard(msg map[string]any) bool {
	if msg == nil {
		return false
	}
	if strings.ToLower(strings.TrimSpace(asString(msg["role"]))) != "system" {
		return false
	}
	content := strings.ToLower(strings.TrimSpace(NormalizeContent(msg["content"])))
	for _, guard := range outputIntegrityGuardPrompts {
		if strings.Contains(content, strings.ToLower(guard[:20])) {
			return true
		}
	}
	return false
}

func NormalizeContent(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case []any:
		parts := make([]string, 0, len(x))
		for _, item := range x {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			typeStr, _ := m["type"].(string)
			typeStr = strings.ToLower(strings.TrimSpace(typeStr))
			if typeStr == "text" || typeStr == "output_text" || typeStr == "input_text" {
				if txt, ok := m["text"].(string); ok && txt != "" {
					parts = append(parts, txt)
					continue
				}
				if txt, ok := m["content"].(string); ok && txt != "" {
					parts = append(parts, txt)
				}
			}
		}
		return strings.Join(parts, "\n")
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
}
