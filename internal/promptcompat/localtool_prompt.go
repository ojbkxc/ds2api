package promptcompat

import (
	"ds2api/internal/config"
	"ds2api/internal/localtool"
	"ds2api/internal/toolcall"
	"crypto/rand"
	"encoding/binary"
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

// webFetchGuidance is injected into the system prompt when local web_fetch is
// available. It tells the model when to use web_fetch instead of web_search.
// CRITICAL: without this guidance, the model sees web_fetch in the tool list
// but does not know when to use it (webSearchGuidance only covers web_search).
const webFetchGuidance = `## Web Fetch Rules

Use the web_fetch tool when any of these apply:
- The user asks you to access, fetch, visit, or read content from a specific URL
- The user provides a link and wants you to retrieve or extract its content
- The user asks you to open a webpage, check a link, or get information from a web address
- The user asks you to summarize or analyze content from a specific web page

### Fetch Flow
1. First output a web_fetch tool call with the target URL
2. The fetch runs automatically; the page content is sent back to you
3. Read the content, then answer the user based on what you retrieved

### Rules
- Always include the full URL with the https:// prefix
- If the fetch fails, explain the error to the user and suggest alternatives
- Use web_fetch for direct URL access, not web_search
- If the user asks you to "access", "visit", "open", "read", or "fetch" a URL, use web_fetch`

// BuildLocalToolDescriptors returns all local tools (web_search, web_fetch, MCP tools)
// as OpenAI-format tool definitions suitable for prompt injection.
func BuildLocalToolDescriptors() []map[string]any {
	descs := localtool.DefaultRegistry.List()
	tools := make([]map[string]any, 0, len(descs))
	for _, desc := range descs {
		// Include all local tools: web_search, web_fetch, and MCP tools
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
// for all local tools (web_search, web_fetch, MCP tools),
// in the same format as buildToolPromptParts in tool_prompt.go.
//
// When skipWebSearch is true, the web_search tool is excluded. This is used
// for models that have native search (e.g. deepseek-v4-flash-search with
// search=true) to avoid conflicting with the model's built-in search.
func BuildLocalToolPromptParts(skipWebSearch bool) (descriptions string, toolNames []string) {
	descs := localtool.DefaultRegistry.List()
	if len(descs) == 0 {
		return "", nil
	}

	var schemas []string
	for _, desc := range descs {
		// Include all tools: web_search, web_fetch, and MCP tools
		// Skip internal memory tools (they are not exposed to the LLM)
		// Skip web_search when the model has native search to avoid confusion
		if strings.HasPrefix(desc.Name, "memory") {
			continue
		}
		if skipWebSearch && desc.Name == "web_search" {
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

	// Randomize the "available tools" header to reduce fingerprinting
	idx := safeRandInt(len(localToolsAvailablePhrases))
	phrase := localToolsAvailablePhrases[idx]
	descriptions = phrase + "\n\n" + strings.Join(schemas, "\n\n")
	return descriptions, toolNames
}

// BuildLocalToolPrompt returns the complete prompt text for local web tools,
// including tool descriptions, DSML format instructions, web search guidance,
// and web fetch guidance.
//
// When skipWebSearch is true, the web_search tool is excluded from descriptions
// and the web search guidance is omitted. The web_fetch tool is always included
// because native search does not handle direct URL fetching.
func BuildLocalToolPrompt(skipWebSearch bool) (promptText string, toolNames []string) {
	descriptions, toolNames := BuildLocalToolPromptParts(skipWebSearch)
	if descriptions == "" {
		return "", nil
	}

	// Generate DSML tool call format instructions for local tools.
	// CRITICAL: without these instructions, the model sees JSON schema
	// descriptions and outputs JSON tool calls (e.g. {"tool":"web_fetch",...}),
	// but the tool sieve only recognizes XML/DSML format
	// (<|DSML|tool_calls> wrapper). This causes tool calls to be silently
	// ignored and treated as plain text.
	instructions := toolcall.BuildToolCallInstructions(toolNames)

	var b strings.Builder
	b.WriteString(descriptions)
	b.WriteString("\n\n")
	b.WriteString(instructions)
	if !skipWebSearch {
		b.WriteString("\n\n")
		b.WriteString(webSearchGuidance)
	}
	b.WriteString("\n\n")
	b.WriteString(webFetchGuidance)
	return b.String(), toolNames
}

// InjectLocalToolsIntoPrompt checks if the model supports local web tools and,
// if so, injects the local tool schemas and web search guidance into the
// system prompt. It merges local tools with any client-provided tools.
//
// For models with native search (search=true, e.g. deepseek-v4-flash-search),
// the local web_search tool is skipped to avoid conflicting with the model's
// built-in search. The web_fetch tool is always injected since native search
// does not handle direct URL fetching.
//
// For vision models, only web_fetch is injected (no web_search).
func InjectLocalToolsIntoPrompt(messages []map[string]any, toolsRaw any, resolvedModel string) ([]map[string]any, []string) {
	if !config.ModelSupportsLocalWebTools(resolvedModel) {
		return messages, nil
	}

	// Skip local web_search when the model has native search enabled OR
	// when the model type doesn't support web_search (e.g. vision).
	_, searchEnabled, _ := config.GetModelConfig(resolvedModel)
	if !config.ModelSupportsLocalWebSearch(resolvedModel) {
		searchEnabled = true // force skip web_search
	}
	localPrompt, localNames := BuildLocalToolPrompt(searchEnabled)
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
// When the model has native search or is vision-only, web_search is excluded.
func MergeLocalToolNames(clientNames []string, resolvedModel string) []string {
	if !config.ModelSupportsLocalWebTools(resolvedModel) {
		return clientNames
	}
	_, searchEnabled, _ := config.GetModelConfig(resolvedModel)
	if !config.ModelSupportsLocalWebSearch(resolvedModel) {
		searchEnabled = true // force skip web_search for vision
	}
	_, localNames := BuildLocalToolPromptParts(searchEnabled)
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