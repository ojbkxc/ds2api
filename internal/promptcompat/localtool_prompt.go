package promptcompat

import (
	"ds2api/internal/config"
	"ds2api/internal/localtool"
	"ds2api/internal/toolcall"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand"
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

// webSearchGuidanceVariants provides randomized phrasing for web search
// guidance while keeping the same functional instructions.
var webSearchGuidanceVariants = []string{
	`## Web Search Rules

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
- Do not invent real-time information without searching`,

	`## Web Search

You can use web_search to look up current information. Call it when:
- The user wants real-time data: news, weather, exchange rates, events, etc.
- You're unsure about something and need up-to-date sources
- The user directly asks you to search or look things up
- You need to verify facts or check citations

How it works:
1. Output a web_search call with search keywords
2. Results come back automatically
3. Use those results to answer the user

Tips:
- Match keywords to the user's language and target sources
- Search again with different terms if the first search isn't enough
- Never make up real-time information — search for it instead`,

	`## Search the Web

web_search lets you find current information online. Use it for:
- Real-time info: news, events, weather, rates, etc.
- Anything you're uncertain about that needs current sources
- When the user asks you to search or look something up
- Fact-checking and verifying citations

Process:
1. Call web_search with your query
2. Wait for results (automatic)
3. Answer based on what you found

Guidelines:
- Use keywords in the user's language, targeting appropriate sources
- Try different keywords if needed
- Always search — don't guess real-time information`,

	`## Using web_search

When to search:
- Real-time or current information (news, weather, events, rates)
- Topics you're unsure about and need fresh sources
- User explicitly says "search" or "look up"
- Verifying facts, data, or sources

How:
1. Send a web_search call with your query
2. Results arrive automatically
3. Answer from those results

Remember:
- Choose keywords that match the user's language
- Retry with different keywords if needed
- Don't fabricate real-time data — use web_search`,
}

// webFetchGuidanceVariants provides randomized phrasing for web fetch
// guidance while keeping the same functional instructions.
var webFetchGuidanceVariants = []string{
	`## Web Fetch Rules

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
- If the user asks you to "access", "visit", "open", "read", or "fetch" a URL, use web_fetch`,

	`## Web Fetch

You can use web_fetch to retrieve content from specific URLs. Use it when:
- The user wants you to access, fetch, visit, or read a specific URL
- They give you a link and want its content extracted
- They ask you to open a webpage, check a link, or get info from a web address
- They want you to summarize or analyze a specific web page

How it works:
1. Call web_fetch with the full URL
2. Page content is sent back automatically
3. Answer based on the retrieved content

Tips:
- Always include https:// in the URL
- If fetch fails, tell the user and suggest alternatives
- For URLs, use web_fetch (not web_search)
- Keywords like "access", "visit", "open", "read", "fetch" mean use web_fetch`,

	`## Fetching Web Pages

web_fetch retrieves content from a given URL. Use it for:
- Accessing, fetching, visiting, or reading a specific URL
- Extracting content from a link the user provided
- Opening a webpage, checking a link, or getting info from a web address
- Summarizing or analyzing a specific web page

Process:
1. Send a web_fetch call with the target URL
2. Content loads automatically
3. Respond based on what you retrieved

Guidelines:
- Include the full URL starting with https://
- Explain errors and suggest alternatives if fetch fails
- web_fetch is for URLs, web_search is for general queries
- When the user says "access", "visit", "open", "read", or "fetch" a URL, that means web_fetch`,

	`## Using web_fetch

When to fetch:
- User asks you to access/fetch/visit/read a specific URL
- User provides a link and wants content from it
- User wants you to open a webpage or check a link
- User asks to summarize or analyze a web page

How:
1. Call web_fetch with the full URL
2. Page content returns automatically
3. Answer based on the content

Remember:
- Always use https:// prefix
- If it fails, explain and suggest alternatives
- web_fetch handles URLs; web_search handles general queries
- Trigger words: "access", "visit", "open", "read", "fetch" → web_fetch`,
}

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
	phrase := localToolsAvailablePhrases[idx] + fmt.Sprintf(" %04d", safeRandInt(10000))
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
	instructions := toolcall.BuildToolCallInstructions(toolNames)

	var b strings.Builder
	b.WriteString(descriptions)
	b.WriteString("\n\n")
	b.WriteString(instructions)
	if !skipWebSearch {
		b.WriteString("\n\n")
		b.WriteString(webSearchGuidanceVariants[rand.Intn(len(webSearchGuidanceVariants))] + fmt.Sprintf(" %04d", rand.Intn(10000)))
	}
	b.WriteString("\n\n")
	b.WriteString(webFetchGuidanceVariants[rand.Intn(len(webFetchGuidanceVariants))] + fmt.Sprintf(" %04d", rand.Intn(10000)))
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
func InjectLocalToolsIntoPrompt(messages []map[string]any, toolsRaw any, resolvedModel string) ([]map[string]any, []string) {
	if !config.ModelSupportsLocalWebTools(resolvedModel) {
		return messages, nil
	}

	// Skip local web_search when the model has native search enabled.
	// This avoids the model being confused about which tool to use.
	_, searchEnabled, _ := config.GetModelConfig(resolvedModel)
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
// When the model has native search, web_search is excluded from the merged list.
func MergeLocalToolNames(clientNames []string, resolvedModel string) []string {
	if !config.ModelSupportsLocalWebTools(resolvedModel) {
		return clientNames
	}
	_, searchEnabled, _ := config.GetModelConfig(resolvedModel)
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