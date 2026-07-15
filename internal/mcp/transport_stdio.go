package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// stdioTransport speaks newline-delimited JSON-RPC 2.0 over a subprocess's
// stdin/stdout. A dedicated readLoop goroutine owns stdout and demuxes each
// response to the waiting call by id. callMu serialises request/response
// round-trips over the shared pipe so that concurrent callers never interleave
// writes. The pending map lets a call abandon a blocking read the moment its
// context is cancelled without leaking the goroutine.
type stdioTransport struct {
	name   string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr io.WriteCloser

	callMu sync.Mutex // one in-flight request/response at a time over the shared pipe

	mu      sync.Mutex
	nextID  int32
	pending map[int]chan jsonRPCResult
	closed  bool
	done    chan struct{}
}

type jsonRPCResult struct {
	result json.RawMessage
	err    *JSONRPCError
}

// newStdioTransport starts a subprocess with the given command and arguments,
// wires up stdin/stdout pipes, and launches the readLoop goroutine.
func newStdioTransport(name, command string, args []string, env map[string]string) (*stdioTransport, error) {
	cmd := exec.Command(command, args...)
	if len(env) > 0 {
		cmd.Env = cmd.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	// stderr is forwarded to logrus at debug level so operators can inspect
	// subprocess diagnostics without mixing them into stdout.
	stderr := logrus.StandardLogger().WriterLevel(logrus.DebugLevel)
	cmd.Stderr = stderr

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		_ = stderr.Close()
		return nil, fmt.Errorf("start command: %w", err)
	}

	t := &stdioTransport{
		name:    name,
		cmd:     cmd,
		stdin:   stdin,
		stdout:  bufio.NewReader(stdout),
		stderr:  stderr,
		pending: make(map[int]chan jsonRPCResult),
		done:    make(chan struct{}),
	}

	go t.readLoop()
	return t, nil
}

// readLoop owns stdout for the transport's lifetime: it reads one JSON-RPC
// message per line, matches each response to a pending call by id, and delivers
// the result. On any read error it fails every pending call and exits.
func (t *stdioTransport) readLoop() {
	defer close(t.done)
	for {
		line, err := t.stdout.ReadBytes('\n')
		if err != nil {
			t.mu.Lock()
			for _, ch := range t.pending {
				close(ch)
			}
			t.pending = nil
			t.mu.Unlock()
			return
		}
		// Skip empty lines.
		if len(line) == 0 {
			continue
		}

		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		t.mu.Lock()
		ch, ok := t.pending[resp.ID]
		if ok {
			delete(t.pending, resp.ID)
		}
		t.mu.Unlock()

		if ok {
			ch <- jsonRPCResult{result: resp.Result, err: resp.Error}
		}
	}
}

// Call sends a JSON-RPC request over the subprocess's stdin and waits for the
// matching response on stdout. It serialises callers via callMu so that writes
// and the pending map are never interleaved.
func (t *stdioTransport) Call(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error) {
	t.callMu.Lock()
	defer t.callMu.Unlock()

	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil, fmt.Errorf("transport closed")
	}
	id := int(atomic.AddInt32(&t.nextID, 1))
	ch := make(chan jsonRPCResult, 1)
	t.pending[id] = ch
	t.mu.Unlock()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	data = append(data, '\n')

	if _, err := t.stdin.Write(data); err != nil {
		t.mu.Lock()
		delete(t.pending, id)
		t.mu.Unlock()
		return nil, fmt.Errorf("write request: %w", err)
	}

	select {
	case res, ok := <-ch:
		if !ok {
			return nil, fmt.Errorf("transport closed")
		}
		if res.err != nil {
			return nil, fmt.Errorf("rpc error %d: %s", res.err.Code, res.err.Message)
		}
		return res.result, nil
	case <-ctx.Done():
		t.mu.Lock()
		delete(t.pending, id)
		t.mu.Unlock()
		return nil, ctx.Err()
	case <-t.done:
		return nil, fmt.Errorf("transport closed")
	}
}

// Close shuts down the transport: it closes stdin to signal EOF, waits up to
// 3 seconds for graceful exit, then kills the process if it is still running.
func (t *stdioTransport) Close() error {
	t.mu.Lock()
	t.closed = true
	t.mu.Unlock()

	_ = t.stdin.Close()
	_ = t.stderr.Close()
	// Give process a moment to exit gracefully.
	done := make(chan struct{})
	go func() {
		_ = t.cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = t.cmd.Process.Kill()
	}
	return nil
}