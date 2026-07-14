package contextcompression

import (
	"fmt"
	"strings"
)

// CompressionLevel represents the aggressiveness of compression.
type CompressionLevel int

const (
	CompressionNone    CompressionLevel = iota
	CompressionSnip                      // Trim long tool outputs, keep head/tail
	CompressionPrune                     // Replace old tool results with placeholders
	CompressionCompact                   // Summarize old messages (requires LLM call)
)

// CompressionConfig holds configuration for compression.
type CompressionConfig struct {
	Enabled          bool    `json:"enabled"`
	SnipRatio        float64 `json:"snip_ratio"`        // default: 0.6 (60% of context window)
	PruneRatio       float64 `json:"prune_ratio"`       // default: 0.8 (80% of context window)
	CompactRatio     float64 `json:"compact_ratio"`     // default: 0.9 (90% of context window)
	ContextWindow    int     `json:"context_window"`    // max tokens
	SnipHeadLines    int     `json:"snip_head_lines"`   // lines to keep at head
	SnipTailLines    int     `json:"snip_tail_lines"`   // lines to keep at tail
	SnipHeadChars    int     `json:"snip_head_chars"`   // chars to keep at head
	SnipTailChars    int     `json:"snip_tail_chars"`   // chars to keep at tail
	MaxToolResultLen int     `json:"max_tool_result_len"` // max length before snipping
}

// DefaultCompressionConfig returns sensible defaults.
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Enabled:          true,
		SnipRatio:        0.6,
		PruneRatio:       0.8,
		CompactRatio:     0.9,
		ContextWindow:    128000, // DeepSeek v4 context
		SnipHeadLines:    80,
		SnipTailLines:    12,
		SnipHeadChars:    10000,
		SnipTailChars:    2000,
		MaxToolResultLen: 5000,
	}
}

// CompressedMessages holds the result of compression.
type CompressedMessages struct {
	Messages       []map[string]any // The compressed messages
	Level          CompressionLevel // What level was applied
	OriginalTokens int              // Token count before compression
	CurrentTokens  int              // Token count after compression
	Archived       []ArchivedMsg    // Archived content for debugging
}

// ArchivedMsg represents a message that was pruned/compacted.
type ArchivedMsg struct {
	Index   int    `json:"index"`
	Role    string `json:"role"`
	Summary string `json:"summary"`
	Tokens  int    `json:"tokens"`
}

// Pruner handles message compression.
type Pruner struct {
	config CompressionConfig
}

// NewPruner creates a new Pruner with the given config.
func NewPruner(config CompressionConfig) *Pruner {
	if config.SnipRatio == 0 {
		config.SnipRatio = 0.6
	}
	if config.PruneRatio == 0 {
		config.PruneRatio = 0.8
	}
	if config.CompactRatio == 0 {
		config.CompactRatio = 0.9
	}
	if config.ContextWindow == 0 {
		config.ContextWindow = 128000
	}
	if config.SnipHeadLines == 0 {
		config.SnipHeadLines = 80
	}
	if config.SnipTailLines == 0 {
		config.SnipTailLines = 12
	}
	if config.SnipHeadChars == 0 {
		config.SnipHeadChars = 10000
	}
	if config.SnipTailChars == 0 {
		config.SnipTailChars = 2000
	}
	if config.MaxToolResultLen == 0 {
		config.MaxToolResultLen = 5000
	}
	return &Pruner{config: config}
}

