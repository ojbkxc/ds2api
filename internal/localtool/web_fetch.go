package localtool

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// privateCIDRs defines IP ranges that must not be reachable via web_fetch
// to prevent SSRF attacks against internal infrastructure.
var privateCIDRs = []string{
	"127.0.0.0/8",    // loopback
	"10.0.0.0/8",     // private
	"172.16.0.0/12",  // private
	"192.168.0.0/16", // private
	"169.254.0.0/16", // link-local
	"0.0.0.0/8",      // current network
	"::1/128",        // IPv6 loopback
	"fc00::/7",       // IPv6 unique local
	"fe80::/10",      // IPv6 link-local
}

var privateNetworks []*net.IPNet

func init() {
	privateNetworks = make([]*net.IPNet, 0, len(privateCIDRs))
	for _, cidr := range privateCIDRs {
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		privateNetworks = append(privateNetworks, n)
	}
}

// isPrivateHost checks whether the given hostname resolves to any private IP.
func isPrivateHost(host string) bool {
	// Direct IP check
	if ip := net.ParseIP(host); ip != nil {
		for _, n := range privateNetworks {
			if n.Contains(ip) {
				return true
			}
		}
		return false
	}
	// DNS resolution check
	ips, err := net.LookupIP(host)
	if err != nil {
		// If DNS fails, err on the side of caution and block.
		return true
	}
	for _, ip := range ips {
		for _, n := range privateNetworks {
			if n.Contains(ip) {
				return true
			}
		}
	}
	return false
}

// fetchHTTPClient is a shared HTTP client tuned for web fetching.
// It uses a longer timeout and a custom transport with explicit
// TLS handshake and dial timeouts to handle slow or distant hosts.
var fetchHTTPClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	},
}

type WebFetchExecutor struct{}

func NewWebFetchExecutor() *WebFetchExecutor {
	return &WebFetchExecutor{}
}

func (e *WebFetchExecutor) GetDescriptor() ToolDescriptor {
	return ToolDescriptor{
		ID:             "local:web:web_fetch",
		Name:           "web_fetch",
		InvocationName: "web_fetch",
		Title:          "Web Fetch",
		Description:    "Fetch and extract text content from a web URL",
		InputSchema: ToolDescriptorSchema{
			Type: "object",
			Properties: map[string]JsonValue{
				"url": map[string]JsonValue{"type": "string", "description": "URL to fetch"},
			},
			Required: []string{"url"},
		},
		Execution: ToolDescriptorExecution{
			Mode:    ToolExecutionModeAuto,
			Enabled: true,
			Risk:    ToolRiskLevelLow,
		},
		Provider: ToolProviderIdentity{
			Kind:        ToolProviderKindLocal,
			ID:          "local-web",
			DisplayName: "Local Web Tools",
			Transport:   ToolTransportKindInProcess,
		},
	}
}

func (e *WebFetchExecutor) Execute(call ToolCall, context ToolExecutionContext) (*ToolResult, error) {
	startTime := time.Now()
	urlStr, ok := call.Payload["url"].(string)
	if !ok || strings.TrimSpace(urlStr) == "" {
		return &ToolResult{Ok: false, Name: call.Name, CallId: call.ID, Summary: "URL is required", Error: &ToolError{Code: "empty_url", Message: "url is required", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return &ToolResult{Ok: false, Name: call.Name, CallId: call.ID, Summary: "Invalid URL", Error: &ToolError{Code: "invalid_url", Message: "invalid URL format", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return &ToolResult{Ok: false, Name: call.Name, CallId: call.ID, Summary: "Unsupported protocol", Error: &ToolError{Code: "unsupported_protocol", Message: "only http and https are supported", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	// SSRF protection: block requests to private/internal IP ranges.
	if isPrivateHost(parsedURL.Hostname()) {
		return &ToolResult{Ok: false, Name: call.Name, CallId: call.ID, Summary: "Access denied", Error: &ToolError{Code: "ssrf_blocked", Message: "access to private/internal addresses is not allowed", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	client := fetchHTTPClient
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return &ToolResult{Ok: false, Name: call.Name, CallId: call.ID, Summary: "Failed to create request", Error: &ToolError{Code: "request_error", Message: err.Error(), Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,text/plain;q=0.8,application/json;q=0.7,*/*;q=0.6")

	resp, err := client.Do(req)
	if err != nil {
		return &ToolResult{Ok: false, Name: call.Name, CallId: call.ID, Summary: "Failed to fetch", Error: &ToolError{Code: "fetch_failed", Message: err.Error(), Retryable: true}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return &ToolResult{Ok: false, Name: call.Name, CallId: call.ID, Summary: "HTTP error", Error: &ToolError{Code: "http_error", Message: "HTTP status " + resp.Status, Retryable: resp.StatusCode >= 500}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	contentType := resp.Header.Get("Content-Type")
	contentTypeLower := strings.ToLower(contentType)
	if strings.Contains(contentTypeLower, "text/html") {
		return e.fetchHTML(resp.Body, call.Name, call.ID, startTime)
	}
	if strings.Contains(contentTypeLower, "text/plain") || strings.Contains(contentTypeLower, "text/xml") ||
		strings.Contains(contentTypeLower, "application/json") || strings.Contains(contentTypeLower, "application/xml") ||
		strings.Contains(contentTypeLower, "text/markdown") || strings.Contains(contentTypeLower, "text/csv") {
		return e.fetchText(resp.Body, call.Name, call.ID, startTime)
	}

	return &ToolResult{Ok: false, Name: call.Name, CallId: call.ID, Summary: "Unsupported content type", Error: &ToolError{Code: "unsupported_content", Message: "content type not supported: " + contentType, Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
}

func (e *WebFetchExecutor) fetchHTML(body io.Reader, name string, callID ToolCallId, startTime time.Time) (*ToolResult, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return &ToolResult{Ok: false, Name: name, CallId: callID, Summary: "Failed to read response", Error: &ToolError{Code: "read_failed", Message: err.Error(), Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	text := extractTextFromHTML(string(data))
	text = cleanExtractedText(text)

	if utf8.RuneCountInString(text) > 5000 {
		runes := []rune(text)
		text = string(runes[:5000]) + "\n\n[Content truncated]"
		return &ToolResult{Ok: true, Name: name, CallId: callID, Summary: "Content fetched (truncated)", Detail: text, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds(), Truncated: true}, nil
	}

	return &ToolResult{Ok: true, Name: name, CallId: callID, Summary: "Content fetched", Detail: text, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
}

func (e *WebFetchExecutor) fetchText(body io.Reader, name string, callID ToolCallId, startTime time.Time) (*ToolResult, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return &ToolResult{Ok: false, Name: name, CallId: callID, Summary: "Failed to read response", Error: &ToolError{Code: "read_failed", Message: err.Error(), Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	text := cleanExtractedText(string(data))
	if utf8.RuneCountInString(text) > 5000 {
		runes := []rune(text)
		text = string(runes[:5000]) + "\n\n[Content truncated]"
		return &ToolResult{Ok: true, Name: name, CallId: callID, Summary: "Content fetched (truncated)", Detail: text, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds(), Truncated: true}, nil
	}

	return &ToolResult{Ok: true, Name: name, CallId: callID, Summary: "Content fetched", Detail: text, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
}

func extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	var sb strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		switch n.Data {
		case "script", "style", "noscript", "iframe", "head":
			return
		}

		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
			sb.WriteString(" ")
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return sb.String()
}

func cleanExtractedText(text string) string {
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`(?m)^\s+$`).ReplaceAllString(text, "")
	text = strings.TrimSpace(text)
	return text
}
