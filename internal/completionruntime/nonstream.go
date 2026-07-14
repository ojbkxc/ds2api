package completionruntime

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ds2api/internal/assistantturn"
	"ds2api/internal/auth"
	"ds2api/internal/config"
	dsclient "ds2api/internal/deepseek/client"
	"ds2api/internal/httpapi/openai/history"
	"ds2api/internal/httpapi/openai/shared"
	"ds2api/internal/promptcompat"
	"ds2api/internal/sse"
)

type DeepSeekCaller interface {
	CreateSession(ctx context.Context, a *auth.RequestAuth, maxAttempts int) (string, error)
	GetPow(ctx context.Context, a *auth.RequestAuth, maxAttempts int) (string, error)
	UploadFile(ctx context.Context, a *auth.RequestAuth, req dsclient.UploadFileRequest, maxAttempts int) (*dsclient.UploadFileResult, error)
	CallCompletion(ctx context.Context, a *auth.RequestAuth, payload map[string]any, powResp string, maxAttempts int) (*http.Response, error)
}

type Options struct {
	StripReferenceMarkers bool
	MaxAttempts           int
	RetryEnabled          bool
	RetryMaxAttempts      int
	CurrentInputFile      history.CurrentInputConfigReader
	MaxAccountSwitches    int
	Store                 AccountDisabler
	SessionPool           SessionPoolAccessor
}

type AccountDisabler interface {
	RuntimeMaxAccountSwitches() int
	RuntimeMaxMessagesPerSession() int
	DisableAccount(identifier string) error
}

type SessionPoolAccessor interface {
	Acquire(accountID string, maxMessages int) (sessionID string, parentMessageID int)
	Register(accountID string, sessionID string)
	Update(accountID string, sessionID string, responseMessageID int)
	Invalidate(accountID string)
}

type NonStreamResult struct {
	SessionID string
	Payload   map[string]any
	Turn      assistantturn.Turn
	Attempts  int
}

type StartResult struct {
	SessionID string
	Payload   map[string]any
	Pow       string
	Response  *http.Response
	Request   promptcompat.StandardRequest
}

func StartCompletion(ctx context.Context, ds DeepSeekCaller, a *auth.RequestAuth, stdReq promptcompat.StandardRequest, opts Options) (StartResult, *assistantturn.OutputError) {
	maxAttempts := opts.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	var prepErr *assistantturn.OutputError
	stdReq, prepErr = prepareCurrentInputFile(ctx, ds, a, stdReq, opts)
	if prepErr != nil {
		return StartResult{Request: stdReq}, prepErr
	}

	// 会话池：尝试复用已有会话
	var sessionID string
	var parentMessageID int
	if opts.SessionPool != nil && a != nil && a.AccountID != "" {
		maxMessages := 50
		if opts.Store != nil {
			if m := opts.Store.RuntimeMaxMessagesPerSession(); m > 0 {
				maxMessages = m
			}
		}
		sessionID, parentMessageID = opts.SessionPool.Acquire(a.AccountID, maxMessages)
	}

	// 池中没有可复用的会话，创建新会话
	if sessionID == "" {
		var err error
		sessionID, err = ds.CreateSession(ctx, a, maxAttempts)
		if err != nil {
			if opts.SessionPool != nil && a != nil {
				opts.SessionPool.Invalidate(a.AccountID)
			}
			return StartResult{Request: stdReq}, authOutputError(a)
		}
		if opts.SessionPool != nil && a != nil {
			opts.SessionPool.Register(a.AccountID, sessionID)
		}
	}

	pow, err := ds.GetPow(ctx, a, maxAttempts)
	if err != nil {
		return StartResult{SessionID: sessionID, Request: stdReq}, &assistantturn.OutputError{Status: http.StatusUnauthorized, Message: "Failed to get PoW (invalid token or unknown error).", Code: "error"}
	}
	payload := stdReq.CompletionPayloadWithParent(sessionID, parentMessageID)
	resp, err := ds.CallCompletion(ctx, a, payload, pow, maxAttempts)
	if err != nil {
		return StartResult{SessionID: sessionID, Payload: payload, Pow: pow, Request: stdReq}, &assistantturn.OutputError{Status: http.StatusInternalServerError, Message: "Failed to get completion.", Code: "error"}
	}
	return StartResult{SessionID: sessionID, Payload: payload, Pow: pow, Response: resp, Request: stdReq}, nil
}

