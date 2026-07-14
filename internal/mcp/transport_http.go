package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// maxHTTPBody caps how much of a JSON/SSE response body we read, so a
// misbehaving server can't make us buffer without bound.
const maxHTTPBody = 16 << 20 // 16 MiB

// httpTransport speaks MCP's Streamable HTTP transport: every JSON-RPC message
// is an HTTP POST to the server URL. The server replies with either
// application/json (one response) or text/event-stream (an SSE stream carrying
// the response plus any server notifications). The Mcp-Session-Id header, once
// the server assigns one, is echoed on every subsequent request.
//
// The mutex serialises a request and its response. That means concurrent tool
// calls to the same server run one at a time; calls to different servers use
// different transports and stay concurrent.
type httpTransport struct {
	name    string
	url     string
	headers map[string]string
	client  *http.Client

	mu      sync.Mutex
	session string // Mcp-Session-Id, captured from responses
}

// newHTTPTransport creates an HTTP transport for the given server URL and
// optional static headers.
func newHTTPTransport(name, url string, headers map[string]string) *httpTransport {
	return &httpTransport{
		name:    name,
		url:     strings.TrimRight(url, "/"),
		headers: headers,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Call sends a JSON-RPC request via HTTP POST and returns the raw result.
func (t *httpTransport) Call(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error) {
	id := time.Now().UnixNano()
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      int(id % 100000),
		Method:  method,
		Params:  params,
	}
	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", t.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	for k, v := range t.headers {
		httpReq.Header.Set(k, v)
	}

	t.mu.Lock()
	session := t.session
	t.mu.Unlock()
	if session != "" {
		httpReq.Header.Set("Mcp-Session-Id", session)
	}

	resp, err := t.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Capture session id for future requests.
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		t.mu.Lock()
		t.session = sid
		t.mu.Unlock()
	}

	contentType := resp.Header.Get("Content-Type")

	if strings.Contains(contentType, "text/event-stream") {
		return t.readSSE(resp.Body, req.ID)
	}

	// Plain JSON response.
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxHTTPBody))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Check if session expired.
		if resp.StatusCode == http.StatusNotFound && strings.Contains(string(bodyBytes), "-32001") {
			t.mu.Lock()
			t.session = ""
			t.mu.Unlock()
			return nil, &sessionExpiredError{}
		}
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var rpcResp JSONRPCResponse
	if err := json.Unmarshal(bodyBytes, &rpcResp); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

// readSSE scans an SSE stream for the JSON-RPC response matching expectedID,
// skipping server notifications and any other-id messages.
func (t *httpTransport) readSSE(body io.Reader, expectedID int) (json.RawMessage, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var dataBuf strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			// Event boundary.
			if dataBuf.Len() > 0 {
				var resp JSONRPCResponse
				if err := json.Unmarshal([]byte(dataBuf.String()), &resp); err == nil {
					if resp.ID == expectedID {
						if resp.Error != nil {
							return nil, fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
						}
						return resp.Result, nil
					}
				}
				dataBuf.Reset()
			}
			continue
		}
		if strings.HasPrefix(line, "data:") {
			data := strings.TrimPrefix(line, "data:")
			dataBuf.WriteString(strings.TrimSpace(data))
		}
	}
	return nil, fmt.Errorf("sse stream ended without matching response")
}

// Close releases idle HTTP connections.
func (t *httpTransport) Close() error {
	t.client.CloseIdleConnections()
	return nil
}

// sessionExpiredError signals that the MCP session has expired and the caller
// should re-initialize.
type sessionExpiredError struct{}

func (e *sessionExpiredError) Error() string {
	return "mcp session expired"
}