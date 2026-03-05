// Package tokencount provides token estimation for LLM text.
package tokencount

import (
	"unicode/utf8"
)

// CharsPerToken is the approximate number of characters per token for GPT-style tokenizers.
const CharsPerToken = 4

// EstimateTokens estimates the number of LLM tokens in the given text.
// Uses a simple heuristic of approximately 4 characters per token,
// which is a common approximation for English text with GPT-style tokenizers.
func EstimateTokens(text string) int {
	charCount := utf8.RuneCountInString(text)

	return charCount / CharsPerToken
}