func prepareCurrentInputFile(ctx context.Context, ds DeepSeekCaller, a *auth.RequestAuth, stdReq promptcompat.StandardRequest, opts Options) (promptcompat.StandardRequest, *assistantturn.OutputError) {
	if opts.CurrentInputFile == nil || stdReq.CurrentInputFileApplied {
		return stdReq, nil
	}
	out, err := (history.Service{Store: opts.CurrentInputFile, DS: ds}).ApplyCurrentInputFile(ctx, a, stdReq)
	if err != nil {
		status, message := history.MapError(err)
		return out, &assistantturn.OutputError{Status: status, Message: message, Code: "error"}
	}
	return out, nil
}

func ExecuteNonStreamWithRetry(ctx context.Context, ds DeepSeekCaller, a *auth.RequestAuth, stdReq promptcompat.StandardRequest, opts Options) (NonStreamResult, *assistantturn.OutputError) {
	start, startErr := StartCompletion(ctx, ds, a, stdReq, opts)
	if startErr != nil {
		return NonStreamResult{SessionID: start.SessionID, Payload: start.Payload}, startErr
	}
	return ExecuteNonStreamStartedWithRetry(ctx, ds, a, start, opts)
}

func ExecuteNonStreamStartedWithRetry(ctx context.Context, ds DeepSeekCaller, a *auth.RequestAuth, start StartResult, opts Options) (NonStreamResult, *assistantturn.OutputError) {
	stdReq := start.Request
	maxAttempts := opts.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	sessionID := start.SessionID
	payload := start.Payload
	pow := start.Pow

	attempts := 0
	currentResp := start.Response
	usagePrompt := stdReq.PromptTokenText
	accumulatedThinking := ""
	accumulatedRawThinking := ""
	accumulatedToolDetectionThinking := ""
	for {
		turn, outErr := collectAttempt(currentResp, stdReq, usagePrompt, opts)
		if outErr != nil {
			if canRetryOnAlternateAccount(ctx, a, outErr, opts.RetryEnabled, &opts) {
				switched, switchErr := startStandardCompletionOnAlternateAccount(ctx, ds, a, stdReq, opts, maxAttempts)
				if switchErr != nil {
					return NonStreamResult{SessionID: sessionID, Payload: payload, Attempts: attempts}, switchErr
				}
				if switched.Response != nil {
					config.Logger.Info("[completion_runtime_account_switch_retry] retrying after 429", "surface", stdReq.Surface, "stream", false, "account", a.AccountID)
					sessionID = switched.SessionID
					payload = switched.Payload
					pow = switched.Pow
					currentResp = switched.Response
					usagePrompt = stdReq.PromptTokenText
					accumulatedThinking = ""
					accumulatedRawThinking = ""
					accumulatedToolDetectionThinking = ""
					continue
				}
			}
			return NonStreamResult{SessionID: sessionID, Payload: payload, Attempts: attempts}, outErr
		}
		accumulatedThinking += sse.TrimContinuationOverlap(accumulatedThinking, turn.Thinking)
		accumulatedRawThinking += sse.TrimContinuationOverlap(accumulatedRawThinking, turn.RawThinking)
		accumulatedToolDetectionThinking += sse.TrimContinuationOverlap(accumulatedToolDetectionThinking, turn.DetectionThinking)
		turn.Thinking = accumulatedThinking
		turn.RawThinking = accumulatedRawThinking
		turn.DetectionThinking = accumulatedToolDetectionThinking
		turn = assistantturn.BuildTurnFromCollected(sse.CollectResult{
			Text:                  turn.RawText,
			Thinking:              turn.RawThinking,
			ToolDetectionThinking: turn.DetectionThinking,
			ContentFilter:         turn.ContentFilter,
			CitationLinks:         turn.CitationLinks,
			ResponseMessageID:     turn.ResponseMessageID,
		}, buildOptions(stdReq, usagePrompt, opts))

		// 回写 response_message_id 到会话池，供下一次请求复用
		if opts.SessionPool != nil && a != nil && turn.ResponseMessageID > 0 {
			opts.SessionPool.Update(a.AccountID, sessionID, turn.ResponseMessageID)
		}

		retryMax := opts.RetryMaxAttempts
		if retryMax <= 0 {
			retryMax = shared.EmptyOutputRetryMaxAttempts()
		}
		if !opts.RetryEnabled || !assistantturn.ShouldRetryEmptyOutput(turn, attempts, retryMax) {
			lastErr := turn.Error
			if lastErr == nil && strings.TrimSpace(turn.Text) == "" {
				status, message, code := assistantturn.UpstreamEmptyOutputDetail(turn.ContentFilter, turn.Text, turn.Thinking)
				lastErr = &assistantturn.OutputError{Status: status, Message: message, Code: code}
			}
			if canRetryOnAlternateAccount(ctx, a, lastErr, opts.RetryEnabled, &opts) {
				switched, switchErr := startStandardCompletionOnAlternateAccount(ctx, ds, a, stdReq, opts, maxAttempts)
				if switchErr != nil {
					return NonStreamResult{SessionID: sessionID, Payload: payload, Turn: turn, Attempts: attempts}, switchErr
				}
				if switched.Response != nil {
					config.Logger.Info("[completion_runtime_account_switch_retry] retrying after error", "surface", stdReq.Surface, "stream", false, "account", a.AccountID, "status", lastErr.Status)
					sessionID = switched.SessionID
					payload = switched.Payload
					pow = switched.Pow
					currentResp = switched.Response
					usagePrompt = stdReq.PromptTokenText
					accumulatedThinking = ""
					accumulatedRawThinking = ""
					accumulatedToolDetectionThinking = ""
					continue
				}
			}
			// 如果没有备用账户，但错误是可重试的（5xx），使用同一个账户重试
			if lastErr != nil && lastErr.Status >= 500 && opts.RetryEnabled {
				config.Logger.Info("[completion_runtime_same_account_retry] retrying after 5xx error", "surface", stdReq.Surface, "stream", false, "account", a.AccountID, "status", lastErr.Status, "attempt", attempts)
				retryPow, powErr := ds.GetPow(ctx, a, maxAttempts)
				if powErr != nil {
					config.Logger.Warn("[completion_runtime_same_account_retry] retry PoW fetch failed", "surface", stdReq.Surface, "error", powErr)
					retryPow = pow
				}
				nextResp, err := ds.CallCompletion(ctx, a, payload, retryPow, maxAttempts)
				if err == nil && nextResp.StatusCode == http.StatusOK {
					currentResp = nextResp
					pow = retryPow
					accumulatedThinking = ""
					accumulatedRawThinking = ""
					accumulatedToolDetectionThinking = ""
					continue
				}
			}
			return NonStreamResult{SessionID: sessionID, Payload: payload, Turn: turn, Attempts: attempts}, turn.Error
		}

		if attempts >= 1 {
			emptyOutputErr := &assistantturn.OutputError{
				Status:  http.StatusServiceUnavailable,
				Message: "Upstream service is unavailable and returned no output.",
				Code:    "upstream_unavailable",
			}
			if strings.TrimSpace(turn.Thinking) != "" {
				emptyOutputErr.Status = http.StatusTooManyRequests
				emptyOutputErr.Message = "Upstream account hit a rate limit and returned reasoning without visible output."
				emptyOutputErr.Code = "upstream_empty_output"
			}
			if canRetryOnAlternateAccount(ctx, a, emptyOutputErr, opts.RetryEnabled, &opts) {
				switched, switchErr := startStandardCompletionOnAlternateAccount(ctx, ds, a, stdReq, opts, maxAttempts)
				if switchErr != nil {
					return NonStreamResult{SessionID: sessionID, Payload: payload, Turn: turn, Attempts: attempts}, switchErr
				}
				if switched.Response != nil {
					config.Logger.Info("[completion_runtime_account_switch_retry] retrying after empty output", "surface", stdReq.Surface, "stream", false, "account", a.AccountID)
					sessionID = switched.SessionID
					payload = switched.Payload
					pow = switched.Pow
					currentResp = switched.Response
					usagePrompt = stdReq.PromptTokenText
					accumulatedThinking = ""
					accumulatedRawThinking = ""
					accumulatedToolDetectionThinking = ""
					continue
				}
			}
		}

		attempts++
		config.Logger.Info("[completion_runtime_empty_retry] attempting synthetic retry", "surface", stdReq.Surface, "stream", false, "retry_attempt", attempts, "parent_message_id", turn.ResponseMessageID)
		retryPow, powErr := ds.GetPow(ctx, a, maxAttempts)
		if powErr != nil {
			config.Logger.Warn("[completion_runtime_empty_retry] retry PoW fetch failed, falling back to original PoW", "surface", stdReq.Surface, "retry_attempt", attempts, "error", powErr)
			retryPow = pow
		}
		retryPayload := shared.ClonePayloadForEmptyOutputRetry(payload, turn.ResponseMessageID)
		nextResp, err := ds.CallCompletion(ctx, a, retryPayload, retryPow, maxAttempts)
		if err != nil {
			return NonStreamResult{SessionID: sessionID, Payload: payload, Turn: turn, Attempts: attempts}, &assistantturn.OutputError{Status: http.StatusInternalServerError, Message: "Failed to get completion.", Code: "error"}
		}
		usagePrompt = shared.UsagePromptWithEmptyOutputRetry(usagePrompt, attempts)
		currentResp = nextResp
	}
}

