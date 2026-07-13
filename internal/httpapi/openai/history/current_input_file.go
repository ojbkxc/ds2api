package history

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"ds2api/internal/auth"
	"ds2api/internal/config"
	dsclient "ds2api/internal/deepseek/client"
	"ds2api/internal/httpapi/openai/shared"
	"ds2api/internal/promptcompat"
)

func randomContentType() string {
	return contentTypes[0]
}

func randomPurpose() string {
	return purposes[0]
}

var contentTypes = []string{
	"text/plain; charset=utf-8",
	"text/plain",
	"application/octet-stream",
	"text/markdown",
	"text/csv",
	"application/json",
}

var purposes = []string{
	"assistants",
	"file-extract",
	"fine-tune",
	"retrieval",
}

type CurrentInputConfigReader interface {
	CurrentInputFileEnabled() bool
	CurrentInputFileMinChars() int
	CurrentInputFileFilenameTemplate() string
	CurrentInputFileDisabledModels() []string
}

type CurrentInputUploader interface {
	UploadFile(ctx context.Context, a *auth.RequestAuth, req dsclient.UploadFileRequest, maxAttempts int) (*dsclient.UploadFileResult, error)
}

type Service struct {
	Store CurrentInputConfigReader
	DS    CurrentInputUploader
}

func (s Service) ApplyCurrentInputFile(ctx context.Context, a *auth.RequestAuth, stdReq promptcompat.StandardRequest) (promptcompat.StandardRequest, error) {
	if stdReq.CurrentInputFileApplied || s.DS == nil || s.Store == nil || a == nil || !s.Store.CurrentInputFileEnabled() {
		return stdReq, nil
	}
	if s.isModelFileUploadDisabled(stdReq.ResolvedModel) {
		return stdReq, nil
	}
	threshold := s.Store.CurrentInputFileMinChars()

	index, text := latestUserInputForFile(stdReq.Messages)
	if index < 0 {
		return stdReq, nil
	}
	if len([]rune(text)) < threshold {
		return stdReq, nil
	}
	filenameTemplate := s.Store.CurrentInputFileFilenameTemplate()
	historyFilename := promptcompat.GenerateCurrentInputFilename(filenameTemplate)
	fileText := promptcompat.BuildOpenAICurrentInputContextTranscriptWithFilename(stdReq.Messages, historyFilename)
	if strings.TrimSpace(fileText) == "" {
		return stdReq, errors.New("current user input file produced empty transcript")
	}
	toolsFilename := promptcompat.GenerateCurrentToolsFilename(historyFilename)
	toolsText, _ := promptcompat.BuildOpenAIToolsContextTranscriptWithFilename(stdReq.ToolsRaw, stdReq.ToolChoice, toolsFilename)
	modelType := "default"
	if resolvedType, ok := config.GetModelType(stdReq.ResolvedModel); ok {
		modelType = resolvedType
	}
	result, err := s.DS.UploadFile(ctx, a, dsclient.UploadFileRequest{
		Filename:    historyFilename,
		ContentType: randomContentType(),
		Purpose:     randomPurpose(),
		ModelType:   modelType,
		Data:        []byte(fileText),
	}, 3)
	if err != nil {
		return stdReq, fmt.Errorf("upload current user input file: %w", err)
	}
	fileID := strings.TrimSpace(result.ID)
	if fileID == "" {
		return stdReq, errors.New("upload current user input file returned empty file id")
	}

	toolFileID := ""
	if strings.TrimSpace(toolsText) != "" {
		result, err := s.DS.UploadFile(ctx, a, dsclient.UploadFileRequest{
			Filename:    toolsFilename,
			ContentType: randomContentType(),
			Purpose:     randomPurpose(),
			ModelType:   modelType,
			Data:        []byte(toolsText),
		}, 3)
		if err != nil {
			return stdReq, fmt.Errorf("upload current tools file: %w", err)
		}
		toolFileID = strings.TrimSpace(result.ID)
		if toolFileID == "" {
			return stdReq, errors.New("upload current tools file returned empty file id")
		}
	}

	messages := []any{
		map[string]any{
			"role":    "user",
			"content": currentInputFilePrompt(historyFilename, toolsFilename, toolFileID != ""),
		},
	}

	stdReq.Messages = messages
	stdReq.HistoryText = fileText
	stdReq.CurrentInputFileApplied = true
	stdReq.CurrentInputFileID = fileID
	stdReq.CurrentToolsFileID = toolFileID
	stdReq.RefFileIDs = prependUniqueRefFileIDs(stdReq.RefFileIDs, fileID, toolFileID)
	stdReq.FinalPrompt, stdReq.ToolNames = promptcompat.BuildOpenAIPromptWithToolInstructionsOnlyAndFilename(messages, stdReq.ToolsRaw, "", stdReq.ToolChoice, stdReq.Thinking, toolsFilename)
	// Token accounting must reflect the actual downstream context:
	// uploaded context files + the continuation live prompt.
	tokenParts := []string{fileText}
	if strings.TrimSpace(toolsText) != "" {
		tokenParts = append(tokenParts, toolsText)
	}
	tokenParts = append(tokenParts, stdReq.FinalPrompt)
	stdReq.PromptTokenText = strings.Join(tokenParts, "\n")
	return stdReq, nil
}

