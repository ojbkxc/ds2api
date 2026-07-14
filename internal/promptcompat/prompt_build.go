package promptcompat

import (
	"ds2api/internal/config"
	"ds2api/internal/prompt"
	"strings"
)

func buildOpenAIFinalPrompt(messagesRaw []any, toolsRaw any, traceID string, thinkingEnabled bool) (string, []string) {
	return BuildOpenAIPrompt(messagesRaw, toolsRaw, traceID, DefaultToolChoicePolicy(), thinkingEnabled)
}

func BuildOpenAIPrompt(messagesRaw []any, toolsRaw any, traceID string, toolPolicy ToolChoicePolicy, thinkingEnabled bool) (string, []string) {
	return buildOpenAIPrompt(messagesRaw, toolsRaw, traceID, toolPolicy, thinkingEnabled, true, "", false)
}

func BuildOpenAIPromptWithToolInstructionsOnly(messagesRaw []any, toolsRaw any, traceID string, toolPolicy ToolChoicePolicy, thinkingEnabled bool) (string, []string) {
	return buildOpenAIPrompt(messagesRaw, toolsRaw, traceID, toolPolicy, thinkingEnabled, false, "", false)
}

func BuildOpenAIPromptWithToolInstructionsOnlyAndFilename(messagesRaw []any, toolsRaw any, traceID string, toolPolicy ToolChoicePolicy, thinkingEnabled bool, toolsFilename string) (string, []string) {
	return buildOpenAIPrompt(messagesRaw, toolsRaw, traceID, toolPolicy, thinkingEnabled, false, toolsFilename, false)
}

func buildOpenAIPrompt(messagesRaw []any, toolsRaw any, traceID string, toolPolicy ToolChoicePolicy, thinkingEnabled bool, includeToolDescriptions bool, toolsFilename string, skipGuard bool) (string, []string) {
	messages := NormalizeOpenAIMessagesForPrompt(messagesRaw, traceID)
	toolNames := []string{}
	if tools, ok := toolsRaw.([]any); ok && len(tools) > 0 {
		if includeToolDescriptions {
			messages, toolNames = injectToolPromptWithDescriptionsAndFilename(messages, tools, toolPolicy, true, toolsFilename)
		} else {
			messages, toolNames = injectToolPromptWithDescriptionsAndFilename(messages, tools, toolPolicy, false, toolsFilename)
		}
	}
	return prompt.MessagesPrepareWithThinkingAndGuard(messages, thinkingEnabled, skipGuard), toolNames
}

// buildOpenAIPromptWithLocalTools is like buildOpenAIPrompt but also injects
// local web tools (web_search, web_fetch) into the system prompt for models
// that support them.
func buildOpenAIPromptWithLocalTools(messagesRaw []any, toolsRaw any, traceID string, toolPolicy ToolChoicePolicy, thinkingEnabled bool, includeToolDescriptions bool, toolsFilename string, skipGuard bool, resolvedModel string) (string, []string) {
	messages := NormalizeOpenAIMessagesForPrompt(messagesRaw, traceID)
	toolNames := []string{}

	// Inject local web tools for models that support them
	if config.ModelSupportsLocalWebTools(resolvedModel) {
		messages, _ = InjectLocalToolsIntoPrompt(messages, toolsRaw, resolvedModel)
	}

	if tools, ok := toolsRaw.([]any); ok && len(tools) > 0 {
		var clientNames []string
		if includeToolDescriptions {
			messages, clientNames = injectToolPromptWithDescriptionsAndFilename(messages, tools, toolPolicy, true, toolsFilename)
		} else {
			messages, clientNames = injectToolPromptWithDescriptionsAndFilename(messages, tools, toolPolicy, false, toolsFilename)
		}
		toolNames = append(toolNames, clientNames...)
	}

	// Merge local tool names
	if config.ModelSupportsLocalWebTools(resolvedModel) {
		toolNames = MergeLocalToolNames(toolNames, resolvedModel)
	}

	return prompt.MessagesPrepareWithThinkingAndGuard(messages, thinkingEnabled, skipGuard), toolNames
}

// BuildOpenAIPromptWithModel is like BuildOpenAIPrompt but also injects local
// web tools for models that support them. This is the primary entry point for
// request normalization that has access to the resolved model.
func BuildOpenAIPromptWithModel(messagesRaw []any, toolsRaw any, traceID string, toolPolicy ToolChoicePolicy, thinkingEnabled bool, resolvedModel string) (string, []string) {
	if strings.TrimSpace(resolvedModel) == "" {
		return BuildOpenAIPrompt(messagesRaw, toolsRaw, traceID, toolPolicy, thinkingEnabled)
	}
	return buildOpenAIPromptWithLocalTools(messagesRaw, toolsRaw, traceID, toolPolicy, thinkingEnabled, true, "", false, resolvedModel)
}

// BuildOpenAIPromptForAdapter exposes the OpenAI-compatible prompt building flow so
// other protocol adapters (for example Gemini) can reuse the same tool/history
// normalization logic and remain behavior-compatible with chat/completions.
func BuildOpenAIPromptForAdapter(messagesRaw []any, toolsRaw any, traceID string, thinkingEnabled bool) (string, []string) {
	return buildOpenAIFinalPrompt(messagesRaw, toolsRaw, traceID, thinkingEnabled)
}
