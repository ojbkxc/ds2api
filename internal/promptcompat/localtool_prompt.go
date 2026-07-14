package promptcompat

import (
	"ds2api/internal/config"
	"ds2api/internal/localtool"
	"encoding/json"
	"fmt"
	"strings"
)

// localToolPromptTemplates provides randomized phrasing for the tools
// available header, similar to the existing tool prompt system.
var localToolsAvailablePhrases = []string{
	"You have access to these built-in tools:",
	"Available built-in tools:",
	"Built-in tools at your disposal:",
	"You can use the following built-in tools:",
	"The following built-in tools are available:",
}

// webSearchGuidance is injected into the system prompt when local web_search is
// available. It follows the same pattern as deepseek-pp's renderWebSearchGuidance.
const webSearchGuidance = `## Web Search Rules

Use the web_search tool when any of these apply:
- The user asks about real-time information, news, events, exchange rates, weather, or similar
- The user asks about knowledge you are not sure about and current sources are needed
- The user explicitly asks you to search or look something up
- You need to verify facts, data, or cited sources

### Search Flow
1. First output a web_search tool call
2. The search runs automatically; results are sent back to you
3. Read the results, then answer based on them

### Rules
- Use keywords that match the user's language and target sources
- If one search is not enough, call web_search again with different keywords
- Do not invent real-time information without searching`

// BuildLocalToolDescriptors returns the local web tools (web_search, web_fetch)
// as OpenAI-format tool definitions suitable for prompt injection.
func BuildLocalToolDescriptors() []map[string]any {
	descs := localtool.DefaultRegistry.List()
	tools := make([]map[string]any, 0, len(descs))
	for _, desc := range descs {
		// Only include web tools, not memory tools
		if desc.Name != "web_search" && desc.Name != "web_fetch" {
			continue
		}
		params := map[string]any{
			"type":       desc.InputSchema.Type,
			"properties": desc.InputSchema.Properties,
		}
		if len(desc.InputSchema.Required) > 0 {
			params["required"] = desc.InputSchema.Required
		}
		tool := map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        desc.Name,
				"description": desc.Description,
				"parameters":  params,
			},
		}
		tools = append(tools, tool)
	}
	return tools
}

// BuildLocalToolPromptParts generates the tool descriptions and instructions
// for the local web tools, in the same format as buildToolPromptParts in
// tool_prompt.go.
func BuildLocalToolPromptParts() (descriptions string, toolNames []string) {
	descs := localtool.DefaultRegistry.List()
	if len(descs) == 0 {
		return "", nil
	}

	var schemas []string
	for _, desc := range descs {
		if desc.Name != "web_search" && desc.Name != "web_fetch" {
			continue
		}
		name := desc.Name
		descText := desc.Description
		if descText == "" {
			descText = "No description available"
		}
		params := map[string]any{
			"type":       desc.InputSchema.Type,
			"properties": desc.InputSchema.Properties,
		}
		if len(desc.InputSchema.Required) > 0 {
			params["required"] = desc.InputSchema.Required
		}
		b, _ := json.Marshal(params)
		schemas = append(schemas, fmt.Sprintf("Tool: %s\nDescription: %s\nParameters: %s", name, descText, string(b)))
		toolNames = append(toolNames, name)
	}

	if len(schemas) == 0 {
		return "", nil
	}

	phrase := localToolsAvailablePhrases[0]
	descriptions = phrase + "\n\n" + strings.Join(schemas, "\n\n")
	return descriptions, toolNames
}

// BuildLocalToolPrompt returns the complete prompt text for local web tools,
// including tool descriptions, usage instructions, and web search guidance.
func BuildLocalToolPrompt() (promptText string, toolNames []string) {
	descriptions, toolNames := BuildLocalToolPromptParts()
	if descriptions == "" {
		return "", nil
	}
	var b strings.Builder
	b.WriteString(descriptions)
	b.WriteString("\n\n")
	b.WriteString(webSearchGuidance)
	return b.String(), toolNames
}

// InjectLocalToolsIntoPrompt checks if the model supports local web tools and,
// if so, injects the local tool schemas and web search guidance into the
// system prompt. It merges local tools with any client-provided tools.
func InjectLocalToolsIntoPrompt(messages []map[string]any, toolsRaw any, resolvedModel string) ([]map[string]any, []string) {
	if !config.ModelSupportsLocalWebTools(resolvedModel) {
		return messages, nil
	}

	localPrompt, localNames := BuildLocalToolPrompt()
	if localPrompt == "" || len(localNames) == 0 {
		return messages, nil
	}

	// Inject into the system prompt
	for i := range messages {
		if messages[i]["role"] == "system" {
			old, _ := messages[i]["content"].(string)
			messages[i]["content"] = strings.TrimSpace(old + "\n\n" + localPrompt)
			return messages, localNames
		}
	}
	// No system message found, prepend one
	messages = append([]map[string]any{{"role": "system", "content": localPrompt}}, messages...)
	return messages, localNames
}

// MergeLocalToolNames merges local tool names with client-provided tool names.
func MergeLocalToolNames(clientNames []string, resolvedModel string) []string {
	if !config.ModelSupportsLocalWebTools(resolvedModel) {
		return clientNames
	}
	_, localNames := BuildLocalToolPromptParts()
	merged := make([]string, 0, len(clientNames)+len(localNames))
	seen := make(map[string]bool)
	for _, name := range localNames {
		if !seen[name] {
			seen[name] = true
			merged = append(merged, name)
		}
	}
	for _, name := range clientNames {
		if !seen[name] {
			seen[name] = true
			merged = append(merged, name)
		}
	}
	return merged
}

// MergeLocalToolsWithClientTools merges local web tool definitions with
// client-provided tools. Local tools are prepended so they take priority.
func MergeLocalToolsWithClientTools(clientTools []any, resolvedModel string) []any {
	if !config.ModelSupportsLocalWebTools(resolvedModel) {
		return clientTools
	}
	localToolDefs := BuildLocalToolDescriptors()
	merged := make([]any, 0, len(localToolDefs)+len(clientTools))
	for _, t := range localToolDefs {
		merged = append(merged, t)
	}
	merged = append(merged, clientTools...)
	return merged
}