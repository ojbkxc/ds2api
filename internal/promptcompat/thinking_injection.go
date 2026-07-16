package promptcompat

import (
	"ds2api/internal/prompt"
	"fmt"
	"math/rand"
	"strings"
)

var thinkingInjectionPrompts = []string{
	"Reasoning Effort: Absolute maximum with no shortcuts permitted.\nYou MUST be very thorough in your thinking and comprehensively decompose the problem to resolve the root cause, rigorously stress-testing your logic against all potential paths, edge cases, and adversarial scenarios.\nExplicitly write out your entire deliberation process, documenting every intermediate step, considered alternative, and rejected hypothesis to ensure absolutely no assumption is left unchecked.",
	"Analytical Depth: Maximum required. Approach this problem systematically, breaking it down into core components and examining each from multiple angles. Document your reasoning process thoroughly, exploring all potential solutions before selecting the optimal path.",
	"Critical Thinking: Apply maximum effort. Carefully analyze the problem, considering all possible approaches and their implications. Document each step of your reasoning to ensure complete transparency and thoroughness.",
	"Deliberation Level: Full analysis required. Think through this problem carefully, considering all edge cases and potential pitfalls. Record your thought process in detail, showing how you arrive at your conclusions.",
	"Reasoning Intensity: Highest level. Decompose the problem systematically, exploring all possible solutions and their consequences. Document every step of your analysis to demonstrate complete understanding.",
	"Cognitive Effort: Maximum engagement. Approach this task with thorough analysis, breaking down complex problems into manageable parts. Document your reasoning process to show comprehensive consideration of all factors.",
	"Deep Analysis Mode: Engage maximum reasoning capacity. Break down the problem into fundamental components and examine each thoroughly. Document your complete thought process, including all considerations and decision points.",
	"Comprehensive Reasoning: Apply thorough analysis to this problem. Consider all aspects, potential issues, and alternative approaches. Provide detailed documentation of your reasoning from start to finish.",
}

func GetRandomThinkingInjectionPrompt() string {
	return thinkingInjectionPrompts[rand.Intn(len(thinkingInjectionPrompts))] + fmt.Sprintf(" %04d", rand.Intn(10000))
}

var DefaultThinkingInjectionPrompt = thinkingInjectionPrompts[0]

// ThinkingInjectionMarker is the default thinking injection text used in tests
// to verify that thinking injection has been appended.
var ThinkingInjectionMarker = DefaultThinkingInjectionPrompt

func AppendThinkingInjectionToLatestUser(messages []any) ([]any, bool) {
	return AppendThinkingInjectionPromptToLatestUser(messages, "")
}

func AppendThinkingInjectionPromptToLatestUser(messages []any, injectionPrompt string) ([]any, bool) {
	if len(messages) == 0 {
		return messages, false
	}
	injectionPrompt = strings.TrimSpace(injectionPrompt)
	if injectionPrompt == "" {
		injectionPrompt = GetRandomThinkingInjectionPrompt()
	}
	for i := len(messages) - 1; i >= 0; i-- {
		msg, ok := messages[i].(map[string]any)
		if !ok {
			continue
		}
		if strings.ToLower(strings.TrimSpace(asString(msg["role"]))) != "user" {
			continue
		}
		content := msg["content"]
		normalizedContent := NormalizeOpenAIContentForPrompt(content)
		if containsThinkingInjection(normalizedContent) {
			return messages, false
		}
		updatedContent := appendThinkingInjectionToContent(content, injectionPrompt)
		out := append([]any(nil), messages...)
		cloned := make(map[string]any, len(msg))
		for k, v := range msg {
			cloned[k] = v
		}
		cloned["content"] = updatedContent
		out[i] = cloned
		return out, true
	}
	return messages, false
}

func containsThinkingInjection(content string) bool {
	// Strip zero-width characters first so attackers cannot bypass
	// detection with obfuscation like "Reasoning\u200B Effort:".
	content = prompt.StripZeroWidthChars(content)
	contentLower := strings.ToLower(content)
	markers := []string{
		"reasoning effort:",
		"analytical depth:",
		"critical thinking:",
		"deliberation level:",
		"reasoning intensity:",
		"cognitive effort:",
		"deep analysis mode:",
		"comprehensive reasoning:",
	}
	for _, marker := range markers {
		if strings.Contains(contentLower, marker) {
			return true
		}
	}
	return false
}

func appendThinkingInjectionToContent(content any, injectionPrompt string) any {
	switch x := content.(type) {
	case string:
		return appendTextBlock(x, injectionPrompt)
	case []any:
		out := append([]any(nil), x...)
		out = append(out, map[string]any{
			"type": "text",
			"text": injectionPrompt,
		})
		return out
	default:
		text := NormalizeOpenAIContentForPrompt(content)
		return appendTextBlock(text, injectionPrompt)
	}
}

func appendTextBlock(base, addition string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return addition
	}
	return base + "\n\n" + addition
}