// Compress applies compression to messages based on token count.
func (p *Pruner) Compress(messages []map[string]any) *CompressedMessages {
	if !p.config.Enabled || len(messages) == 0 {
		return &CompressedMessages{
			Messages:       messages,
			Level:          CompressionNone,
			OriginalTokens: EstimateTokensForMessages(messages),
			CurrentTokens:  EstimateTokensForMessages(messages),
		}
	}

	originalTokens := EstimateTokensForMessages(messages)
	result := &CompressedMessages{
		Messages:       messages,
		Level:          CompressionNone,
		OriginalTokens: originalTokens,
		CurrentTokens:  originalTokens,
	}

	maxTokens := p.config.ContextWindow

	// Level 1: Snip - trim long tool outputs
	if float64(originalTokens) > float64(maxTokens)*p.config.SnipRatio {
		result.Messages = p.snipToolOutputs(result.Messages)
		result.Level = CompressionSnip
		result.CurrentTokens = EstimateTokensForMessages(result.Messages)
	}

	// Level 2: Prune - replace old tool results with placeholders
	if float64(result.CurrentTokens) > float64(maxTokens)*p.config.PruneRatio {
		result.Messages, result.Archived = p.pruneOldToolResults(result.Messages)
		result.Level = CompressionPrune
		result.CurrentTokens = EstimateTokensForMessages(result.Messages)
	}

	// Level 3: Compact - note: actual summarization requires LLM call
	// This marks messages that need summarization, handled by caller
	if float64(result.CurrentTokens) > float64(maxTokens)*p.config.CompactRatio {
		result.Level = CompressionCompact
	}

	return result
}

// snipToolOutputs trims long tool output messages to reduce token usage.
func (p *Pruner) snipToolOutputs(messages []map[string]any) []map[string]any {
	result := make([]map[string]any, len(messages))
	for i, msg := range messages {
		result[i] = msg
		role, _ := msg["role"].(string)
		if role != "tool" {
			continue
		}
		content, _ := msg["content"].(string)
		if len(content) <= p.config.MaxToolResultLen {
			continue
		}

		// Snip: keep head and tail
		snipMsg := make(map[string]any)
		for k, v := range msg {
			snipMsg[k] = v
		}
		snipMsg["content"] = snipContent(content, p.config.SnipHeadLines, p.config.SnipTailLines, p.config.SnipHeadChars, p.config.SnipTailChars)
		result[i] = snipMsg
	}
	return result
}

// pruneOldToolResults replaces old tool results with placeholder markers.
// Keeps recent tool results intact (last 30% of messages).
func (p *Pruner) pruneOldToolResults(messages []map[string]any) ([]map[string]any, []ArchivedMsg) {
	if len(messages) <= 4 {
		return messages, nil
	}

	// Keep the most recent ~30% of messages intact
	keepFrom := int(float64(len(messages)) * 0.7)
	if keepFrom < 2 {
		keepFrom = 2
	}

	var archived []ArchivedMsg
	result := make([]map[string]any, len(messages))

	for i, msg := range messages {
		role, _ := msg["role"].(string)

		if role == "tool" && i < keepFrom {
			content, _ := msg["content"].(string)
			tokens := EstimateTokens(content)

			archived = append(archived, ArchivedMsg{
				Index:   i,
				Role:    role,
				Summary: truncateStr(content, 100),
				Tokens:  tokens,
			})

			// Replace with placeholder
			prunedMsg := make(map[string]any)
			for k, v := range msg {
				prunedMsg[k] = v
			}
			prunedMsg["content"] = fmt.Sprintf("[elided tool result — %d bytes archived]", len(content))
			result[i] = prunedMsg
		} else {
			result[i] = msg
		}
	}

	return result, archived
}

// snipContent trims a long string, keeping head and tail portions.
func snipContent(content string, headLines, tailLines, headChars, tailChars int) string {
	lines := strings.Split(content, "\n")

	if len(lines) <= headLines+tailLines {
		return content
	}

	var b strings.Builder
	// Head
	headEnd := headLines
	if headEnd > len(lines) {
		headEnd = len(lines)
	}
	for _, line := range lines[:headEnd] {
		if len(line) > headChars {
			b.WriteString(line[:headChars])
			b.WriteString("...\n")
		} else {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	b.WriteString(fmt.Sprintf("\n... [%d lines snipped] ...\n\n", len(lines)-headLines-tailLines))

	// Tail
	tailStart := len(lines) - tailLines
	if tailStart < headEnd {
		tailStart = headEnd
	}
	for _, line := range lines[tailStart:] {
		if len(line) > tailChars {
			b.WriteString(line[:tailChars])
			b.WriteString("...\n")
		} else {
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	return strings.TrimSpace(b.String())
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}