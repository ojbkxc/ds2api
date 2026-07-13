package promptcompat

import (
	"ds2api/internal/prompt"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"ds2api/internal/toolcall"
)

var toolsFilenamePrefixes = []string{
	"tools",
	"functions",
	"actions",
	"api",
	"commands",
	"methods",
	"operations",
	"utilities",
	"helpers",
	"library",
	"toolbox",
	"kit",
}

func randomToolsFilename() string {
	return "tools.txt"
}

// CurrentToolsContextFilename is the fallback filename when no dynamic name is available.
const CurrentToolsContextFilename = "tools.txt"

var toolsTranscriptSummaries = []string{
	"Available tool descriptions and parameter schemas for this request.",
	"Tool definitions and their parameter specifications.",
	"List of available tools with descriptions and parameters.",
	"Tools you can use with their specifications.",
	"Available functions and their parameter schemas.",
	"Tool catalog with descriptions and parameters.",
	"Function definitions for this task.",
	"Tools available for use with details.",
	"API tools and their parameter structures.",
	"Available utilities with their schemas.",
}

var toolDescriptionTemplates = []string{
	"Tool: %s\nDescription: %s\nParameters: %s",
	"- %s: %s\n  Args: %s",
	"Function: %s\nPurpose: %s\nSchema: %s",
	"%s\n  Description: %s\n  Parameters: %s",
	"Action: %s\nWhat it does: %s\nInputs: %s",
	"Name: %s\nAbout: %s\nParameters: %s",
}

var toolsAvailablePhrases = []string{
	"You have access to these tools:",
	"Available tools:",
	"Tools at your disposal:",
	"You can use the following tools:",
	"The following tools are available:",
	"Here are the tools you can use:",
}

var readToolCacheGuards = []string{
	"Read-tool cache guard: If a Read/read_file-style tool result says the file is unchanged, already available in history, should be referenced from previous context, or otherwise provides no file body, treat that result as missing content. Do not repeatedly call the same read request for that missing body. Request a full-content read if the tool supports it, or tell the user that the file contents need to be provided again.",
	"File reading optimization: When a read tool returns that content is unchanged, already in history, or unavailable, treat it as missing data. Avoid repeated identical read attempts. Request full content or inform the user if needed.",
	"Cache-aware reading: If a file read tool indicates content is stale, already present, or not provided, treat this as missing. Do not loop on the same request. Get fresh content or notify the user.",
	"Read operation guard: When read tools report unchanged, cached, or empty results, treat as missing content. Skip redundant calls. Fetch complete content or ask user to provide the file again.",
	"Content retrieval safeguard: If read tool results show no new data or unavailable content, consider it missing. Avoid repeated identical requests. Obtain full content or inform the user accordingly.",
}

var toolsReferencePrompts = []string{
	"Refer to %s for tool definitions and schemas. Only use tools defined there.",
	"Tool specifications are in %s. Use only the tools listed in that file.",
	"See %s for available tools and their parameters. Stick to those tools only.",
	"%s contains the tool catalog. Use only tools defined in that reference.",
	"Check %s for tool definitions. Only call tools specified there.",
	"Tool schemas are documented in %s. Use those tools and follow their specifications.",
	"Refer to %s for available functions. Only use tools defined in that file.",
	"%s has the tool specifications. Use only the tools listed there.",
}

func buildToolsReferencePrompt(filename string) string {
	return "Treat tools.txt as the authoritative list of callable tools and schemas"
}

func GenerateCurrentToolsFilename(historyFilename string) string {
	return CurrentToolsContextFilename
}

func ToolsTranscriptTitle(filename string) string {
	name := strings.TrimSpace(filename)
	if name == "" {
		name = randomToolsFilename()
	}
	return "# " + name
}

type toolPromptParts struct {
	Descriptions string
	Instructions string
	Names        []string
}

func InjectToolPromptWithFilename(messages []map[string]any, tools []any, policy ToolChoicePolicy, toolsFilename string) ([]map[string]any, []string) {
	return injectToolPromptWithDescriptionsAndFilename(messages, tools, policy, true, toolsFilename)
}

func InjectToolPromptInstructionsOnlyWithFilename(messages []map[string]any, tools []any, policy ToolChoicePolicy, toolsFilename string) ([]map[string]any, []string) {
	return injectToolPromptWithDescriptionsAndFilename(messages, tools, policy, false, toolsFilename)
}

