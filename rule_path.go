package main

import (
	"strings"
)

// RulePath represents a rule path mapping
// Format: "source_path:target_path"
type RulePath string

// SourcePath returns the source path part
func (rp RulePath) SourcePath() string {
	parts := strings.SplitN(string(rp), ":", 2)
	return parts[0]
}

// TargetPath returns the target path part
func (rp RulePath) TargetPath() string {
	parts := strings.SplitN(string(rp), ":", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	// If no target path specified, use source as target
	return parts[0]
}

// NewRulePath creates a RulePath from source and target paths
func NewRulePath(source, target string) RulePath {
	return RulePath(source + ":" + target)
}
