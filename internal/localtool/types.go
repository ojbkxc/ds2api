package localtool

import "time"

type ToolResult struct {
	Ok          bool                   `json:"ok"`
	Name        string                 `json:"name"`
	Summary     string                 `json:"summary"`
	Detail      string                 `json:"detail,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       *ToolError             `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at,omitempty"`
	CompletedAt time.Time              `json:"completed_at,omitempty"`
	DurationMs  int64                  `json:"duration_ms,omitempty"`
}

type ToolError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Retryable bool                   `json:"retryable"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type ToolCall struct {
	Name    string                 `json:"name"`
	Payload map[string]interface{} `json:"payload"`
	Raw     string                 `json:"raw"`
}

type ToolDescriptor struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	InvocationName string                 `json:"invocation_name"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	InputSchema    map[string]interface{} `json:"input_schema"`
	Execution      struct {
		Mode    string `json:"mode"`
		Enabled bool   `json:"enabled"`
		Risk    string `json:"risk"`
	} `json:"execution"`
}

type ToolExecutor interface {
	Execute(call ToolCall) (*ToolResult, error)
	GetDescriptor() ToolDescriptor
}

type ToolRegistry struct {
	executors map[string]ToolExecutor
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		executors: make(map[string]ToolExecutor),
	}
}

func (r *ToolRegistry) Register(executor ToolExecutor) {
	r.executors[executor.GetDescriptor().Name] = executor
	r.executors[executor.GetDescriptor().InvocationName] = executor
}

func (r *ToolRegistry) Get(name string) (ToolExecutor, bool) {
	executor, ok := r.executors[name]
	return executor, ok
}

func (r *ToolRegistry) List() []ToolDescriptor {
	descriptors := make([]ToolDescriptor, 0, len(r.executors))
	seen := make(map[string]bool)
	for _, executor := range r.executors {
		desc := executor.GetDescriptor()
		if seen[desc.ID] {
			continue
		}
		seen[desc.ID] = true
		descriptors = append(descriptors, desc)
	}
	return descriptors
}

func (r *ToolRegistry) Has(name string) bool {
	_, ok := r.executors[name]
	return ok
}