func (s Service) ReuploadAppliedCurrentInputFile(ctx context.Context, a *auth.RequestAuth, stdReq promptcompat.StandardRequest) (promptcompat.StandardRequest, error) {
	if !stdReq.CurrentInputFileApplied || s.DS == nil || a == nil {
		return stdReq, nil
	}
	fileText := strings.TrimSpace(stdReq.HistoryText)
	if fileText == "" {
		return stdReq, nil
	}
	modelType := "default"
	if resolvedType, ok := config.GetModelType(stdReq.ResolvedModel); ok {
		modelType = resolvedType
	}
	filenameTemplate := s.Store.CurrentInputFileFilenameTemplate()
	historyFilename := promptcompat.GenerateCurrentInputFilename(filenameTemplate)
	result, err := s.DS.UploadFile(ctx, a, dsclient.UploadFileRequest{
		Filename:    historyFilename,
		ContentType: randomContentType(),
		Purpose:     randomPurpose(),
		ModelType:   modelType,
		Data:        []byte(stdReq.HistoryText),
	}, 3)
	if err != nil {
		return stdReq, fmt.Errorf("upload current user input file: %w", err)
	}
	fileID := strings.TrimSpace(result.ID)
	if fileID == "" {
		return stdReq, errors.New("upload current user input file returned empty file id")
	}

	toolsFilename := promptcompat.GenerateCurrentToolsFilename(historyFilename)
	toolsText, _ := promptcompat.BuildOpenAIToolsContextTranscriptWithFilename(stdReq.ToolsRaw, stdReq.ToolChoice, toolsFilename)
	toolFileID := ""
	if strings.TrimSpace(toolsText) != "" {
		result, err := s.DS.UploadFile(ctx, a, dsclient.UploadFileRequest{
			Filename:    toolsFilename,
			ContentType: randomContentType(),
			Purpose:     randomPurpose(),
			ModelType:   modelType,
			Data:        []byte(toolsText),
		}, 3)
		if err != nil {
			return stdReq, fmt.Errorf("upload current tools file: %w", err)
		}
		toolFileID = strings.TrimSpace(result.ID)
		if toolFileID == "" {
			return stdReq, errors.New("upload current tools file returned empty file id")
		}
	}

	stdReq.RefFileIDs = replaceGeneratedCurrentInputRefs(stdReq.RefFileIDs, stdReq.CurrentInputFileID, stdReq.CurrentToolsFileID, fileID, toolFileID)
	stdReq.CurrentInputFileID = fileID
	stdReq.CurrentToolsFileID = toolFileID
	return stdReq, nil
}

func latestUserInputForFile(messages []any) (int, string) {
	for i := len(messages) - 1; i >= 0; i-- {
		msg, ok := messages[i].(map[string]any)
		if !ok {
			continue
		}
		role := strings.ToLower(strings.TrimSpace(shared.AsString(msg["role"])))
		if role != "user" {
			continue
		}
		text := promptcompat.NormalizeOpenAIContentForPrompt(msg["content"])
		if strings.TrimSpace(text) == "" {
			return -1, ""
		}
		return i, text
	}
	return -1, ""
}

