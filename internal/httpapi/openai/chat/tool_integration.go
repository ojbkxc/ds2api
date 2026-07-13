package chat

import (
	"context"
	"net/http"
	"time"

	"ds2api/internal/assistantturn"
	"ds2api/internal/auth"
	"ds2api/internal/completionruntime"
	"ds2api/internal/httpapi/openai/history"
	"ds2api/internal/localtool"
	"ds2api/internal/promptcompat"
	"ds2api/internal/toolcall"
)

func executeWithToolCalls(ctx context.Context, ds completionruntime.DeepSeekCaller, store history.CurrentInputConfigReader, a *auth.RequestAuth, stdReq promptcompat.StandardRequest) (completionruntime.NonStreamResult, *assistantturn.OutputError) {
	maxToolIterations := 5
	for i := 0; i < maxToolIterations; i++ {
		result, outErr := completionruntime.ExecuteNonStreamWithRetry(ctx, ds, a, stdReq, completionruntime.Options{
			RetryEnabled:     true,
			CurrentInputFile: store,
		})
		if outErr != nil {
			return result, outErr
		}
		if len(result.Turn.ToolCalls) == 0 {
			return result, nil
		}
		toolResults, execErr := executeToolCalls(result.Turn.ToolCalls)
		if execErr != nil {
			return result, &assistantturn.OutputError{Status: http.StatusInternalServerError, Message: execErr.Error(), Code: "tool_execution_error"}
		}
		stdReq = appendToolResultsToMessages(stdReq, result.Turn.ToolCalls, toolResults)
	}
	return completionruntime.ExecuteNonStreamWithRetry(ctx, ds, a, stdReq, completionruntime.Options{
		RetryEnabled:     true,
		CurrentInputFile: store,
	})
}

func executeToolCalls(calls []toolcall.ParsedToolCall) ([]*localtool.ToolResult, error) {
	results := make([]*localtool.ToolResult, 0, len(calls))
	for _, call := range calls {
		callID := localtool.NewToolCallId()
		result, err := localtool.Execute(localtool.ToolCall{
			ID:             callID,
			Name:           call.Name,
			InvocationName: call.Name,
			Payload:        call.Input,
			CreatedAt:      time.Now(),
			Source: &localtool.ToolCallSource{
				Trigger: localtool.ToolExecutionTriggerManualChat,
			},
		}, localtool.ToolExecutionContext{
			Trigger:   localtool.ToolExecutionTriggerManualChat,
			RequestId: string(callID),
		})
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func appendToolResultsToMessages(stdReq promptcompat.StandardRequest, calls []toolcall.ParsedToolCall, results []*localtool.ToolResult) promptcompat.StandardRequest {
	if len(calls) == 0 || len(results) == 0 {
		return stdReq
	}
	for i, call := range calls {
		if i >= len(results) {
			break
		}
		result := results[i]
		content := ""
		if result.Ok {
			if result.Detail != "" {
				content = result.Detail
			} else if result.Summary != "" {
				content = result.Summary
			}
		} else {
			if result.Error != nil {
				content = "Error: " + result.Error.Message
			} else if result.Detail != "" {
				content = "Error: " + result.Detail
			}
		}
		callID := string(results[i].CallId)
		stdReq.Messages = append(stdReq.Messages, map[string]any{
			"role":       "assistant",
			"tool_calls": []map[string]any{{"id": callID, "type": "function", "function": map[string]any{"name": call.Name, "arguments": call.Input}}},
		})
		stdReq.Messages = append(stdReq.Messages, map[string]any{
			"role":         "tool",
			"tool_call_id": callID,
			"content":      content,
		})
	}
	finalPrompt, toolNames := promptcompat.BuildOpenAIPrompt(stdReq.Messages, stdReq.ToolsRaw, "", stdReq.ToolChoice, stdReq.Thinking)
	stdReq.FinalPrompt = finalPrompt
	stdReq.PromptTokenText = finalPrompt
	stdReq.ToolNames = toolNames
	return stdReq
}