func canRetryOnAlternateAccount(ctx context.Context, a *auth.RequestAuth, outErr *assistantturn.OutputError, retryEnabled bool, opts *Options) bool {
	if outErr == nil {
		return false
	}
	if !isAccountSwitchRetryable(outErr) {
		return false
	}
	if !retryEnabled {
		return false
	}
	if a == nil || !a.UseConfigToken {
		return false
	}

	// 自动禁用：当账号遇到封禁类错误时标记为禁用
	maybeAutoDisableAccount(a, outErr, opts)

	// 检查换号次数限制
	maxSwitches := 3
	if opts != nil && opts.MaxAccountSwitches > 0 {
		maxSwitches = opts.MaxAccountSwitches
	} else if opts != nil && opts.Store != nil {
		maxSwitches = opts.Store.RuntimeMaxAccountSwitches()
	}
	if a.SwitchCount >= maxSwitches {
		config.Logger.Warn("[account_switch] max switch count reached", "account", a.AccountID, "switches", a.SwitchCount, "max", maxSwitches)
		return false
	}

	switched := a.SwitchAccount(ctx)
	if switched {
		a.SwitchCount++
	}
	return switched
}

func maybeAutoDisableAccount(a *auth.RequestAuth, outErr *assistantturn.OutputError, opts *Options) {
	if a == nil || a.AccountID == "" || outErr == nil {
		return
	}
	// 只在封禁类错误时自动禁用（非临时性错误）
	if outErr.Status == http.StatusForbidden || outErr.Status == http.StatusUnauthorized {
		if opts != nil && opts.Store != nil {
			config.Logger.Warn("[account_auto_disable] disabling account due to auth error", "account", a.AccountID, "status", outErr.Status, "message", outErr.Message)
			_ = opts.Store.DisableAccount(a.AccountID)
		}
	}
	// 429 (too many requests) 不自动禁用，可能是临时限流
}

