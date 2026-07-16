package promptcompat

import (
	"ds2api/internal/prompt"
	"fmt"
	"math/rand"
	"strings"
)

var thinkingInjectionPrompts = []string{
	// Formal / technical style (original variants)
	"Reasoning Effort: Absolute maximum with no shortcuts permitted.\nYou MUST be very thorough in your thinking and comprehensively decompose the problem to resolve the root cause, rigorously stress-testing your logic against all potential paths, edge cases, and adversarial scenarios.\nExplicitly write out your entire deliberation process, documenting every intermediate step, considered alternative, and rejected hypothesis to ensure absolutely no assumption is left unchecked.",
	"Analytical Depth: Maximum required. Approach this problem systematically, breaking it down into core components and examining each from multiple angles. Document your reasoning process thoroughly, exploring all potential solutions before selecting the optimal path.",
	"Critical Thinking: Apply maximum effort. Carefully analyze the problem, considering all possible approaches and their implications. Document each step of your reasoning to ensure complete transparency and thoroughness.",
	"Deliberation Level: Full analysis required. Think through this problem carefully, considering all edge cases and potential pitfalls. Record your thought process in detail, showing how you arrive at your conclusions.",
	// Mixed conversational / informal style (new)
	"Think step by step. Break down the problem, consider different angles, and explain your reasoning clearly before giving the final answer.",
	"Take your time and think carefully. Walk through your thought process, explore alternatives, and show how you reach your conclusion.",
	"Reason through this thoroughly before responding. Consider edge cases, weigh different approaches, and document your thinking along the way.",
	"Before answering, think this through step by step. What are the key considerations? What could go wrong? Work through it methodically.",
	"Please think carefully about this problem. Analyze it from multiple perspectives, then explain your reasoning and final answer.",
	"Approach this with careful, step-by-step reasoning. Think about the problem deeply, consider alternatives, and show your work.",
	"Spend time thinking through this problem. Break it into parts, examine each one, and explain your reasoning process.",
	"Work through this problem methodically. Consider all angles, think about potential issues, and document your thought process.",
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
		// Formal variants
		"reasoning effort:",
		"analytical depth:",
		"critical thinking:",
		"deliberation level:",
		// Conversational variants (distinctive phrases)
		"walk through your thought process",
		"think this through step by step",
		"break it into parts, examine each",
		"think about the problem deeply",
		"work through this problem methodically",
		"spend time thinking through this",
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
