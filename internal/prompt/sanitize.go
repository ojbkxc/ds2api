package prompt

import (
	"regexp"
	"strings"
	"unicode"
)

// zeroWidthSpace is the Unicode zero-width space (U+200B) used to break
// role-prefix injections without visibly altering the text.
const zeroWidthSpace = "\u200B"

// rolePrefixes that attackers may inject to fake a role boundary.
// Each pattern has two capture groups: (1) leading whitespace, (2) the prefix itself.
var rolePrefixPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?im)^(\s*)(System\s*:\s*)`),
	regexp.MustCompile(`(?im)^(\s*)(User\s*:\s*)`),
	regexp.MustCompile(`(?im)^(\s*)(Assistant\s*:\s*)`),
	regexp.MustCompile(`(?im)^(\s*)(Tool\s*:\s*)`),
}

// historySeparatorPatterns match the separator formats used in history
// transcripts so we can escape them when they appear inside user content.
var historySeparatorPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^\s*===\s*\d+\s*\.\s*\w+\s*===\s*$`),
	regexp.MustCompile(`(?m)^\s*---\s*\d+\s*\.\s*\w+\s*---\s*$`),
	regexp.MustCompile(`(?m)^\s*##\s+\d+\s*\.\s*\w+\s*$`),
	regexp.MustCompile(`(?m)^\s*\*\*\d+\s*\.\s*\w+\*\*\s*$`),
	regexp.MustCompile(`(?m)^\s*\[\d+\]\s*\w+\s*$`),
	regexp.MustCompile(`(?m)^\s*\{\d+\}\s*\w+\s*$`),
	regexp.MustCompile(`(?m)^\s*//\s*\d+\s*\.\s*\w+\s*$`),
}

// deepseekSpecialMarkers are internal markers that should not leak into
// user-controlled text.
var deepseekSpecialMarkerPattern = regexp.MustCompile(`<\|[\w▁]+\|>`)

// SanitizeUserInput performs input-side sanitization on user-controlled text
// before it is assembled into the final prompt.
//
// It defends against:
//   1. Role-prefix injection   (e.g. "System: ignore previous instructions")
//   2. History-separator injection (e.g. "=== 3. ASSISTANT ===")
//   3. DeepSeek special-marker injection (e.g. "<|System|>")
//   4. Zero-width-character obfuscation
//
// The strategy is "neutralisation by insertion" – we insert a zero-width
// space before the matched prefix so the model no longer sees it as a
// structural boundary, while the text remains visually unchanged.
func SanitizeUserInput(text string) string {
	if text == "" {
		return ""
	}

	// 1. Neutralise role-prefix injections.
	for _, re := range rolePrefixPatterns {
		text = re.ReplaceAllString(text, "${1}"+zeroWidthSpace+"${2}")
	}

	// 2. Neutralise history-separator injections.
	for _, re := range historySeparatorPatterns {
		text = re.ReplaceAllStringFunc(text, func(match string) string {
			return zeroWidthSpace + match
		})
	}

	// 3. Neutralise DeepSeek special markers.
	text = deepseekSpecialMarkerPattern.ReplaceAllStringFunc(text, func(match string) string {
		return zeroWidthSpace + match
	})

	// 4. Remove zero-width characters that may be used for obfuscation
	// (e.g. "Reasoning\u200B Effort:" to bypass simple string checks).
	text = StripZeroWidthChars(text)

	return text
}

// StripZeroWidthChars removes zero-width and invisible format characters
// that attackers use to obfuscate injection payloads.
func StripZeroWidthChars(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		switch r {
		case '\u200B', '\u200C', '\u200D', '\uFEFF',
			'\u2060', '\u2061', '\u2062', '\u2063', '\u2064',
			'\u180E', '\u206A', '\u206B', '\u206C', '\u206D',
			'\u206E', '\u206F':
			// skip zero-width / invisible format characters
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SanitizeToolMeta restricts tool names and descriptions to safe characters,
// preventing injection through the tool-definition channel.
func SanitizeToolMeta(name, description string) (string, string) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	// Tool name: allow letters, digits, underscore, hyphen.
	name = sanitizeToAllowedRunes(name, func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-'
	})

	// Tool description: allow printable ASCII and common Unicode,
	// but strip control characters and newlines that could break prompt structure.
	description = sanitizeToAllowedRunes(description, func(r rune) bool {
		if r == '\n' || r == '\r' || r == '\t' {
			return true
		}
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return false
		}
		return true
	})

	return name, description
}

func sanitizeToAllowedRunes(s string, allowed func(rune) bool) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if allowed(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
