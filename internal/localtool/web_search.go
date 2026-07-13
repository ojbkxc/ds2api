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

type WebSearchExecutor struct{}

func NewWebSearchExecutor() *WebSearchExecutor {
	return &WebSearchExecutor{}
}

func (e *WebSearchExecutor) GetDescriptor() ToolDescriptor {
	return ToolDescriptor{
		ID:             "local:web:web_search",
		Name:           "web_search",
		InvocationName: "web_search",
		Title:          "Web Search",
		Description:    "Search the web using Bing search engine",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query",
				},
				"topK": map[string]interface{}{
					"type":        "integer",
					"description": "Number of results to return (1-10)",
				},
			},
			"required": []string{"query"},
		},
		Execution: struct {
			Mode    string `json:"mode"`
			Enabled bool   `json:"enabled"`
			Risk    string `json:"risk"`
		}{
			Mode:    "auto",
			Enabled: true,
			Risk:    "low",
		},
	}
}

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

func (e *WebSearchExecutor) Execute(call ToolCall) (*ToolResult, error) {
	startTime := time.Now()

	query, ok := call.Payload["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return &ToolResult{
			Ok:      false,
			Name:    call.Name,
			Summary: "Query is required",
			Error: &ToolError{
				Code:      "empty_query",
				Message:   "query is required",
				Retryable: false,
			},
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			DurationMs:  time.Since(startTime).Milliseconds(),
		}, nil
	}

	topK := 5
	if v, ok := call.Payload["topK"].(float64); ok {
		k := int(v)
		if k < 1 {
			k = 1
		}
		if k > 10 {
			k = 10
		}
		topK = k
	}

	domains := []string{"cn.bing.com", "www.bing.com"}
	var lastError string
	maxTime := 18 * time.Second
	start := time.Now()

	for _, domain := range domains {
		if time.Since(start) > maxTime {
			lastError = "Search timed out (>18s)"
			break
		}

		results, err := bingSearch(domain, query, topK)
		if err != nil {
			lastError = err.Error()
			if strings.Contains(lastError, "opaque") || strings.Contains(lastError, "status 0") {
				break
			}
			continue
		}

		if len(results) == 0 {
			lastError = domain + " returned no parseable search results"
			continue
		}

		output := make([]map[string]string, 0, len(results))
		var detail strings.Builder
		for i, r := range results {
			output = append(output, map[string]string{
				"title":   r.Title,
				"url":     r.URL,
				"snippet": r.Snippet,
			})
			detail.WriteString(renderSearchResult(i+1, r))
			if i < len(results)-1 {
				detail.WriteString("\n")
			}
		}

		return &ToolResult{
			Ok:      true,
			Name:    call.Name,
			Summary: "Search completed with " + string(rune('0'+len(results))) + " results",
			Detail:  detail.String(),
			Output: map[string]interface{}{
				"results": output,
			},
			StartedAt:   startTime,
			CompletedAt: time.Now(),
			DurationMs:  time.Since(startTime).Milliseconds(),
		}, nil
	}

	isPermissionError := strings.Contains(lastError, "Failed to fetch") ||
		strings.Contains(lastError, "NetworkError") ||
		strings.Contains(lastError, "opaque") ||
		strings.Contains(lastError, "status 0")
	hasNoParseableResults := strings.Contains(lastError, "no parseable search results")

	summary := "Search failed"
	if hasNoParseableResults {
		summary = "No search results found"
	}

	return &ToolResult{
		Ok:      false,
		Name:    call.Name,
		Summary: summary,
		Detail:  lastError,
		Error: &ToolError{
			Code: func() string {
				if isPermissionError {
					return "search_permission_denied"
				}
				if hasNoParseableResults {
					return "search_no_results"
				}
				return "search_failed"
			}(),
			Message:   lastError,
			Retryable: !isPermissionError,
		},
		StartedAt:   startTime,
		CompletedAt: time.Now(),
		DurationMs:  time.Since(startTime).Milliseconds(),
	}, nil
}

func bingSearch(domain, query string, topK int) ([]SearchResult, error) {
	u, err := url.Parse("https://" + domain + "/search")
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	client := &http.Client{
		Timeout: 8 * time.Second,
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 0 {
		return nil, ErrHostPermissionDenied(domain)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrHTTPStatus(domain, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrResponseBodyUnreadable(domain)
	}

	if len(body) < 200 {
		return nil, ErrEmptyResponse(domain, len(body))
	}

	return parseBingResults(string(body), topK)
}

func parseBingResults(htmlContent string, topK int) ([]SearchResult, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	var f func(*html.Node)
	f = func(n *html.Node) {
		if len(results) >= topK {
			return
		}
		if n.Type == html.ElementNode && n.Data == "li" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && containsClass(attr.Val, "b_algo") {
					result := extractSearchResultFromLi(n)
					if result.Title != "" && result.URL != "" {
						results = append(results, result)
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return results, nil
}

func extractSearchResultFromLi(n *html.Node) SearchResult {
	var result SearchResult

	var findH2 func(*html.Node)
	findH2 = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h2" {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.Data == "a" {
					for _, attr := range c.Attr {
						if attr.Key == "href" {
							result.URL = attr.Val
							break
						}
					}
					result.Title = textContent(c)
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findH2(c)
		}
	}
	findH2(n)

	if result.URL == "" {
		var findAInLi func(*html.Node)
		findAInLi = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "a" {
				for _, attr := range c.Attr {
					if attr.Key == "href" {
						result.URL = attr.Val
						break
					}
				}
				result.Title = textContent(n)
				return
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				findAInLi(c)
			}
		}
		findAInLi(n)
	}

	var findCaption func(*html.Node)
	findCaption = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && containsClass(attr.Val, "b_caption") {
					var findP func(*html.Node)
					findP = func(n *html.Node) {
						if n.Type == html.ElementNode && n.Data == "p" {
							result.Snippet = textContent(n)
							return
						}
						for c := n.FirstChild; c != nil; c = c.NextSibling {
							findP(c)
						}
					}
					findP(n)
					if result.Snippet == "" {
						result.Snippet = textContent(n)
					}
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findCaption(c)
		}
	}
	findCaption(n)

	result.URL = normalizeURL(result.URL)
	result.Title = cleanText(result.Title)
	result.Snippet = cleanText(result.Snippet)

	return result
}

func textContent(n *html.Node) string {
	var sb strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return sb.String()
}

func containsClass(classList, className string) bool {
	for _, c := range strings.Split(classList, " ") {
		if c == className {
			return true
		}
	}
	return false
}

func normalizeURL(u string) string {
	if strings.HasPrefix(u, "//") {
		return "https:" + u
	}
	return u
}

func cleanText(text string) string {
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

func renderSearchResult(index int, r SearchResult) string {
	return string(rune('0'+index)) + ". [" + r.Title + "](" + r.URL + ")\n   " + r.Snippet
}

func ErrHostPermissionDenied(domain string) error {
	return &SearchError{message: "Host permission denied (opaque response) for " + domain}
}

func ErrHTTPStatus(domain string, status int) error {
	return &SearchError{message: domain + " returned status " + string(rune('0'+status/100)) + string(rune('0'+(status/10)%10)) + string(rune('0'+status%10))}
}

func ErrResponseBodyUnreadable(domain string) error {
	return &SearchError{message: domain + " response body unreadable"}
}

func ErrEmptyResponse(domain string, bytes int) error {
	return &SearchError{message: domain + " returned an empty or blocked response (" + string(rune('0'+bytes/1000)) + "KB)"}
}

type SearchError struct {
	message string
}

func (e *SearchError) Error() string {
	return e.message
}