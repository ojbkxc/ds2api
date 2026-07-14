package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MCPServerSpec describes a single MCP server to connect to.
type MCPServerSpec struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"` // "stdio" or "http"
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	URL         string            `json:"url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Enabled     bool              `json:"enabled"`
	TimeoutSecs int               `json:"timeout_secs,omitempty"`
}

// Host manages the lifecycle of multiple MCP clients. It starts and stops
// transports, collects tool descriptors, and routes tool calls to the correct
// server.
type Host struct {
	mu      sync.RWMutex
	clients map[string]*Client
	specs   []MCPServerSpec
	started bool
	logger  *logrus.Entry
}

// NewHost creates a new Host from the given server specifications.
func NewHost(specs []MCPServerSpec) *Host {
	return &Host{
		clients: make(map[string]*Client),
		specs:   specs,
		logger:  logrus.WithField("module", "mcp-host"),
	}
}

// Start initializes all enabled MCP servers. Each server is given a timeout for
// the initialize handshake; servers that fail to start are logged and skipped
// rather than failing the entire host.
func (h *Host) Start(ctx context.Context) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.started {
		return fmt.Errorf("host already started")
	}

	for _, spec := range h.specs {
		if !spec.Enabled {
			continue
		}

		timeout := time.Duration(spec.TimeoutSecs) * time.Second
		if timeout == 0 {
			timeout = 10 * time.Second
		}
		initCtx, cancel := context.WithTimeout(ctx, timeout)

		transport, err := h.createTransport(spec)
		if err != nil {
			cancel()
			h.logger.Warnf("failed to create transport for %s: %v", spec.Name, err)
			continue
		}

		client := NewClient(spec.Name, transport)
		if err := client.Initialize(initCtx); err != nil {
			cancel()
			client.Close()
			h.logger.Warnf("failed to initialize %s: %v", spec.Name, err)
			continue
		}
		cancel()

		h.clients[spec.Name] = client
		h.logger.Infof("MCP server %s started with %d tools", spec.Name, len(client.Tools()))
	}

	h.started = true
	return nil
}

// createTransport builds the appropriate Transport implementation based on the
// spec type.
func (h *Host) createTransport(spec MCPServerSpec) (Transport, error) {
	switch spec.Type {
	case "stdio":
		return newStdioTransport(spec.Name, spec.Command, spec.Args, spec.Env)
	case "http", "streamable-http":
		return newHTTPTransport(spec.Name, spec.URL, spec.Headers), nil
	default:
		return nil, fmt.Errorf("unsupported transport type: %s", spec.Type)
	}
}

// GetAllTools returns the combined list of tool descriptors from all connected
// MCP servers.
func (h *Host) GetAllTools() []ToolInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var allTools []ToolInfo
	for _, client := range h.clients {
		allTools = append(allTools, client.Tools()...)
	}
	return allTools
}

// GetClient returns the MCP client for the named server, or false if not found.
func (h *Host) GetClient(serverName string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, ok := h.clients[serverName]
	return client, ok
}

// CallTool invokes a tool on the named MCP server.
func (h *Host) CallTool(ctx context.Context, serverName, toolName string, args map[string]interface{}) (*CallToolResult, error) {
	client, ok := h.GetClient(serverName)
	if !ok {
		return nil, fmt.Errorf("mcp server not found: %s", serverName)
	}
	return client.CallTool(ctx, toolName, args)
}

// Close shuts down all MCP clients and releases their resources.
func (h *Host) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for name, client := range h.clients {
		h.logger.Infof("closing MCP server %s", name)
		client.Close()
	}
	h.clients = make(map[string]*Client)
	h.started = false
}

// IsStarted reports whether the host has been started.
func (h *Host) IsStarted() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.started
}

// ClientNames returns the names of all currently connected MCP servers.
func (h *Host) ClientNames() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var names []string
	for name := range h.clients {
		names = append(names, name)
	}
	return names
}