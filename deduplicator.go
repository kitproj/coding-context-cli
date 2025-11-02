package main

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// RuleDeduplicator handles deduplication of rule content at multiple levels
type RuleDeduplicator struct {
	// L0: Content-based hashing - exact duplicate detection
	contentHashes map[string]bool

	// L1: Structure-aware chunk similarity
	chunks []ruleChunk
}

// ruleChunk represents a chunk of content from a rule file
type ruleChunk struct {
	heading string
	content string
	hash    string
}

// NewRuleDeduplicator creates a new RuleDeduplicator
func NewRuleDeduplicator() *RuleDeduplicator {
	return &RuleDeduplicator{
		contentHashes: make(map[string]bool),
		chunks:        make([]ruleChunk, 0),
	}
}

// AddContent attempts to add content to the deduplicator
// Returns the deduplicated content (may be empty if fully duplicated)
func (d *RuleDeduplicator) AddContent(content string) string {
	// L0: Check for exact content duplicate using hash
	contentHash := hashContent(content)
	if d.contentHashes[contentHash] {
		// Exact duplicate found, skip entirely
		return ""
	}

	// Mark this content as seen
	d.contentHashes[contentHash] = true

	// L1: Split content into chunks by H1 headings and check similarity
	newChunks := splitIntoChunks(content)
	var result strings.Builder

	for _, newChunk := range newChunks {
		// Check if this chunk is similar to any existing chunk
		if d.isChunkSimilar(newChunk) {
			// Similar chunk found, skip it
			continue
		}

		// Add this chunk to our collection
		d.chunks = append(d.chunks, newChunk)

		// Build the result with the chunk
		if newChunk.heading != "" {
			result.WriteString(newChunk.heading)
			result.WriteString("\n")
		}
		if newChunk.content != "" {
			result.WriteString(newChunk.content)
			result.WriteString("\n")
		}
	}

	return result.String()
}

// hashContent generates a SHA256 hash of the content
func hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// splitIntoChunks splits content by H1 headings (# Title)
func splitIntoChunks(content string) []ruleChunk {
	// Handle empty content
	if strings.TrimSpace(content) == "" {
		return []ruleChunk{}
	}

	lines := strings.Split(content, "\n")
	var chunks []ruleChunk
	var currentHeading string
	var currentLines []string

	for _, line := range lines {
		// Check if this is an H1 heading (starts with "# " but not "##")
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			// Save previous chunk if it exists
			if currentHeading != "" || len(currentLines) > 0 {
				chunkContent := strings.Join(currentLines, "\n")
				// Include heading in hash for uniqueness
				fullChunkContent := currentHeading + "\n" + chunkContent
				if currentHeading != "" || strings.TrimSpace(chunkContent) != "" {
					chunks = append(chunks, ruleChunk{
						heading: currentHeading,
						content: chunkContent,
						hash:    hashContent(fullChunkContent),
					})
				}
			}

			// Start new chunk
			currentHeading = line
			currentLines = []string{}
		} else {
			currentLines = append(currentLines, line)
		}
	}

	// Don't forget the last chunk
	if currentHeading != "" || len(currentLines) > 0 {
		chunkContent := strings.Join(currentLines, "\n")
		// Include heading in hash for uniqueness
		fullChunkContent := currentHeading + "\n" + chunkContent
		if currentHeading != "" || strings.TrimSpace(chunkContent) != "" {
			chunks = append(chunks, ruleChunk{
				heading: currentHeading,
				content: chunkContent,
				hash:    hashContent(fullChunkContent),
			})
		}
	}

	return chunks
}

// isChunkSimilar checks if a chunk is similar to any existing chunks
func (d *RuleDeduplicator) isChunkSimilar(newChunk ruleChunk) bool {
	// First check for exact hash match
	for _, existingChunk := range d.chunks {
		if existingChunk.hash == newChunk.hash {
			return true
		}
	}

	// Special case: if content is empty or very short (< 5 chars), don't check similarity
	// Only exact hash matches should deduplicate empty/minimal content
	if len(strings.TrimSpace(newChunk.content)) < 5 {
		return false
	}

	// Then check for similarity using Levenshtein distance
	// Strategy: Use different thresholds based on content length
	for _, existingChunk := range d.chunks {
		// Also skip similarity check if existing chunk has minimal content
		if len(strings.TrimSpace(existingChunk.content)) < 5 {
			continue
		}

		distance := levenshteinDistance(existingChunk.content, newChunk.content)
		maxLen := len(existingChunk.content)
		if len(newChunk.content) > maxLen {
			maxLen = len(newChunk.content)
		}

		// For very short content (< 50 chars), only consider similar if distance <= 2
		// This catches typos but not substantive changes
		if maxLen < 50 {
			if distance <= 2 {
				return true
			}
			continue
		}

		// For medium content (50-200 chars), use 10% threshold
		if maxLen < 200 {
			threshold := float64(maxLen) * 0.10
			if float64(distance) < threshold {
				return true
			}
			continue
		}

		// For longer content (200+ chars), use 20% threshold
		threshold := float64(maxLen) * 0.20
		if float64(distance) < threshold {
			return true
		}
	}

	return false
}

// levenshteinDistance calculates the Levenshtein distance between two strings
// This measures the minimum number of single-character edits needed to change one string into another
func levenshteinDistance(s1, s2 string) int {
	// Handle edge cases
	if s1 == s2 {
		return 0
	}
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create a 2D matrix to store distances
	// We only need to keep two rows at a time for optimization
	prevRow := make([]int, len(s2)+1)
	currRow := make([]int, len(s2)+1)

	// Initialize first row
	for j := 0; j <= len(s2); j++ {
		prevRow[j] = j
	}

	// Calculate distances
	for i := 1; i <= len(s1); i++ {
		currRow[0] = i

		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			// Minimum of:
			// - deletion: currRow[j-1] + 1
			// - insertion: prevRow[j] + 1
			// - substitution: prevRow[j-1] + cost
			currRow[j] = min(
				currRow[j-1]+1,
				min(prevRow[j]+1, prevRow[j-1]+cost),
			)
		}

		// Swap rows
		prevRow, currRow = currRow, prevRow
	}

	return prevRow[len(s2)]
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
