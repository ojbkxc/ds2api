package mcp

import (
	"context"
	"encoding/json"
)

// Transport defines the interface for sending JSON-RPC requests and receiving
// responses over a protocol-specific channel (stdio subprocess, HTTP, etc.).
type Transport interface {
	// Call sends a JSON-RPC request with the given method and params, and
	// returns the raw result on success. The caller is responsible for
	// unmarshalling the result into the appropriate type.
	Call(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error)

	// Close releases any resources held by the transport (e.g. kills the
	// subprocess for stdio, closes idle connections for HTTP).
	Close() error
}