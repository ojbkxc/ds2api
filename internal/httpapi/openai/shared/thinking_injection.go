package shared

import (
	"strings"

	"ds2api/internal/promptcompat"
)

func ApplyThinkingInjection(store ConfigReader, stdReq promptcompat.StandardRequest) promptcompat.StandardRequest {
	if store == nil || !store.ThinkingInjectionEnabled() || !stdReq.Thinking {
		return stdReq
	}
	messages, changed := promptcompat.AppendThinkingInjectionPromptToLatestUser(stdReq.Messages, store.ThinkingInjectionPrompt())
	if !changed {
		return stdReq
	}
	isDisabledModel := isModelFileUploadDisabled(store, stdReq.ResolvedModel)
	var finalPrompt string
	var toolNames []string
	if isDisabledModel {
		finalPrompt, toolNames = promptcompat.BuildOpenAIPromptSkipGuard(messages, stdReq.ToolsRaw, "", stdReq.ToolChoice, stdReq.Thinking)
	} else {
		finalPrompt, toolNames = promptcompat.BuildOpenAIPrompt(messages, stdReq.ToolsRaw, "", stdReq.ToolChoice, stdReq.Thinking)
	}
	if len(toolNames) == 0 && len(stdReq.ToolNames) > 0 {
		toolNames = stdReq.ToolNames
	}
	stdReq.Messages = messages
	stdReq.FinalPrompt = finalPrompt
	stdReq.ToolNames = toolNames
	return stdReq
}

func isModelFileUploadDisabled(store ConfigReader, model string) bool {
	disabledModels := store.CurrentInputFileDisabledModels()
	if len(disabledModels) == 0 {
		return false
	}
	modelLower := strings.ToLower(strings.TrimSpace(model))
	for _, m := range disabledModels {
		pattern := strings.ToLower(strings.TrimSpace(m))
		if pattern == "" {
			continue
		}
		if pattern == modelLower || strings.HasPrefix(modelLower, pattern) {
			return true
		}
	}
	return false
}
