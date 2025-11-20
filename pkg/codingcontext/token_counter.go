package codingcontext

import (
	"unicode/utf8"
)

// estimateTokens estimates the number of LLM tokens in the given text.
// Uses a simple heuristic of approximately 4 characters per token,
// which is a common approximation for English text with GPT-style tokenizers.
func estimateTokens(text string) int {
	charCount := utf8.RuneCountInString(text)
	// Approximate: 1 token ≈ 4 characters
	return charCount / 4
}
