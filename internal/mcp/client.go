package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// Client wraps an MCP transport and provides high-level operations: handshake,
// tool discovery, and tool invocation. It handles session expiry for HTTP
// transports by automatically reconnecting and retrying once.
type Client struct {
	name       string
	transport  Transport
	tools      []ToolInfo
	serverInfo ServerInfo
	logger     *logrus.Entry
}

// NewClient creates a new MCP client with the given display name and transport.
func NewClient(name string, transport Transport) *Client {
	return &Client{
		name:      name,
		transport: transport,
		logger:    logrus.WithField("mcp", name),
	}
}

// Initialize performs the MCP handshake: sends initialize, receives server
// capabilities, sends the initialized notification, and discovers tools.
func (c *Client) Initialize(ctx context.Context) error {
	params, _ := json.Marshal(InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		ClientInfo: ClientInfo{
			Name:    "ds2api",
			Version: "1.0.1",
		},
	})

	result, err := c.transport.Call(ctx, "initialize", params)
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	var initResult InitializeResult
	if err := json.Unmarshal(result, &initResult); err != nil {
		return fmt.Errorf("parse initialize result: %w", err)
	}
	c.serverInfo = initResult.ServerInfo
	c.logger.Infof("connected to %s v%s", initResult.ServerInfo.Name, initResult.ServerInfo.Version)

	// Send initialized notification.
	notifParams, _ := json.Marshal(map[string]interface{}{})
	_, _ = c.transport.Call(ctx, "notifications/initialized", notifParams)

	// Discover tools if the server advertises tool capability.
	if caps, ok := initResult.Capabilities["tools"]; ok && caps != nil {
		return c.discoverTools(ctx)
	}

	return nil
}

// discoverTools sends tools/list and caches the result.
func (c *Client) discoverTools(ctx context.Context) error {
	params, _ := json.Marshal(map[string]interface{}{})
	result, err := c.transport.Call(ctx, "tools/list", params)
	if err != nil {
		return fmt.Errorf("list tools: %w", err)
	}

	var listResult ListToolsResult
	if err := json.Unmarshal(result, &listResult); err != nil {
		return fmt.Errorf("parse tools list: %w", err)
	}

	c.tools = listResult.Tools
	c.logger.Infof("discovered %d tools", len(c.tools))
	return nil
}

// CallTool invokes a tool by name on the MCP server. If the session has
// expired (HTTP transport only), it re-initializes and retries once.
func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (*CallToolResult, error) {
	params, _ := json.Marshal(CallToolParams{
		Name:      name,
		Arguments: args,
	})

	result, err := c.transport.Call(ctx, "tools/call", params)
	if err != nil {
		// Check for session expired.
		if _, ok := err.(*sessionExpiredError); ok {
			c.logger.Warn("session expired, reconnecting...")
			if initErr := c.Initialize(ctx); initErr != nil {
				return nil, fmt.Errorf("reconnect failed: %w", initErr)
			}
			// Retry once.
			result, err = c.transport.Call(ctx, "tools/call", params)
			if err != nil {
				return nil, fmt.Errorf("tool call after reconnect: %w", err)
			}
		} else {
			return nil, fmt.Errorf("tool call: %w", err)
		}
	}

	var callResult CallToolResult
	if err := json.Unmarshal(result, &callResult); err != nil {
		return nil, fmt.Errorf("parse tool result: %w", err)
	}

	return &callResult, nil
}

// Tools returns the cached list of tools discovered from this server.
func (c *Client) Tools() []ToolInfo {
	return c.tools
}

// Name returns the client's display name.
func (c *Client) Name() string {
	return c.name
}

// MCPName returns a namespaced tool name in the form mcp__<server>__<tool>,
// suitable for use as a unique invocation name.
func (c *Client) MCPName(name string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, name)
	return fmt.Sprintf("mcp__%s__%s", c.name, safe)
}

// Close releases the underlying transport.
func (c *Client) Close() error {
	return c.transport.Close()
}