package contextcompression

import (
	"sync"

	"github.com/hupe1980/go-tiktoken"
)

var (
	encodingOnce   sync.Once
	sharedEncoding *tiktoken.Encoding
	encodingErr    error
)

// getEncoding returns a shared tiktoken encoding for gpt-4o (cl100k_base, used by GPT-4/DeepSeek).
func getEncoding() (*tiktoken.Encoding, error) {
	encodingOnce.Do(func() {
		sharedEncoding, encodingErr = tiktoken.NewEncodingForModel("gpt-4o")
	})
	return sharedEncoding, encodingErr
}

// EstimateTokens estimates the number of tokens in a text string.
func EstimateTokens(text string) int {
	enc, err := getEncoding()
	if err != nil {
		// Fallback: rough estimate ~4 chars per token
		return len(text) / 4
	}
	tokens, _, err := enc.Encode(text, nil, nil)
	if err != nil {
		return len(text) / 4
	}
	return len(tokens)
}

// EstimateTokensForMessages estimates total tokens for a list of messages.
func EstimateTokensForMessages(messages []map[string]any) int {
	total := 0
	for _, msg := range messages {
		content, _ := msg["content"].(string)
		total += EstimateTokens(content)
		// Add overhead per message (~4 tokens for role markers)
		total += 4
	}
	return total
}

// EstimateTokensForPrompt estimates token count for a full prompt string.
func EstimateTokensForPrompt(prompt string) int {
	return EstimateTokens(prompt)
}