package promptcompat

import (
	"ds2api/internal/prompt"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var historyTranscriptTitles = []string{
	"# context transcript",
	"# chat history",
	"# conversation log",
	"# session record",
	"# dialog trace",
	"# message archive",
	"# transcript file",
	"# chat output",
	"# history dump",
	"# conversation data",
}

func randomHistoryTranscriptTitle() string {
	return historyTranscriptTitles[rand.Intn(len(historyTranscriptTitles))]
}

// CurrentInputContextFilename is the fallback filename when no template is configured.
const CurrentInputContextFilename = "deepseek.txt"

var historyTranscriptSummaries = []string{
	"Prior conversation history and tool progress.",
	"Conversation context and interaction history.",
	"Previous chat messages and tool interactions.",
	"History of the conversation so far.",
	"Chat transcript and tool usage history.",
	"Context from prior messages and actions.",
	"Record of previous conversation turns.",
	"Full chat history and tool results.",
	"Conversation log with prior interactions.",
	"Historical context from the dialogue.",
	"Previous exchanges and tool calls.",
	"Chat history with tool execution records.",
}

var contextFilenamePrefixes = []string{
	"context",
	"chat",
	"history",
	"session",
	"conversation",
	"dialog",
	"message",
	"transcript",
	"log",
	"notes",
}

var messageSeparatorFormats = []string{
	"=== %d. %s ===",
	"--- %d. %s ---",
	"## %d. %s",
	"**%d. %s**",
	"[%d] %s",
	"%d> %s",
	"{%d} %s",
	"// %d. %s",
}

var roleLabels = map[string][]string{
	"user": {
		"USER",
		"HUMAN",
		"PERSON",
		"CLIENT",
		"REQUESTER",
		"Questioner",
		"Customer",
		"Visitor",
	},
	"assistant": {
		"ASSISTANT",
		"AI",
		"BOT",
		"AGENT",
		"HELPER",
		"RESPONDER",
		"SERVANT",
		"ADVISOR",
	},
	"system": {
		"SYSTEM",
		"INSTRUCTIONS",
		"CONFIG",
		"SETUP",
		"DIRECTIVES",
		"GUIDELINES",
		"RULES",
		"SYSTEM_MSG",
	},
	"tool": {
		"TOOL",
		"FUNCTION",
		"ACTION",
		"EXECUTION",
		"RESULT",
		"RESPONSE",
		"OUTPUT",
		"CALL",
	},
	"unknown": {
		"UNKNOWN",
		"OTHER",
		"UNDEFINED",
		"MISC",
		"UNKNOWN_ROLE",
		"UNLABELED",
		"UNCLASSIFIED",
		"UNKNOWN_MSG",
	},
}

func getRandomRoleLabel(role string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	labels, exists := roleLabels[role]
	if !exists {
		return strings.ToUpper(role)
	}
	return labels[0]
}

func GenerateCurrentInputFilename(template string) string {
	if template == "" {
		return CurrentInputContextFilename
	}
	timestamp := time.Now().Unix()
	last4 := fmt.Sprintf("%04d", timestamp%10000)
	result := strings.ReplaceAll(template, "{time}", last4)
	result = strings.ReplaceAll(result, "{timestamp}", fmt.Sprintf("%d", timestamp))
	result = strings.ReplaceAll(result, "{rand}", fmt.Sprintf("%04d", rand.Intn(10000)))
	result = strings.ReplaceAll(result, "{prefix}", contextFilenamePrefixes[rand.Intn(len(contextFilenamePrefixes))])
	return result
}

func HistoryTranscriptTitle(filename string) string {
	name := strings.TrimSpace(filename)
	if name == "" {
		name = CurrentInputContextFilename
	}
	return "# " + name
}

func BuildOpenAIHistoryTranscript(messages []any) string {
	return buildOpenAIHistoryTranscript(messages)
}

func BuildOpenAICurrentUserInputTranscript(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	return buildOpenAIHistoryTranscript([]any{
		map[string]any{"role": "user", "content": text},
	})
}

func BuildOpenAICurrentInputContextTranscript(messages []any) string {
	return buildOpenAIHistoryTranscriptWithTitle(messages, "")
}

func BuildOpenAICurrentInputContextTranscriptWithFilename(messages []any, filename string) string {
	return buildOpenAIHistoryTranscriptWithTitle(messages, filename)
}

func buildOpenAIHistoryTranscriptWithTitle(messages []any, filename string) string {
	return buildOpenAIHistoryTranscriptImpl(messages, HistoryTranscriptTitle(filename))
}

func buildOpenAIHistoryTranscript(messages []any) string {
	return buildOpenAIHistoryTranscriptImpl(messages, randomHistoryTranscriptTitle())
}

func buildOpenAIHistoryTranscriptImpl(messages []any, title string) string {
	if len(messages) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(historyTranscriptSummaries[0])
	b.WriteString("\n\n")

	separatorFormat := messageSeparatorFormats[0]
	entry := 0
	for _, raw := range messages {
		msg, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		role := normalizeOpenAIRoleForPrompt(strings.ToLower(strings.TrimSpace(asString(msg["role"]))))
		content := strings.TrimSpace(buildOpenAIHistoryEntry(role, msg))
		if content == "" {
			continue
		}
		// Sanitize user-controlled content to prevent separator-format
		// injection that could corrupt the transcript structure.
		if role == "user" || role == "tool" {
			content = prompt.SanitizeUserInput(content)
		}
		entry++
		fmt.Fprintf(&b, separatorFormat+"\n%s\n\n", entry, roleLabelForHistory(role), content)
	}

	transcript := strings.TrimSpace(b.String())
	if transcript == "" {
		return ""
	}
	return transcript + "\n"
}

func buildOpenAIHistoryEntry(role string, msg map[string]any) string {
	switch role {
	case "assistant":
		return strings.TrimSpace(buildAssistantContentForPrompt(msg))
	case "tool", "function":
		return strings.TrimSpace(buildToolHistoryContent(msg))
	case "system", "user":
		return strings.TrimSpace(NormalizeOpenAIContentForPrompt(msg["content"]))
	default:
		return strings.TrimSpace(NormalizeOpenAIContentForPrompt(msg["content"]))
	}
}

func buildToolHistoryContent(msg map[string]any) string {
	content := strings.TrimSpace(NormalizeOpenAIContentForPrompt(msg["content"]))
	parts := make([]string, 0, 2)
	if name := strings.TrimSpace(asString(msg["name"])); name != "" {
		parts = append(parts, "name="+name)
	}
	if callID := strings.TrimSpace(asString(msg["tool_call_id"])); callID != "" {
		parts = append(parts, "tool_call_id="+callID)
	}
	header := ""
	if len(parts) > 0 {
		header = "[" + strings.Join(parts, " ") + "]"
	}
	switch {
	case header != "" && content != "":
		return header + "\n" + content
	case header != "":
		return header
	default:
		return content
	}
}

func roleLabelForHistory(role string) string {
	role = strings.ToLower(strings.TrimSpace(role))
	switch role {
	case "function":
		return getRandomRoleLabel("tool")
	case "":
		return getRandomRoleLabel("unknown")
	default:
		return getRandomRoleLabel(role)
	}
}
