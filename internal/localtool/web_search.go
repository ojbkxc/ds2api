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
		InputSchema: ToolDescriptorSchema{
			Type: "object",
			Properties: map[string]JsonValue{
				"query": map[string]JsonValue{"type": "string", "description": "Search query"},
				"topK":  map[string]JsonValue{"type": "integer", "description": "Number of results (1-10)"},
			},
			Required: []string{"query"},
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

type SearchResult struct{ Title, URL, Snippet string }

func (e *WebSearchExecutor) Execute(call ToolCall, context ToolExecutionContext) (*ToolResult, error) {
	startTime := time.Now()
	query, ok := call.Payload["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return &ToolResult{Ok: false, Name: call.Name, Summary: "Query is required", Error: &ToolError{Code: "empty_query", Message: "query is required", Retryable: false}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
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
			output = append(output, map[string]string{"title": r.Title, "url": r.URL, "snippet": r.Snippet})
			detail.WriteString(renderSearchResult(i+1, r))
			if i < len(results)-1 {
				detail.WriteString("\n")
			}
		}
		return &ToolResult{Ok: true, Name: call.Name, Summary: "Search completed with " + fmtInt(len(results)) + " results", Detail: detail.String(), Output: map[string]interface{}{"results": output}, StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds()}, nil
	}

	isPermErr := strings.Contains(lastError, "Failed to fetch") || strings.Contains(lastError, "NetworkError") || strings.Contains(lastError, "opaque") || strings.Contains(lastError, "status 0")
	hasNoResults := strings.Contains(lastError, "no parseable search results")

	return &ToolResult{
		Ok:        false,
		Name:      call.Name,
		Summary:   map[bool]string{true: "No search results found", false: "Search failed"}[hasNoResults],
		Detail:    lastError,
		Error:     &ToolError{Code: map[bool]string{true: map[bool]string{true: "search_permission_denied", false: "search_no_results"}[isPermErr], false: "search_failed"}[hasNoResults], Message: lastError, Retryable: !isPermErr},
		StartedAt: startTime, CompletedAt: time.Now(), DurationMs: time.Since(startTime).Milliseconds(),
	}, nil
}

func fmtInt(n int) string {
	return string(rune('0') + rune(n))
}

func bingSearch(domain, query string, topK int) ([]SearchResult, error) {
	u, _ := url.Parse("https://" + domain + "/search")
	q := u.Query()
	q.Set("q", query)
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 8 * time.Second}
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &SearchError{message: domain + " returned status " + resp.Status}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) < 200 {
		return nil, &SearchError{message: domain + " response body unreadable or empty"}
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
				if attr.Key == "class" && strings.Contains(attr.Val, "b_algo") {
					result := extractSearchResult(n)
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

func extractSearchResult(n *html.Node) SearchResult {
	var result SearchResult

	var findLink func(*html.Node)
	findLink = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					result.URL = attr.Val
					break
				}
			}
			result.Title = textContent(n)
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findLink(c)
		}
	}
	findLink(n)

	var findSnippet func(*html.Node)
	findSnippet = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Contains(attr.Val, "b_caption") {
					result.Snippet = textContent(n)
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findSnippet(c)
		}
	}
	findSnippet(n)

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
	return fmtInt(index) + ". [" + r.Title + "](" + r.URL + ")\n   " + r.Snippet
}

type SearchError struct{ message string }

func (e *SearchError) Error() string {
	return e.message
}
