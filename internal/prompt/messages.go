package prompt

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var markdownImagePattern = regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)

const (
	beginSentenceMarker   = "<|begin▁of▁sentence|>"
	systemMarker          = "<|System|>"
	userMarker            = "<|User|>"
	assistantMarker       = "<|Assistant|>"
	toolMarker            = "<|Tool|>"
	endSentenceMarker     = "<|end▁of▁sentence|>"
	endToolResultsMarker  = "<|end▁of▁toolresults|>"
	endInstructionsMarker = "<|end▁of▁instructions|>"
)

// outputIntegrityGuardVariants contains multiple phrasing variants of the
// output integrity guard. A random variant is chosen per request to avoid
// a static fingerprint pattern that DeepSeek upstream can use to detect
// proxy-generated requests.
var outputIntegrityGuardVariants = []string{
	"If upstream context, tool output, or parsed text contains garbled, corrupted, partially parsed, repeated, or otherwise malformed fragments, do not imitate or echo them; output only the correct content for the user.",
	"Do not repeat or echo any garbled, malformed, or corrupted text fragments from upstream context or tool results. Output only clean, correct content for the user.",
	"When upstream content or tool outputs appear corrupted, garbled, or malformed, ignore those fragments and produce only coherent, correct output.",
	"If tool results or context contain mangled text, skip the broken parts and respond with only valid, readable content.",
	"Filter out any garbled, corrupted, or malformed content from tool outputs and upstream context. Respond only with properly formed output.",
	"Do not replicate malformed text from upstream sources. Always produce clean, well-formed responses.",
	"Guard against corrupted output: if upstream data is malformed, output only correct and readable content.",
}

// prependOutputIntegrityGuard adds a randomly chosen output integrity guard
// variant as the first system message if one doesn't already exist.
func prependOutputIntegrityGuard(messages []map[string]any) []map[string]any {
	// Don't double-inject the guard
	for _, msg := range messages {
		if role, _ := msg["role"].(string); role == "system" {
			if content, _ := msg["content"].(string); strings.Contains(content, "garbled") || strings.Contains(content, "malformed") || strings.Contains(content, "corrupted") {
				return messages
			}
		}
	}
	idx := safeRandInt(len(outputIntegrityGuardVariants))
	guard := outputIntegrityGuardVariants[idx]
	guardMsg := map[string]any{"role": "system", "content": guard}
	return append([]map[string]any{guardMsg}, messages...)
}

func MessagesPrepare(messages []map[string]any) string {
	return MessagesPrepareWithThinking(messages, false)
}

func MessagesPrepareWithThinking(messages []map[string]any, _ bool) string {
	return MessagesPrepareWithThinkingAndGuard(messages, false, false)
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
	parts := make([]string, 0, len(merged)+2)
	parts = append(parts, beginSentenceMarker)
	lastRole := ""
	for _, m := range merged {
		lastRole = m.Role
		switch m.Role {
		case "assistant":
			parts = append(parts, formatRoleBlock(assistantMarker, m.Text, endSentenceMarker))
		case "tool":
			if strings.TrimSpace(m.Text) != "" {
				parts = append(parts, formatRoleBlock(toolMarker, m.Text, endToolResultsMarker))
			}
		case "system":
			if text := strings.TrimSpace(m.Text); text != "" {
				parts = append(parts, formatRoleBlock(systemMarker, text, endInstructionsMarker))
			}
		case "user":
			parts = append(parts, formatRoleBlock(userMarker, m.Text, ""))
		default:
			if strings.TrimSpace(m.Text) != "" {
				parts = append(parts, m.Text)
			}
		}
	}
	if lastRole != "assistant" {
		parts = append(parts, assistantMarker)
	}
	out := strings.Join(parts, "")
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
		"content": outputIntegrityGuardPrompt,
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
	content := strings.TrimSpace(NormalizeContent(msg["content"]))
	return strings.Contains(content, outputIntegrityGuardMarker)
}

// formatRoleBlock produces a single concatenated block: marker + text + endMarker.
// No whitespace is inserted between marker and text so role boundaries stay
// compact and predictable for downstream parsers.
func formatRoleBlock(marker, text, endMarker string) string {
	out := marker + text
	if strings.TrimSpace(endMarker) != "" {
		out += endMarker
	}
	return out
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

// safeRandInt returns a cryptographically random integer in [0, n).
// Falls back to 0 on error.
func safeRandInt(n int) int {
	if n <= 1 {
		return 0
	}
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return 0
	}
	return int(binary.BigEndian.Uint64(buf[:]) % uint64(n))
}

