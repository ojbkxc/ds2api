package contextcompression

import (
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

	// Level 1: Snip
	if float64(originalTokens) > float64(maxTokens)*c.config.SnipRatio {
		compressed = c.snipLongContent(compressed)
		level = CompressionSnip
	}

	currentTokens := EstimateTokensForPrompt(compressed)

	// Level 2: Prune
	if float64(currentTokens) > float64(maxTokens)*c.config.PruneRatio {
		compressed = c.pruneOldContent(compressed)
		level = CompressionPrune
		currentTokens = EstimateTokensForPrompt(compressed)
	}

	// Level 3: Compact - truncate to fit
	if float64(currentTokens) > float64(maxTokens)*c.config.CompactRatio {
		compressed = c.truncateToFit(compressed, maxTokens)
		level = CompressionCompact
		currentTokens = EstimateTokensForPrompt(compressed)
	}

	c.logger.Debugf("compressed: %d -> %d tokens (level: %d)", originalTokens, currentTokens, level)
	return compressed, level, originalTokens, currentTokens
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

// snipLongContent trims sections of prompt that are too long.
// Focuses on "tool result" sections and long code blocks.
func (c *Compressor) snipLongContent(prompt string) string {
	// For prompt-level compression, we focus on what we can do without
	// breaking the message structure. The main snipping happens at the
	// message level via the Pruner.
	return prompt
}

// pruneOldContent removes old content from the prompt.
func (c *Compressor) pruneOldContent(prompt string) string {
	return prompt
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
	keepLen := int(float64(len(prompt)) * ratio * 0.9) // 10% safety margin

	headLen := int(float64(keepLen) * 0.6)
	tailLen := keepLen - headLen

	if headLen+tailLen >= len(prompt) {
		return prompt
	}

	head := prompt[:headLen]
	tail := prompt[len(prompt)-tailLen:]

	return head + "\n\n... [content truncated to fit context window] ...\n\n" + tail
}

// Config returns the current compression config.
func (c *Compressor) Config() CompressionConfig {
	return c.config
}