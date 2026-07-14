package chat

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"ds2api/internal/assistantturn"
	"ds2api/internal/auth"
	"ds2api/internal/completionruntime"
	"ds2api/internal/config"
	dsprotocol "ds2api/internal/deepseek/protocol"
	openaifmt "ds2api/internal/format/openai"
	"ds2api/internal/httpapi/openai/history"
	"ds2api/internal/localtool"
	"ds2api/internal/promptcompat"
	"ds2api/internal/sse"
	streamengine "ds2api/internal/stream"
	"ds2api/internal/toolcall"

	"github.com/google/uuid"
)

// maxToolIterations defines the maximum number of tool-call rounds before the
// loop terminates. This prevents infinite loops when the model keeps
// requesting tool calls.
const maxToolIterations = 5

// maxToolErrorRunes limits error messages fed back to the model to prevent
// excessively long error strings from consuming context window.
const maxToolErrorRunes = 500

func executeWithToolCalls(ctx context.Context, ds completionruntime.DeepSeekCaller, store history.CurrentInputConfigReader, a *auth.RequestAuth, stdReq promptcompat.StandardRequest) (completionruntime.NonStreamResult, *assistantturn.OutputError) {
	for i := 0; i < maxToolIterations; i++ {
		result, outErr := completionruntime.ExecuteNonStreamWithRetry(ctx, ds, a, stdReq, completionruntime.Options{
			RetryEnabled:     true,
			CurrentInputFile: store,
			Store:            nil,
			SessionPool:      nil,
		})
		if outErr != nil {
			return result, outErr
		}
		if len(result.Turn.ToolCalls) == 0 {
			return result, nil
		}
		// Execute tools — errors are converted to tool results, not aborts.
		toolResults := executeToolCallsTolerant(result.Turn.ToolCalls)
		stdReq = appendToolResultsToMessages(stdReq, result.Turn.ToolCalls, toolResults)
		// Clear ToolChoice so the model can freely respond in the next round.
		stdReq.ToolChoice = promptcompat.DefaultToolChoicePolicy()
	}
	return completionruntime.ExecuteNonStreamWithRetry(ctx, ds, a, stdReq, completionruntime.Options{
		RetryEnabled:     true,
		CurrentInputFile: store,
	})
}

// executeToolCallsTolerant executes all tool calls in parallel with panic
// recovery. Unlike the old executeToolCalls, errors do NOT abort the batch —
// a failed tool produces an error result that is fed back to the model so it
// can adjust its strategy. This mirrors goai-main's executeToolsParallel.
func executeToolCallsTolerant(calls []toolcall.ParsedToolCall) []*localtool.ToolResult {
	results := make([]*localtool.ToolResult, len(calls))
	var wg sync.WaitGroup
	for i, call := range calls {
		wg.Add(1)
		go func(idx int, c toolcall.ParsedToolCall) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					results[idx] = &localtool.ToolResult{
						CallId: localtool.NewToolCallId(),
						Ok:     false,
						Error: &localtool.ToolError{
							Message: fmt.Sprintf("tool %q panicked: %v", c.Name, r),
						},
					}
					config.Logger.Warn("[tool_loop] tool execution panicked",
						"tool", c.Name, "panic", r)
				}
			}()
			callID := localtool.NewToolCallId()
			result, err := localtool.Execute(localtool.ToolCall{
				ID:             callID,
				Name:           c.Name,
				InvocationName: c.Name,
				Payload:        c.Input,
				CreatedAt:      time.Now(),
				Source: &localtool.ToolCallSource{
					Trigger: localtool.ToolExecutionTriggerManualChat,
				},
			}, localtool.ToolExecutionContext{
				Trigger:   localtool.ToolExecutionTriggerManualChat,
				RequestId: string(callID),
			})
			if err != nil {
				results[idx] = &localtool.ToolResult{
					CallId: callID,
					Ok:     false,
					Error: &localtool.ToolError{
						Message: truncateError(err.Error()),
					},
				}
				config.Logger.Warn("[tool_loop] tool execution failed",
					"tool", c.Name, "error", err.Error())
				return
			}
			results[idx] = result
		}(i, call)
	}
	wg.Wait()
	return results
}

