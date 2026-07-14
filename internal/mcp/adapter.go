package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ds2api/internal/localtool"

	"github.com/sirupsen/logrus"
)

// MCPToolAdapter adapts a single MCP tool into a localtool.ToolExecutor so it
// can be registered in the ds2api tool registry and invoked by the LLM like any
// other local tool.
type MCPToolAdapter struct {
	host       *Host
	serverName string
	toolInfo   ToolInfo
	descriptor localtool.ToolDescriptor
	logger     *logrus.Entry
}

// NewMCPToolAdapter creates a new adapter for the given MCP tool. It builds a
// localtool.ToolDescriptor from the tool's MCP metadata, including JSON Schema
// input constraints.
func NewMCPToolAdapter(host *Host, serverName string, tool ToolInfo) *MCPToolAdapter {
	// Build namespaced invocation name.
	mcpName := fmt.Sprintf("mcp__%s__%s", serverName, sanitizeName(tool.Name))

	// Parse the inputSchema into a map for the descriptor.
	var schema map[string]interface{}
	if err := json.Unmarshal(tool.InputSchema, &schema); err != nil {
		schema = map[string]interface{}{
			"type":                 "object",
			"properties":           map[string]interface{}{},
			"additionalProperties": true,
		}
	}

	// Ensure type is set.
	if _, ok := schema["type"]; !ok {
		schema["type"] = "object"
	}

	properties, _ := schema["properties"].(map[string]interface{})
	required, _ := schema["required"].([]interface{})

	desc := localtool.ToolDescriptor{
		ID:   fmt.Sprintf("mcp_%s_%s", serverName, sanitizeName(tool.Name)),
		Name: mcpName,
		Provider: localtool.ToolProviderIdentity{
			Kind:        localtool.ToolProviderKindMCP,
			ID:          serverName,
			DisplayName: fmt.Sprintf("MCP: %s/%s", serverName, tool.Name),
			Transport:   localtool.ToolTransportKindStreamableHTTP,
		},
		InvocationName: mcpName,
		Title:          fmt.Sprintf("%s (%s)", tool.Name, serverName),
		Description:    tool.Description,
		InputSchema: localtool.ToolDescriptorSchema{
			Type:                 "object",
			Properties:           convertProperties(properties),
			Required:             convertRequired(required),
			AdditionalProperties: true,
		},
		Execution: localtool.ToolDescriptorExecution{
			Mode:      localtool.ToolExecutionModeAuto,
			Enabled:   true,
			Risk:      localtool.ToolRiskLevelMedium,
			TimeoutMs: 30000,
		},
	}

	return &MCPToolAdapter{
		host:       host,
		serverName: serverName,
		toolInfo:   tool,
		descriptor: desc,
		logger:     logrus.WithField("mcp-tool", mcpName),
	}
}

// Execute calls the MCP tool through the host and returns a localtool.ToolResult.
func (a *MCPToolAdapter) Execute(call localtool.ToolCall, context localtool.ToolExecutionContext) (*localtool.ToolResult, error) {
	startTime := time.Now()

	ctx, cancel := newGoContext(context)
	defer cancel()

	result, err := a.host.CallTool(ctx, a.serverName, a.toolInfo.Name, call.Payload)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		return &localtool.ToolResult{
			Ok:         false,
			Summary:    "MCP tool call failed",
			Error: &localtool.ToolError{
				Code:      "mcp_error",
				Message:   err.Error(),
				Retryable: false,
			},
			DurationMs: duration,
		}, nil
	}

	// Extract text from content items.
	var output strings.Builder
	for _, item := range result.Content {
		if item.Type == "text" {
			output.WriteString(item.Text)
		}
	}

	outputStr := output.String()
	if result.IsError {
		return &localtool.ToolResult{
			Ok:         false,
			Summary:    truncate(outputStr, 200),
			Detail:     outputStr,
			Output:     outputStr,
			Error: &localtool.ToolError{
				Code:      "mcp_tool_error",
				Message:   truncate(outputStr, 500),
				Retryable: false,
			},
			DurationMs: duration,
		}, nil
	}

	return &localtool.ToolResult{
		Ok:         true,
		Summary:    truncate(outputStr, 200),
		Detail:     outputStr,
		Output:     outputStr,
		DurationMs: duration,
	}, nil
}

// GetDescriptor returns the tool descriptor for registration with the tool
// registry.
func (a *MCPToolAdapter) GetDescriptor() localtool.ToolDescriptor {
	return a.descriptor
}

// sanitizeName replaces any characters that are not alphanumeric, underscore,
// or hyphen with an underscore.
func sanitizeName(name string) string {
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, name)
}

// convertProperties converts the raw MCP schema properties map into
// localtool.JsonValue entries.
func convertProperties(props map[string]interface{}) map[string]localtool.JsonValue {
	result := make(map[string]localtool.JsonValue, len(props))
	for k, v := range props {
		result[k] = localtool.JsonValue(v)
	}
	return result
}

// convertRequired converts the raw required slice from []interface{} (JSON
// unmarshalling) to []string.
func convertRequired(req []interface{}) []string {
	if req == nil {
		return []string{}
	}
	out := make([]string, 0, len(req))
	for _, r := range req {
		if s, ok := r.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// truncate returns s truncated to maxLen, appending "..." if shortened.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// newGoContext creates a context.Context with a deadline derived from the tool
// execution context's timeout.
func newGoContext(tc localtool.ToolExecutionContext) (context.Context, context.CancelFunc) {
	timeout := 30 * time.Second
	if tc.TimeoutMs > 0 {
		timeout = time.Duration(tc.TimeoutMs) * time.Millisecond
	}
	return context.WithTimeout(context.Background(), timeout)
}