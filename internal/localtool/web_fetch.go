package localtool

import (
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

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
		Description:    "Fetch and extract text content from a URL",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "URL to fetch",
				},
			},
			"required": []string{"url"},
		},
		Execution: struct {
			Mode    string `json:"mode"`
			Enabled bool   `json:"enabled"`
			Risk    string `json:"risk"`
		}{
			Mode:    "manual",
			Enabled: true,
			Risk:    "medium",
		},
	}
}

func (e *WebFetchExecutor) Execute(call ToolCall) (*ToolResult, error) {
	startTime := time.Now()

	urlStr, ok := call.Payload["url"].(string)
	if !ok || strings.TrimSpace(urlStr) == "" {
		return &ToolResult{
			Ok:      false,
			Name:    call.Name,
			Summary: "URL is required",
			Error: &ToolError{
				Code:      "empty_url",
				Message:   "url is required",
				Retryable: false,
			},
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			DurationMs:  time.Since(startTime).Milliseconds(),
		}, nil
	}

	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return &ToolResult{
			Ok:      false,
			Name:    call.Name,
			Summary: "Invalid URL",
			Detail:  "Invalid URL: " + urlStr,
			Error: &ToolError{
				Code:      "invalid_url",
				Message:   "Invalid URL: " + urlStr,
				Retryable: false,
			},
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			DurationMs:  time.Since(startTime).Milliseconds(),
		}, nil
	}

	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return &ToolResult{
			Ok:      false,
			Name:    call.Name,
			Summary: "Failed to fetch",
			Detail:  err.Error(),
			Error: &ToolError{
				Code:      "fetch_failed",
				Message:   err.Error(),
				Retryable: false,
			},
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			DurationMs:  time.Since(startTime).Milliseconds(),
		}, nil
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		message := err.Error()
		isPermissionError := strings.Contains(message, "Failed to fetch") ||
			strings.Contains(message, "NetworkError") ||
			strings.Contains(message, "opaque") ||
			strings.Contains(message, "status 0")

		return &ToolResult{
			Ok:      false,
			Name:    call.Name,
			Summary: "Failed to fetch",
			Detail:  message,
			Error: &ToolError{
				Code: func() string {
					if isPermissionError {
						return "fetch_permission_denied"
					}
					return "fetch_failed"
				}(),
				Message: func() string {
					if isPermissionError {
						return "Host permission for " + parsedUrl.Host + " is not granted."
					}
					return message
				}(),
				Retryable: isPermissionError,
			},
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			DurationMs:  time.Since(startTime).Milliseconds(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &ToolResult{
			Ok:      false,
			Name:    call.Name,
			Summary: "HTTP " + resp.Status,
			Detail:  "HTTP " + resp.Status,
			Error: &ToolError{
				Code:      "fetch_http_error",
				Message:   "HTTP " + resp.Status,
				Retryable: true,
			},
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			DurationMs:  time.Since(startTime).Milliseconds(),
		}, nil
	}

	contentType := resp.Header.Get("content-type")

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolResult{
			Ok:      false,
			Name:    call.Name,
			Summary: "Failed to read response",
			Detail:  err.Error(),
			Error: &ToolError{
				Code:      "fetch_read_error",
				Message:   err.Error(),
				Retryable: true,
			},
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			DurationMs:  time.Since(startTime).Milliseconds(),
		}, nil
	}

	if !isTextContentType(contentType) {
		return &ToolResult{
			Ok:      true,
			Name:    call.Name,
			Summary: "Content type: " + contentType,
			Detail:  "The URL returned non-text content (" + contentType + "): " + urlStr,
			Output: map[string]interface{}{
				"url":          urlStr,
				"contentType":  contentType,
				"contentLength": len(body),
			},
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			DurationMs:  time.Since(startTime).Milliseconds(),
		}, nil
	}

	text := string(body)
	var extracted string
	if strings.Contains(contentType, "text/html") {
		extracted = extractTextFromHTML(text)
	} else {
		extracted = text
	}

	maxLength := 50000
	truncated := len(extracted) > maxLength
	var outputText string
	if truncated {
		outputText = extracted[:maxLength] + "\n...[truncated]"
	} else {
		outputText = extracted
	}

	var detail string
	if truncated {
		detail = "Content length: " + string(rune('0'+len(extracted)/1000)) + "KB (truncated to 50KB)"
	} else {
		detail = "Content length: " + string(rune('0'+len(extracted)/1000)) + "KB"
	}

	return &ToolResult{
		Ok:      true,
		Name:    call.Name,
		Summary: "Fetched: " + urlStr,
		Detail:  detail,
		Output: map[string]interface{}{
			"url":          urlStr,
			"content":      outputText,
			"contentType":  contentType,
			"truncated":    truncated,
			"contentLength": len(extracted),
		},
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		DurationMs:  time.Since(startTime).Milliseconds(),
	}, nil
}

func isTextContentType(contentType string) bool {
	return strings.Contains(contentType, "text/html") ||
		strings.Contains(contentType, "text/plain") ||
		strings.Contains(contentType, "application/json")
}

func extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	var extract func(*html.Node)
	var sb strings.Builder

	extract = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
			return
		}
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "nav", "footer", "header":
				return
			case "br", "p", "div", "h1", "h2", "h3", "h4", "h5", "h6":
				sb.WriteString("\n")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(doc)

	text := sb.String()
	text = regexp.MustCompile(`[\r\n]+`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\n\s+\n`).ReplaceAllString(text, "\n")
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")

	text = htmlUnescape(text)

	return strings.TrimSpace(text)
}

func htmlUnescape(text string) string {
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	return text
}