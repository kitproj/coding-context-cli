package main

import (
	"strings"
)

// RulePath represents a rule path with normalization
// Format: "source_path:normalized_path"
type RulePath string

// Source returns the source path part
func (rp RulePath) Source() string {
	parts := strings.SplitN(string(rp), ":", 2)
	return parts[0]
}

// Normalized returns the normalized path part
func (rp RulePath) Normalized() string {
	parts := strings.SplitN(string(rp), ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	// If no normalized path specified, use source as normalized
	return parts[0]
}

// NewRulePath creates a RulePath from source and normalized paths
func NewRulePath(source, normalized string) RulePath {
	return RulePath(source + ":" + normalized)
}