// truncateError limits error strings to maxToolErrorRunes to prevent
// excessively long errors from consuming the model's context window.
func truncateError(s string) string {
	runeCount := utf8.RuneCountInString(s)
	if runeCount <= maxToolErrorRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxToolErrorRunes]) + "..."
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
				content = "Error: " + truncateError(result.Error.Message)
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
	finalPrompt, toolNames := promptcompat.BuildOpenAIPromptWithModel(stdReq.Messages, stdReq.ToolsRaw, "", stdReq.ToolChoice, stdReq.Thinking, stdReq.ResolvedModel)
	stdReq.FinalPrompt = finalPrompt
	stdReq.PromptTokenText = finalPrompt
	stdReq.ToolNames = toolNames
	return stdReq
}

func newCompletionID() string {
	return "chatcmpl-" + uuid.NewString()[:16]
}

// executeStreamWithToolCalls runs a streaming chat completion with a local
// tool execution loop. When the model emits tool calls for local tools
// (web_search, web_fetch, etc.), they are executed silently on the server
// and the result is fed back into the model. The client only sees a
// continuous content stream — tool calls are completely hidden.
//
// Design borrowed from goai-main's streamWithToolLoop:
//   - Tool errors are fed back to the model as tool results (not aborts)
//   - Tools execute in parallel with panic recovery
//   - ToolChoice is cleared after each round to let the model freely respond
//   - Error messages are truncated to prevent context exhaustion
func (h *Handler) executeStreamWithToolCalls(
	w http.ResponseWriter,
	r *http.Request,
	a *auth.RequestAuth,
	stdReq promptcompat.StandardRequest,
	sessionIDRef *string,
	historySession *chatHistorySession,
) {
	ctx := r.Context()

	// Set up SSE headers and create the stream runtime once.
	// It is reused across iterations to preserve client-visible state
	// (completion ID, created timestamp, firstChunkSent flag, etc.).
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	rc := http.NewResponseController(w)
	_, canFlush := w.(http.Flusher)
	if !canFlush {
		config.Logger.Warn("[stream_tool_loop] response writer does not support flush; streaming may be buffered")
	}

	completionID := newCompletionID()
	created := time.Now().Unix()
	stripReferenceMarkers := stripReferenceMarkersEnabled()

	// bufferToolContent must be true so the tool sieve can detect tool calls.
	// emitEarlyToolDeltas is false so tool calls are never sent to the client.
	bufferToolContent := true
	emitEarlyToolDeltas := false

	streamRuntime := newChatStreamRuntime(
		w, rc, canFlush, completionID, created, stdReq.ResponseModel,
		stdReq.FinalPrompt,
		stdReq.Thinking, stdReq.Search, stripReferenceMarkers,
		stdReq.ToolNames, stdReq.ToolsRaw, stdReq.ToolChoice,
		bufferToolContent, emitEarlyToolDeltas,
	)
	streamRuntime.refFileTokens = stdReq.RefFileTokens

	initialType := "text"
	if stdReq.Thinking {
		initialType = "thinking"
	}

	var lastSessionID string
	var finalTurn assistantturn.Turn
	toolIterations := 0

	// Accumulate visible text/thinking across all iterations for chat history.
	var accumulatedText strings.Builder
	var accumulatedThinking strings.Builder
	var accumulatedRawText strings.Builder
	var accumulatedRawThinking strings.Builder
	var accumulatedUsage struct {
		inputTokens     int
		outputTokens    int
		totalTokens     int
		reasoningTokens int
	}

	for i := 0; i < maxToolIterations; i++ {
		toolIterations = i + 1

		// Start a new completion for this iteration.
		start, outErr := completionruntime.StartCompletion(ctx, h.DS, a, stdReq, completionruntime.Options{
			CurrentInputFile: h.Store,
			Store:            h.Store,
			SessionPool:      h.DS.SessionPool(),
		})
		lastSessionID = start.SessionID
		if sessionIDRef != nil {
			*sessionIDRef = start.SessionID
		}
		if outErr != nil {
			if historySession != nil {
				historySession.error(outErr.Status, outErr.Message, outErr.Code, "", "")
			}
			writeOpenAIErrorWithCode(w, outErr.Status, outErr.Message, outErr.Code)
			return
		}

		// If this is not the first iteration, reset per-iteration state.
		// Client-visible state (firstChunkSent, completionID, etc.) is preserved.
		if i > 0 {
			streamRuntime.resetForNextIteration()
			streamRuntime.finalPrompt = stdReq.FinalPrompt
		}

		resp := start.Response
		shouldContinue := h.consumeToolLoopIteration(ctx, resp, streamRuntime, initialType, stdReq.Thinking, historySession)
		_ = resp.Body.Close()

		if !shouldContinue {
			// Stream ended with an error or cancellation — already handled.
			return
		}

		// Accumulate this iteration's text/thinking for the final history entry.
		accumulatedText.WriteString(streamRuntime.accumulator.Text.String())
		accumulatedThinking.WriteString(streamRuntime.accumulator.Thinking.String())
		accumulatedRawText.WriteString(streamRuntime.accumulator.RawText.String())
		accumulatedRawThinking.WriteString(streamRuntime.accumulator.RawThinking.String())

		// Build the turn from the accumulator to check for tool calls.
		turn := streamRuntime.buildTurnFromAccumulator("stop")
		finalTurn = turn

		// Accumulate usage across all iterations.
		accumulatedUsage.inputTokens += turn.Usage.InputTokens
		accumulatedUsage.outputTokens += turn.Usage.OutputTokens
		accumulatedUsage.totalTokens += turn.Usage.TotalTokens
		accumulatedUsage.reasoningTokens += turn.Usage.ReasoningTokens

		// If no tool calls, this is the final iteration.
		if len(turn.ToolCalls) == 0 {
			// Check for empty output — upstream returned nothing useful.
			// This mirrors the same check in the normal stream path so that
			// clients receive a failure frame instead of a silent success.
			outcome := assistantturn.FinalizeTurn(turn, assistantturn.FinalizeOptions{})
			if outcome.ShouldFail {
				streamRuntime.sendFailedChunk(outcome.Error.Status, outcome.Error.Message, outcome.Error.Code)
				if historySession != nil {
					finalHistText := historyTextForArchive(accumulatedRawText.String()+streamRuntime.accumulator.RawText.String(), accumulatedText.String()+streamRuntime.accumulator.Text.String())
					finalHistThink := historyThinkingForArchive(accumulatedRawThinking.String()+streamRuntime.accumulator.RawThinking.String(), "", accumulatedThinking.String()+streamRuntime.accumulator.Thinking.String())
					historySession.error(outcome.Error.Status, outcome.Error.Message, outcome.Error.Code, finalHistThink, finalHistText)
				}
				config.Logger.Info("[stream_tool_loop] empty output detected, sending failure frame",
					"iterations", toolIterations,
					"status", outcome.Error.Status,
					"code", outcome.Error.Code)
				return
			}
			// Drain any remaining buffered content before sending finish.
			streamRuntime.flushBufferedContentOnly()
			break
		}

		config.Logger.Info("[stream_tool_loop] detected tool calls, executing locally",
			"iteration", i+1,
			"tool_count", len(turn.ToolCalls),
		)

		// Drain any remaining buffered content (but NOT tool calls) to the client.
		streamRuntime.flushBufferedContentOnly()

		// Execute all tool calls locally — errors are tolerated and fed back
		// to the model as tool results, not aborts.
		toolResults := executeToolCallsTolerant(turn.ToolCalls)

		// Update session pool with the response message ID so the next
		// iteration reuses the same session with correct parent.
		if h.DS.SessionPool() != nil && a != nil && a.AccountID != "" && streamRuntime.responseMessageID > 0 {
			h.DS.SessionPool().Update(a.AccountID, lastSessionID, streamRuntime.responseMessageID)
		}

		// Append tool results to messages and rebuild prompt.
		stdReq = appendToolResultsToMessages(stdReq, turn.ToolCalls, toolResults)
		// Clear ToolChoice so the model can freely respond in the next round
		// instead of being forced to call tools again (prevents infinite loops).
		stdReq.ToolChoice = promptcompat.DefaultToolChoicePolicy()
	}

	// Final iteration — send finish chunk + [DONE].
	// Build final history text from all accumulated iterations.
	finalHistoryText := historyTextForArchive(accumulatedRawText.String(), accumulatedText.String())
	finalHistoryThinking := historyThinkingForArchive(accumulatedRawThinking.String(), "", accumulatedThinking.String())

	// Build final usage from accumulated iterations.
	finalUsage := map[string]any{
		"prompt_tokens":     accumulatedUsage.inputTokens,
		"completion_tokens": accumulatedUsage.outputTokens,
		"total_tokens":      accumulatedUsage.totalTokens,
		"completion_tokens_details": map[string]any{
			"reasoning_tokens": accumulatedUsage.reasoningTokens,
		},
	}

	if toolIterations > 0 && len(finalTurn.ToolCalls) == 0 {
		// Normal end: no more tool calls.
		streamRuntime.finalFinishReason = "stop"
		streamRuntime.finalUsage = finalUsage
		finishChunk := openaifmt.BuildChatStreamChunk(
			completionID,
			created,
			stdReq.ResponseModel,
			[]map[string]any{openaifmt.BuildChatStreamFinishChoice(0, "stop")},
			finalUsage,
		)
		streamRuntime.sendChunk(finishChunk)
		streamRuntime.sendDone()

		if historySession != nil {
			historySession.success(http.StatusOK, finalHistoryThinking, finalHistoryText, "stop", finalUsage)
		}
		config.Logger.Info("[stream_tool_loop] completed", "iterations", toolIterations)
	} else if toolIterations >= maxToolIterations {
		// Hit max iterations — still send finish.
		streamRuntime.finalFinishReason = "stop"
		streamRuntime.finalUsage = finalUsage
		finishChunk := openaifmt.BuildChatStreamChunk(
			completionID,
			created,
			stdReq.ResponseModel,
			[]map[string]any{openaifmt.BuildChatStreamFinishChoice(0, "stop")},
			finalUsage,
		)
		streamRuntime.sendChunk(finishChunk)
		streamRuntime.sendDone()
		if historySession != nil {
			historySession.success(http.StatusOK, finalHistoryThinking, finalHistoryText, "stop", finalUsage)
		}
		config.Logger.Warn("[stream_tool_loop] reached max iterations", "iterations", toolIterations)
	}
}

