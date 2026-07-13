package localtool

import (
	"time"

	"github.com/google/uuid"
)

type JsonPrimitive = interface{}
type JsonValue = interface{}
type ToolPayload = map[string]interface{}
type ToolProviderId = string
type ToolDescriptorId = string
type ToolCallId = string

type ToolExecutionTrigger string

const (
	ToolExecutionTriggerManualChat   ToolExecutionTrigger = "manual_chat"
	ToolExecutionTriggerAgentRun     ToolExecutionTrigger = "agent_run"
	ToolExecutionTriggerAutomation   ToolExecutionTrigger = "automation"
	ToolExecutionTriggerTest         ToolExecutionTrigger = "test"
	ToolExecutionTriggerSidepanelChat ToolExecutionTrigger = "sidepanel_chat"
)

type ToolExecutionMode string

const (
	ToolExecutionModeAuto     ToolExecutionMode = "auto"
	ToolExecutionModeManual   ToolExecutionMode = "manual"
	ToolExecutionModeDisabled ToolExecutionMode = "disabled"
)

type ToolRiskLevel string

const (
	ToolRiskLevelLow    ToolRiskLevel = "low"
	ToolRiskLevelMedium ToolRiskLevel = "medium"
	ToolRiskLevelHigh   ToolRiskLevel = "high"
)

type ToolTransportKind string

const (
	ToolTransportKindInProcess     ToolTransportKind = "in_process"
	ToolTransportKindHTTP          ToolTransportKind = "http"
	ToolTransportKindSSE           ToolTransportKind = "sse"
	ToolTransportKindStreamableHTTP ToolTransportKind = "streamable_http"
	ToolTransportKindStdioBridge   ToolTransportKind = "stdio_bridge"
	ToolTransportKindNativeMessaging ToolTransportKind = "native_messaging"
)

type ToolProviderKind string

const (
	ToolProviderKindLocal ToolProviderKind = "local"
	ToolProviderKindMCP   ToolProviderKind = "mcp"
)

type ToolProviderIdentity struct {
	Kind        ToolProviderKind   `json:"kind"`
	ID          ToolProviderId     `json:"id"`
	DisplayName string             `json:"display_name"`
	Transport   ToolTransportKind  `json:"transport"`
}

type ToolDescriptorSchema struct {
	Type                 string                 `json:"type"`
	Properties           map[string]JsonValue   `json:"properties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	AdditionalProperties bool                   `json:"additional_properties,omitempty"`
	Description          string                 `json:"description,omitempty"`
}

type ToolDescriptorExecution struct {
	Mode           ToolExecutionMode `json:"mode"`
	Enabled        bool              `json:"enabled"`
	Risk           ToolRiskLevel     `json:"risk"`
	TimeoutMs      int               `json:"timeout_ms,omitempty"`
	MaxResultBytes int               `json:"max_result_bytes,omitempty"`
}

type ToolDescriptor struct {
	ID             ToolDescriptorId         `json:"id"`
	Provider       ToolProviderIdentity     `json:"provider"`
	Name           string                   `json:"name"`
	InvocationName string                   `json:"invocation_name"`
	Title          string                   `json:"title"`
	Description    string                   `json:"description"`
	InputSchema    ToolDescriptorSchema     `json:"input_schema"`
	OutputSchema   *ToolDescriptorSchema    `json:"output_schema,omitempty"`
	Execution      ToolDescriptorExecution  `json:"execution"`
	Annotations    map[string]string        `json:"annotations,omitempty"`
}

type ToolCallSource struct {
	Trigger         ToolExecutionTrigger `json:"trigger"`
	RequestId       string               `json:"request_id,omitempty"`
	ChatSessionId   string               `json:"chat_session_id,omitempty"`
	ParentMessageId int                  `json:"parent_message_id,omitempty"`
	TaskId          string               `json:"task_id,omitempty"`
	RunId           string               `json:"run_id,omitempty"`
	MessageId       int                  `json:"message_id,omitempty"`
	AutomationId    string               `json:"automation_id,omitempty"`
	AutomationRunId string               `json:"automation_run_id,omitempty"`
}

type ToolCall struct {
	ID           ToolCallId        `json:"id,omitempty"`
	DescriptorId ToolDescriptorId  `json:"descriptor_id,omitempty"`
	Provider     *ToolProviderIdentity `json:"provider,omitempty"`
	Name         string            `json:"name"`
	InvocationName string          `json:"invocation_name,omitempty"`
	Payload      ToolPayload       `json:"payload"`
	Raw          string            `json:"raw"`
	ParseError   *ToolError        `json:"parse_error,omitempty"`
	Source       *ToolCallSource   `json:"source,omitempty"`
	CreatedAt    time.Time         `json:"created_at,omitempty"`
}

type ToolError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Retryable bool                   `json:"retryable"`
	Details   ToolPayload            `json:"details,omitempty"`
}

type ToolResult struct {
	Ok          bool                 `json:"ok"`
	Summary     string               `json:"summary"`
	Detail      string               `json:"detail,omitempty"`
	CallId      ToolCallId           `json:"call_id,omitempty"`
	DescriptorId ToolDescriptorId    `json:"descriptor_id,omitempty"`
	Provider    *ToolProviderIdentity `json:"provider,omitempty"`
	Name        string               `json:"name,omitempty"`
	Output      JsonValue            `json:"output,omitempty"`
	Error       *ToolError           `json:"error,omitempty"`
	StartedAt   time.Time            `json:"started_at,omitempty"`
	CompletedAt time.Time            `json:"completed_at,omitempty"`
	DurationMs  int64                `json:"duration_ms,omitempty"`
	Truncated   bool                 `json:"truncated,omitempty"`
}

type ToolExecutionContext struct {
	Trigger         ToolExecutionTrigger `json:"trigger"`
	RequestId       string               `json:"request_id"`
	ChatSessionId   string               `json:"chat_session_id,omitempty"`
	TaskId          string               `json:"task_id,omitempty"`
	RunId           string               `json:"run_id,omitempty"`
	TimeoutMs       int                  `json:"timeout_ms,omitempty"`
	MaxResultBytes  int                  `json:"max_result_bytes,omitempty"`
}

type ToolExecutor interface {
	Execute(call ToolCall, context ToolExecutionContext) (*ToolResult, error)
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
	desc := executor.GetDescriptor()
	r.executors[desc.Name] = executor
	r.executors[desc.InvocationName] = executor
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

type ToolCallHistoryRecord struct {
	ID        string        `json:"id"`
	Call      ToolCall      `json:"call"`
	Result    ToolResult    `json:"result"`
	CreatedAt time.Time     `json:"created_at"`
	Source    ToolExecutionTrigger `json:"source"`
}

type ToolRegistrySnapshot struct {
	Providers  []ToolProviderIdentity `json:"providers"`
	Tools      []ToolDescriptor       `json:"tools"`
	RefreshedAt time.Time             `json:"refreshed_at"`
}

func NewToolCallId() ToolCallId {
	return ToolCallId(uuid.New().String())
}