func currentInputFilePrompt(historyFilename string, toolsFilename string, hasToolsFile bool) string {
	templates := []string{
		"Continue from the latest state in the attached %s context. Treat it as the current working state and answer the latest user request directly.",
		"Use the attached %s as your working context. Start from the most recent state and address the latest user request.",
		"The full context is in the attached %s. Pick up from where it left off and respond to the latest request.",
		"Refer to the attached %s for prior context. Continue from the last state and answer directly.",
		"Attached %s contains the working context. Resume from the latest state and respond to the newest user input.",
		"Review the attached %s for context, then continue from the most recent exchange to address the latest request.",
		"The attached %s holds the current context. Work from the latest state and provide a direct answer.",
		"Use %s as reference context. Start from the last state and answer the latest user message directly.",
		"Please review the attached %s file for context. Continue from the most recent conversation and respond appropriately.",
		"Your working context is provided in %s. Continue the conversation from where it stopped and address the latest query.",
		"I've attached %s with the conversation history. Use this context to continue and respond to the current request.",
		"The attached %s file contains prior conversation. Please continue from the last state and answer the user's latest message.",
		"Referencing %s for context. Continue the conversation naturally from the most recent exchange.",
		"Using %s as context reference. Continue from the latest message and provide a suitable response.",
		"Here's the conversation context in %s. Pick up where we left off and address the current request.",
		"%s contains the prior conversation. Use it to maintain context and respond to the latest input.",
	}
	prompt := fmt.Sprintf(templates[rand.Intn(len(templates))], historyFilename)
	if hasToolsFile {
		toolTemplates := []string{
			" Available tool descriptions and parameter schemas are attached in %s; use only those tools and follow the tool-call format rules in this prompt.",
			" Tool definitions and schemas are in %s; only use those tools and adhere to the tool-call format specified in this prompt.",
			" Refer to %s for available tools and their schemas. Use only those tools and follow the format rules.",
			" The attached %s contains tool descriptions and parameter schemas. Use exclusively those tools as specified.",
			" Tool information is provided in %s. Utilize only these tools and follow the specified call format.",
			" Find available tools and their specifications in %s. Stick to these tools and format accordingly.",
			" %s includes tool descriptions and parameters. Use only the provided tools with correct formatting.",
			" Consult %s for tool definitions. Employ only those tools and maintain proper call structure.",
		}
		prompt += fmt.Sprintf(toolTemplates[rand.Intn(len(toolTemplates))], toolsFilename)
	}
	return prompt
}

func prependUniqueRefFileIDs(existing []string, fileIDs ...string) []string {
	out := make([]string, 0, len(existing)+len(fileIDs))
	seen := map[string]struct{}{}
	for _, fileID := range fileIDs {
		trimmed := strings.TrimSpace(fileID)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		out = append(out, trimmed)
		seen[key] = struct{}{}
	}
	for _, id := range existing {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		out = append(out, trimmed)
		seen[key] = struct{}{}
	}
	return out
}

func replaceGeneratedCurrentInputRefs(existing []string, oldHistoryID, oldToolsID, newHistoryID, newToolsID string) []string {
	filtered := make([]string, 0, len(existing))
	old := map[string]struct{}{}
	for _, id := range []string{oldHistoryID, oldToolsID} {
		trimmed := strings.ToLower(strings.TrimSpace(id))
		if trimmed != "" {
			old[trimmed] = struct{}{}
		}
	}
	for _, id := range existing {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := old[strings.ToLower(trimmed)]; ok {
			continue
		}
		filtered = append(filtered, trimmed)
	}
	return prependUniqueRefFileIDs(filtered, newHistoryID, newToolsID)
}

func (s Service) isModelFileUploadDisabled(model string) bool {
	// DeepSeek expert mode does not support file uploads by default.
	if !config.ModelSupportsFileUpload(model) {
		return true
	}
	disabledModels := s.Store.CurrentInputFileDisabledModels()
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
