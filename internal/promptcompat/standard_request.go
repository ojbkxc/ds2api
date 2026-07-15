package promptcompat

import (
	"ds2api/internal/config"
	"strings"
)

type StandardRequest struct {
	Surface                 string
	RequestedModel          string
	ResolvedModel           string
	ResponseModel           string
	Messages                []any
	HistoryText             string
	PromptTokenText         string
	CurrentInputFileApplied bool
	CurrentInputFileID      string
	CurrentToolsFileID      string
	ToolsRaw                any
	FinalPrompt             string
	ToolNames               []string
	ToolChoice              ToolChoicePolicy
	Stream                  bool
	Thinking                bool
	Search                  bool
	RefFileIDs              []string
	RefFileTokens           int
	PassThrough             map[string]any
}

type ToolChoiceMode string

const (
	ToolChoiceAuto     ToolChoiceMode = "auto"
	ToolChoiceNone     ToolChoiceMode = "none"
	ToolChoiceRequired ToolChoiceMode = "required"
	ToolChoiceForced   ToolChoiceMode = "forced"
)

type ToolChoicePolicy struct {
	Mode       ToolChoiceMode
	ForcedName string
	Allowed    map[string]struct{}
}

func DefaultToolChoicePolicy() ToolChoicePolicy {
	return ToolChoicePolicy{Mode: ToolChoiceAuto}
}

func (p ToolChoicePolicy) IsNone() bool {
	return p.Mode == ToolChoiceNone
}

func (p ToolChoicePolicy) IsRequired() bool {
	return p.Mode == ToolChoiceRequired || p.Mode == ToolChoiceForced
}

func (p ToolChoicePolicy) Allows(name string) bool {
	if len(p.Allowed) == 0 {
		return true
	}
	_, ok := p.Allowed[name]
	return ok
}

func (r StandardRequest) CompletionPayload(sessionID string) map[string]any {
	return r.CompletionPayloadWithParent(sessionID, 0)
}

// StripHistoryForSessionReuse strips old messages from the prompt when the
// session is reused (parentMessageID > 0) and the model does not support file
// uploads (e.g. deepseek-v4-pro). In this case DeepSeek already has the full
// conversation history from session context — keeping the full history in the
// prompt would send duplicate context and bloat every request.
//
// Only the system message (if any) and the latest user message are kept;
// everything else is dropped, and the FinalPrompt is rebuilt.
func (r *StandardRequest) StripHistoryForSessionReuse() {
	if len(r.Messages) <= 1 {
		return
	}

	model := r.ResolvedModel
	if model == "" {
		model = r.RequestedModel
	}
	if config.ModelSupportsFileUpload(model) {
		return // flash/vision models use current_input_file instead
	}

	// Keep only the first system message and the latest user message.
	kept := make([]any, 0, 2)
	for _, m := range r.Messages {
		msg, ok := m.(map[string]any)
		if !ok {
			continue
		}
		role := strings.ToLower(strings.TrimSpace(asString(msg["role"])))
		if role == "system" && len(kept) == 0 {
			kept = append(kept, m)
			break
		}
	}
	// Find the latest user message.
	for i := len(r.Messages) - 1; i >= 0; i-- {
		msg, ok := r.Messages[i].(map[string]any)
		if !ok {
			continue
		}
		role := strings.ToLower(strings.TrimSpace(asString(msg["role"])))
		if role == "user" {
			kept = append(kept, r.Messages[i])
			break
		}
	}

	r.Messages = kept
	r.FinalPrompt, r.ToolNames = BuildOpenAIPromptWithModel(kept, r.ToolsRaw, "", r.ToolChoice, r.Thinking, model)
}

func (r StandardRequest) CompletionPayloadWithParent(sessionID string, parentMessageID int) map[string]any {
	modelID := r.ResolvedModel
	if modelID == "" {
		modelID = r.RequestedModel
	}
	modelType := "default"
	if resolvedType, ok := config.GetModelType(modelID); ok {
		modelType = resolvedType
	}
	refFileIDs := make([]any, 0, len(r.RefFileIDs))
	for _, fileID := range r.RefFileIDs {
		if fileID == "" {
			continue
		}
		refFileIDs = append(refFileIDs, fileID)
	}
	payload := map[string]any{
		"chat_session_id":   sessionID,
		"model_type":        modelType,
		"parent_message_id": nil,
		"prompt":            r.FinalPrompt,
		"ref_file_ids":      refFileIDs,
		"thinking_enabled":  r.Thinking,
		"search_enabled":    r.Search,
	}
	if parentMessageID > 0 {
		payload["parent_message_id"] = parentMessageID
	}
	for k, v := range r.PassThrough {
		payload[k] = v
	}
	return payload
}