func injectToolPromptWithDescriptionsAndFilename(messages []map[string]any, tools []any, policy ToolChoicePolicy, includeDescriptions bool, toolsFilename string) ([]map[string]any, []string) {
	if policy.IsNone() {
		return messages, nil
	}
	parts := buildToolPromptParts(tools, policy)
	if parts.Instructions == "" {
		return messages, parts.Names
	}
	toolPrompt := parts.Instructions
	if includeDescriptions && parts.Descriptions != "" {
		toolPrompt = parts.Descriptions + "\n\n" + toolPrompt
	} else if !includeDescriptions && parts.Descriptions != "" {
		displayName := strings.TrimSpace(toolsFilename)
		if displayName == "" {
			displayName = CurrentToolsContextFilename
		}
		toolPrompt = buildToolsReferencePrompt(displayName) + "\n\n" + toolPrompt
	}

	for i := range messages {
		if messages[i]["role"] == "system" {
			old, _ := messages[i]["content"].(string)
			messages[i]["content"] = strings.TrimSpace(old + "\n\n" + toolPrompt)
			return messages, parts.Names
		}
	}
	messages = append([]map[string]any{{"role": "system", "content": toolPrompt}}, messages...)
	return messages, parts.Names
}

func buildToolPromptParts(tools []any, policy ToolChoicePolicy) toolPromptParts {
	toolSchemas := make([]string, 0, len(tools))
	names := make([]string, 0, len(tools))
	isAllowed := func(name string) bool {
		if strings.TrimSpace(name) == "" {
			return false
		}
		if len(policy.Allowed) == 0 {
			return true
		}
		_, ok := policy.Allowed[name]
		return ok
	}

	for _, t := range tools {
		tool, ok := t.(map[string]any)
		if !ok {
			continue
		}
		name, desc, schema := toolcall.ExtractToolMeta(tool)
		name = strings.TrimSpace(name)
		if !isAllowed(name) {
			continue
		}
		// Sanitize tool metadata to prevent injection through the
		// tool-definition channel (e.g. description containing role prefixes).
		name, desc = prompt.SanitizeToolMeta(name, desc)
		if name == "" {
			continue
		}
		names = append(names, name)
		if desc == "" {
			desc = "No description available"
		}
		b, _ := json.Marshal(schema)
		template := toolDescriptionTemplates[0]
		toolSchemas = append(toolSchemas, fmt.Sprintf(template, name, desc, string(b)))
	}
	if len(toolSchemas) == 0 {
		return toolPromptParts{Names: names}
	}
	phrase := toolsAvailablePhrases[0]
	descriptions := phrase + "\n\n" + strings.Join(toolSchemas, "\n\n")
	instructions := toolcall.BuildToolCallInstructions(names)
	if hasReadLikeTool(names) {
		instructions += "\n\n" + readToolCacheGuards[0]
	}
	if policy.Mode == ToolChoiceRequired {
		instructions += "\n7) For this response, you MUST call at least one tool from the allowed list."
	}
	if policy.Mode == ToolChoiceForced && strings.TrimSpace(policy.ForcedName) != "" {
		instructions += "\n7) For this response, you MUST call exactly this tool name: " + strings.TrimSpace(policy.ForcedName)
		instructions += "\n8) Do not call any other tool."
	}
	return toolPromptParts{
		Descriptions: descriptions,
		Instructions: instructions,
		Names:        names,
	}
}

func BuildOpenAIToolsContextTranscript(toolsRaw any, policy ToolChoicePolicy) (string, []string) {
	return BuildOpenAIToolsContextTranscriptWithFilename(toolsRaw, policy, "")
}

func BuildOpenAIToolsContextTranscriptWithFilename(toolsRaw any, policy ToolChoicePolicy, toolsFilename string) (string, []string) {
	if policy.IsNone() {
		return "", nil
	}
	tools, ok := toolsRaw.([]any)
	if !ok || len(tools) == 0 {
		return "", nil
	}
	parts := buildToolPromptParts(tools, policy)
	if strings.TrimSpace(parts.Descriptions) == "" {
		return "", parts.Names
	}
	var b strings.Builder
	b.WriteString(ToolsTranscriptTitle(toolsFilename))
	b.WriteString("\n")
	b.WriteString(toolsTranscriptSummaries[0])
	b.WriteString("\n\n")
	b.WriteString(parts.Descriptions)
	b.WriteString("\n")
	return b.String(), parts.Names
}

func hasReadLikeTool(names []string) bool {
	for _, name := range names {
		switch normalizeToolNameForGuard(name) {
		case "read", "readfile":
			return true
		}
	}
	return false
}

func normalizeToolNameForGuard(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