// consumeToolLoopIteration consumes a single streaming response for the tool
// loop. It returns true if the stream completed successfully and we should
// check for tool calls, or false if the stream failed / was cancelled.
func (h *Handler) consumeToolLoopIteration(
	ctx context.Context,
	resp *http.Response,
	streamRuntime *chatStreamRuntime,
	initialType string,
	thinkingEnabled bool,
	historySession *chatHistorySession,
) bool {
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		streamRuntime.sendFailedChunk(resp.StatusCode, string(body), "upstream_error")
		if historySession != nil {
			historySession.error(resp.StatusCode, string(body), "upstream_error", streamRuntime.historyThinking(), streamRuntime.historyText())
		}
		return false
	}

	finalReason := "stop"
	streamengine.ConsumeSSE(streamengine.ConsumeConfig{
		Context:             ctx,
		Body:                resp.Body,
		ThinkingEnabled:     thinkingEnabled,
		InitialType:         initialType,
		KeepAliveInterval:   time.Duration(dsprotocol.KeepAliveTimeout) * time.Second,
		IdleTimeout:         time.Duration(dsprotocol.StreamIdleTimeout) * time.Second,
		MaxKeepAliveNoInput: dsprotocol.MaxKeepaliveCount,
	}, streamengine.ConsumeHooks{
		OnKeepAlive: streamRuntime.sendKeepAlive,
		OnParsed: func(parsed sse.LineResult) streamengine.ParsedDecision {
			decision := streamRuntime.onParsed(parsed)
			if historySession != nil {
				historySession.progress(streamRuntime.historyThinking(), streamRuntime.historyText())
			}
			return decision
		},
		OnFinalize: func(reason streamengine.StopReason, _ error) {
			if string(reason) == "content_filter" {
				finalReason = "content_filter"
			}
		},
		OnContextDone: func() {
			streamRuntime.markContextCancelled()
			if historySession != nil {
				historySession.stopped(streamRuntime.historyThinking(), streamRuntime.historyText(), string(streamengine.StopReasonContextCancelled))
			}
		},
	})

	if streamRuntime.finalErrorCode == string(streamengine.StopReasonContextCancelled) {
		// Drain any remaining response body to prevent goroutine leaks.
		_, _ = io.Copy(io.Discard, resp.Body)
		return false
	}

	// For content_filter, we need to send the error and stop.
	if finalReason == "content_filter" {
		turn := streamRuntime.buildTurnFromAccumulator("content_filter")
		outcome := assistantturn.FinalizeTurn(turn, assistantturn.FinalizeOptions{})
		if outcome.ShouldFail {
			streamRuntime.sendFailedChunk(outcome.Error.Status, outcome.Error.Message, outcome.Error.Code)
			if historySession != nil {
				historySession.error(outcome.Error.Status, outcome.Error.Message, outcome.Error.Code, streamRuntime.historyThinking(), streamRuntime.historyText())
			}
			return false
		}
	}

	return true
}