func isAccountSwitchRetryable(outErr *assistantturn.OutputError) bool {
	if outErr == nil {
		return false
	}
	switch outErr.Status {
	case http.StatusTooManyRequests:
		return true
	case http.StatusForbidden:
		return true
	case http.StatusUnauthorized:
		return true
	}
	return outErr.Status >= 500
}

func startStandardCompletionOnAlternateAccount(ctx context.Context, ds DeepSeekCaller, a *auth.RequestAuth, stdReq promptcompat.StandardRequest, opts Options, maxAttempts int) (StartResult, *assistantturn.OutputError) {
	var prepErr *assistantturn.OutputError
	stdReq, prepErr = reuploadCurrentInputFileForAccount(ctx, ds, a, stdReq, opts)
	if prepErr != nil {
		return StartResult{Request: stdReq}, prepErr
	}
	// 换号时使旧会话失效，创建新会话
	if opts.SessionPool != nil && a != nil {
		opts.SessionPool.Invalidate(a.AccountID)
	}
	sessionID, err := ds.CreateSession(ctx, a, maxAttempts)
	if err != nil {
		return StartResult{}, authOutputError(a)
	}
	if opts.SessionPool != nil && a != nil {
		opts.SessionPool.Register(a.AccountID, sessionID)
	}
	pow, err := ds.GetPow(ctx, a, maxAttempts)
	if err != nil {
		return StartResult{SessionID: sessionID}, &assistantturn.OutputError{Status: http.StatusUnauthorized, Message: "Failed to get PoW (invalid token or unknown error).", Code: "error"}
	}
	payload := stdReq.CompletionPayload(sessionID)
	resp, err := ds.CallCompletion(ctx, a, payload, pow, maxAttempts)
	if err != nil {
		return StartResult{SessionID: sessionID, Payload: payload, Pow: pow}, &assistantturn.OutputError{Status: http.StatusInternalServerError, Message: "Failed to get completion.", Code: "error"}
	}
	return StartResult{SessionID: sessionID, Payload: payload, Pow: pow, Response: resp, Request: stdReq}, nil
}

