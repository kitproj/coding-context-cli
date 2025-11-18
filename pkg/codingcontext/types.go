package codingcontext

import (
	"io"
	"log/slog"
	"os/exec"
)

// Params is a map of parameter key-value pairs for template substitution
type Params map[string]string

// FrontMatter represents parsed YAML frontmatter from markdown files
type FrontMatter map[string]any

// Selectors stores selector key-value pairs where values are stored in inner maps
// Multiple values for the same key use OR logic (match any value in the inner map)
// Each value can be represented exactly once per key
type Selectors map[string]map[string]bool

// Context holds the configuration and state for assembling coding context
type Context struct {
	workDir             string
	resume              bool
	params              Params
	includes            Selectors
	remotePaths         []string
	emitTaskFrontmatter bool

	downloadedDirs   []string
	matchingTaskFile string
	taskFrontmatter  FrontMatter // Parsed task frontmatter
	taskContent      string      // Parsed task content (before parameter expansion)
	totalTokens      int
	output           io.Writer
	logger           *slog.Logger
	cmdRunner        func(cmd *exec.Cmd) error
}
