package contextcompression

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
)

// Compressor is the main compression orchestrator.
type Compressor struct {
	config CompressionConfig
	pruner *Pruner
	logger *logrus.Entry
}

// GlobalCompressor is the package-level compressor instance, set at startup.
var GlobalCompressor *Compressor

// SetGlobal sets the global compressor instance.
func SetGlobal(c *Compressor) {
	GlobalCompressor = c
}

// NewCompressor creates a new Compressor.
func NewCompressor(config CompressionConfig) *Compressor {
	return &Compressor{
		config: config,
		pruner: NewPruner(config),
		logger: logrus.WithField("module", "context-compression"),
	}
}

// CompressPrompt compresses the prompt string using progressive compression.
// Returns the compressed prompt, the compression level applied, and token stats.
func (c *Compressor) CompressPrompt(prompt string) (string, CompressionLevel, int, int) {
	if !c.config.Enabled {
		return prompt, CompressionNone, 0, 0
	}

	originalTokens := EstimateTokensForPrompt(prompt)
	maxTokens := c.config.ContextWindow

	if float64(originalTokens) <= float64(maxTokens)*c.config.SnipRatio {
		return prompt, CompressionNone, originalTokens, originalTokens
	}

	c.logger.Debugf("compressing prompt: %d tokens (limit: %d)", originalTokens, maxTokens)

	level := CompressionNone
	compressed := prompt

	// Level 1: Snip — trim sections that look like tool results
	// (early return above already ensures we exceed SnipRatio)
	compressed = c.snipToolSections(compressed)
	level = CompressionSnip

	currentTokens := EstimateTokensForPrompt(compressed)

	// Level 2: Prune — replace old sections with placeholders
	if float64(currentTokens) > float64(maxTokens)*c.config.PruneRatio {
		compressed = c.pruneOldSections(compressed)
		level = CompressionPrune
		currentTokens = EstimateTokensForPrompt(compressed)
	}

	// Level 3: Compact — truncate to fit
	if float64(currentTokens) > float64(maxTokens)*c.config.CompactRatio {
		compressed = c.truncateToFit(compressed, maxTokens)
		level = CompressionCompact
		currentTokens = EstimateTokensForPrompt(compressed)
	}

	c.logger.Debugf("compressed: %d -> %d tokens (level: %d)", originalTokens, currentTokens, level)
	return compressed, level, originalTokens, currentTokens
}

