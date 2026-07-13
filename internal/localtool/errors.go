package localtool

type ErrorCategory string

const (
	ErrorCategoryValidation ErrorCategory = "validation"
	ErrorCategoryNetwork    ErrorCategory = "network"
	ErrorCategoryPermission ErrorCategory = "permission"
	ErrorCategoryRateLimit  ErrorCategory = "rate_limit"
	ErrorCategoryInternal   ErrorCategory = "internal"
	ErrorCategoryTimeout    ErrorCategory = "timeout"
	ErrorCategoryNotFound   ErrorCategory = "not_found"
	ErrorCategoryConflict   ErrorCategory = "conflict"
)

type ErrorClassification struct {
	Category  ErrorCategory `json:"category"`
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Retryable bool          `json:"retryable"`
	Details   ToolPayload   `json:"details,omitempty"`
}

func ClassifyError(code string, message string, retryable bool) ErrorClassification {
	switch {
	case code == "empty_query" || code == "empty_url" || code == "invalid_url" || code == "unsupported_protocol" ||
		code == "memory_invalid_payload" || code == "memory_invalid_id":
		return ErrorClassification{
			Category:  ErrorCategoryValidation,
			Code:      code,
			Message:   message,
			Retryable: false,
		}
	case code == "fetch_failed" || stringsContains(message, "network") || stringsContains(message, "connection"):
		return ErrorClassification{
			Category:  ErrorCategoryNetwork,
			Code:      code,
			Message:   message,
			Retryable: retryable,
		}
	case code == "search_permission_denied" || stringsContains(message, "permission") || stringsContains(message, "denied"):
		return ErrorClassification{
			Category:  ErrorCategoryPermission,
			Code:      code,
			Message:   message,
			Retryable: false,
		}
	case code == "memory_not_found" || code == "tool_not_found":
		return ErrorClassification{
			Category:  ErrorCategoryNotFound,
			Code:      code,
			Message:   message,
			Retryable: false,
		}
	case stringsContains(message, "timeout"):
		return ErrorClassification{
			Category:  ErrorCategoryTimeout,
			Code:      code,
			Message:   message,
			Retryable: retryable,
		}
	case stringsContains(message, "rate") || stringsContains(message, "limit"):
		return ErrorClassification{
			Category:  ErrorCategoryRateLimit,
			Code:      code,
			Message:   message,
			Retryable: retryable,
		}
	default:
		return ErrorClassification{
			Category:  ErrorCategoryInternal,
			Code:      code,
			Message:   message,
			Retryable: retryable,
		}
	}
}

func stringsContains(s, substr string) bool {
	if s == "" || substr == "" {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