func reuploadCurrentInputFileForAccount(ctx context.Context, ds DeepSeekCaller, a *auth.RequestAuth, stdReq promptcompat.StandardRequest, opts Options) (promptcompat.StandardRequest, *assistantturn.OutputError) {
	if opts.CurrentInputFile == nil || !stdReq.CurrentInputFileApplied {
		return stdReq, nil
	}
	out, err := (history.Service{Store: opts.CurrentInputFile, DS: ds}).ReuploadAppliedCurrentInputFile(ctx, a, stdReq)
	if err != nil {
		status, message := history.MapError(err)
		return out, &assistantturn.OutputError{Status: status, Message: message, Code: "error"}
	}
	return out, nil
}

func collectAttempt(resp *http.Response, stdReq promptcompat.StandardRequest, usagePrompt string, opts Options) (assistantturn.Turn, *assistantturn.OutputError) {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			config.Logger.Warn("[completion_runtime] response body close failed", "surface", stdReq.Surface, "error", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}
		return assistantturn.Turn{}, &assistantturn.OutputError{Status: resp.StatusCode, Message: message, Code: "error"}
	}
	result := sse.CollectStream(resp, stdReq.Thinking, false)
	return assistantturn.BuildTurnFromCollected(result, buildOptions(stdReq, usagePrompt, opts)), nil
}

func buildOptions(stdReq promptcompat.StandardRequest, prompt string, opts Options) assistantturn.BuildOptions {
	return assistantturn.BuildOptions{
		Model:                 stdReq.ResponseModel,
		Prompt:                prompt,
		RefFileTokens:         stdReq.RefFileTokens,
		SearchEnabled:         stdReq.Search,
		StripReferenceMarkers: opts.StripReferenceMarkers,
		ToolNames:             stdReq.ToolNames,
		ToolsRaw:              stdReq.ToolsRaw,
		ToolChoice:            stdReq.ToolChoice,
	}
}

func authOutputError(a *auth.RequestAuth) *assistantturn.OutputError {
	if a != nil && a.UseConfigToken {
		return &assistantturn.OutputError{Status: http.StatusUnauthorized, Message: "Account token is invalid. Please re-login the account in admin.", Code: "error"}
	}
	return &assistantturn.OutputError{Status: http.StatusUnauthorized, Message: "Invalid token. If this should be a DS2API key, add it to config.keys first.", Code: "error"}
}

func Errorf(status int, format string, args ...any) *assistantturn.OutputError {
	return &assistantturn.OutputError{Status: status, Message: fmt.Sprintf(format, args...), Code: "error"}
}