// snipToolSections identifies long "Tool:" sections in the prompt string
// and trims them, keeping head and tail portions.
func (c *Compressor) snipToolSections(prompt string) string {
	// Prompts from MessagesPrepareWithThinkingAndGuard use role markers
	// like "tool:" or "Tool:" followed by the content. We look for long
	// sections and snip them by keeping the first and last portion.
	const maxSectionLen = 3000
	lines := strings.Split(prompt, "\n")
	var result strings.Builder
	var sectionBuf strings.Builder
	inToolSection := false

	flushBuffer := func() {
		if sectionBuf.Len() == 0 {
			return
		}
		content := sectionBuf.String()
		// Use rune count for consistent multi-byte text handling.
		runeLen := utf8.RuneCountInString(content)
		if runeLen > maxSectionLen {
			headRunes := maxSectionLen / 2
			tailRunes := maxSectionLen / 4
			head := truncateRunes(content, headRunes)
			tailStart := runeLen - tailRunes
			if tailStart < 0 {
				tailStart = 0
			}
			tail := string([]rune(content)[tailStart:])
			result.WriteString(head)
			fmt.Fprintf(&result, "\n... [%d chars snipped] ...\n", runeLen-headRunes-tailRunes)
			result.WriteString(tail)
		} else {
			result.WriteString(content)
		}
		sectionBuf.Reset()
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "tool:") || strings.HasPrefix(lower, "tool output:") {
			flushBuffer()
			inToolSection = true
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}
		// Heuristic: tool section ends at next role marker
		if inToolSection && (strings.HasPrefix(lower, "user:") || strings.HasPrefix(lower, "assistant:") ||
			strings.HasPrefix(lower, "system:") || strings.HasPrefix(lower, "tool:") || strings.HasPrefix(lower, "tool output:")) {
			flushBuffer()
			inToolSection = false
		}
		if inToolSection {
			sectionBuf.WriteString(line)
			sectionBuf.WriteString("\n")
		} else {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	flushBuffer()
	return result.String()
}

// pruneOldSections replaces old tool result sections with placeholders.
// Keeps the most recent ~50% of lines intact; older tool sections are elided.
func (c *Compressor) pruneOldSections(prompt string) string {
	lines := strings.Split(prompt, "\n")
	totalLines := len(lines)
	if totalLines < 20 {
		return prompt
	}

	// pruneEnd marks the boundary: lines before this index are "old" and
	// eligible for pruning; lines at or after this index are "recent" and kept.
	pruneEnd := int(float64(totalLines) * 0.5)
	var result strings.Builder
	inToolSection := false
	toolSectionLines := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "tool:") || strings.HasPrefix(lower, "tool output:") {
			if i < pruneEnd {
				inToolSection = true
				toolSectionLines = 0
				result.WriteString(line)
				result.WriteString("\n")
				continue
			}
			inToolSection = true
			toolSectionLines = 0
			result.WriteString(line)
			result.WriteString("\n")
			result.WriteString("[elided tool result]\n")
			continue
		}
		if inToolSection {
			if strings.HasPrefix(lower, "user:") || strings.HasPrefix(lower, "assistant:") ||
				strings.HasPrefix(lower, "system:") || strings.HasPrefix(lower, "tool:") || strings.HasPrefix(lower, "tool output:") {
				inToolSection = false
				result.WriteString(line)
				result.WriteString("\n")
				continue
			}
			toolSectionLines++
			if i < pruneEnd {
				// Skip old tool section content, already wrote placeholder
				continue
			}
		}
		if !inToolSection {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	return result.String()
}

// CompressMessages compresses a list of message maps.
func (c *Compressor) CompressMessages(messages []map[string]any) *CompressedMessages {
	if !c.config.Enabled {
		return &CompressedMessages{
			Messages:       messages,
			Level:          CompressionNone,
			OriginalTokens: EstimateTokensForMessages(messages),
			CurrentTokens:  EstimateTokensForMessages(messages),
		}
	}

	originalTokens := EstimateTokensForMessages(messages)
	result := c.pruner.Compress(messages)
	result.OriginalTokens = originalTokens

	c.logger.Debugf("compressed messages: %d -> %d tokens (level: %d)", originalTokens, result.CurrentTokens, result.Level)
	return result
}

// CompressAnyMessages compresses a list of messages in the []any format
// used by the prompt builder. Each element must be a map[string]any.
func (c *Compressor) CompressAnyMessages(messages []any) ([]any, CompressionLevel) {
	if !c.config.Enabled || len(messages) == 0 {
		return messages, CompressionNone
	}
	msgMaps := make([]map[string]any, 0, len(messages))
	for _, m := range messages {
		if mm, ok := m.(map[string]any); ok {
			msgMaps = append(msgMaps, mm)
		}
	}
	if len(msgMaps) == 0 {
		return messages, CompressionNone
	}
	result := c.CompressMessages(msgMaps)
	if result.Level == CompressionNone {
		return messages, CompressionNone
	}
	out := make([]any, len(result.Messages))
	for i, m := range result.Messages {
		out[i] = m
	}
	return out, result.Level
}

// NeedsCompression checks if the text exceeds the configured ratio.
func (c *Compressor) NeedsCompression(text string) bool {
	if !c.config.Enabled {
		return false
	}
	tokens := EstimateTokensForPrompt(text)
	return float64(tokens) > float64(c.config.ContextWindow)*c.config.SnipRatio
}

// GetTokenCount returns estimated token count for text.
func (c *Compressor) GetTokenCount(text string) int {
	return EstimateTokensForPrompt(text)
}

// truncateToFit truncates the prompt to fit within maxTokens.
func (c *Compressor) truncateToFit(prompt string, maxTokens int) string {
	// Simple truncation: keep the beginning and end
	// More sophisticated truncation would need to understand message boundaries
	currentTokens := EstimateTokensForPrompt(prompt)
	if currentTokens <= maxTokens {
		return prompt
	}

	// Keep 60% from beginning, 40% from end
	// This is a rough heuristic - better to keep the system prompt and recent messages
	ratio := float64(maxTokens) / float64(currentTokens)
	runeLen := utf8.RuneCountInString(prompt)
	keepRunes := int(float64(runeLen) * ratio * 0.9) // 10% safety margin

	headRunes := int(float64(keepRunes) * 0.6)
	tailRunes := keepRunes - headRunes

	if headRunes+tailRunes >= runeLen {
		return prompt
	}

	head := truncateRunes(prompt, headRunes)
	tailStart := runeLen - tailRunes
	if tailStart < 0 {
		tailStart = 0
	}
	tail := string([]rune(prompt)[tailStart:])

	return head + "\n\n... [content truncated to fit context window] ...\n\n" + tail
}

// Config returns the current compression config.
func (c *Compressor) Config() CompressionConfig {
	return c.config
}
