package codingcontext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
)

// LoadedFileKind identifies the role of a file loaded during context assembly.
type LoadedFileKind string

// Valid LoadedFileKind values.
const (
	LoadedFileKindTask      LoadedFileKind = "task"
	LoadedFileKindRule      LoadedFileKind = "rule"
	LoadedFileKindCommand   LoadedFileKind = "command"
	LoadedFileKindSkill     LoadedFileKind = "skill"
	LoadedFileKindPathRef   LoadedFileKind = "path-ref"
	LoadedFileKindBootstrap LoadedFileKind = "bootstrap"
)

// LintErrorKind identifies the category of a structural problem.
type LintErrorKind string

// Valid LintErrorKind values.
const (
	LintErrorKindParse           LintErrorKind = "parse"
	LintErrorKindMissingCommand  LintErrorKind = "missing-command"
	LintErrorKindSkillValidation LintErrorKind = "skill-validation"
	LintErrorKindSelectorNoMatch LintErrorKind = "selector-no-match"
)

// LoadedFile records a file accessed during context assembly.
type LoadedFile struct {
	Path string
	Kind LoadedFileKind
}

// LintError records a non-fatal structural problem found during linting.
type LintError struct {
	Path    string // May be empty
	Kind    LintErrorKind
	Message string
	Line    int // 1-indexed; 0 means unknown (only set for parse errors)
	Column  int // 1-indexed; 0 means unknown (only set for parse errors)
}

// LintResult is returned by Lint(). It embeds the assembled Result plus tracking
// data collected during the dry run.
type LintResult struct {
	*Result

	LoadedFiles []LoadedFile
	Errors      []LintError
}

// lintCollector is internal state attached to Context during lint mode.
type lintCollector struct {
	files  []LoadedFile
	errors []LintError
	// allFrontmatterValues tracks key→value pairs seen across ALL discovered markdown
	// files (matched or skipped), used to validate selector coverage after assembly.
	allFrontmatterValues map[string]map[string]bool
}

func (lc *lintCollector) recordFile(path string, kind LoadedFileKind) {
	lc.files = append(lc.files, LoadedFile{Path: path, Kind: kind})
}

func (lc *lintCollector) recordError(path string, kind LintErrorKind, message string) {
	lc.errors = append(lc.errors, LintError{Path: path, Kind: kind, Message: message})
}

func (lc *lintCollector) recordParseError(pe *markdown.ParseError) {
	lc.errors = append(lc.errors, LintError{
		Path:    pe.File,
		Kind:    LintErrorKindParse,
		Message: pe.Message,
		Line:    pe.Line,
		Column:  pe.Column,
	})
}

func (lc *lintCollector) recordFrontmatterValues(fm markdown.BaseFrontMatter) {
	if lc.allFrontmatterValues == nil {
		lc.allFrontmatterValues = make(map[string]map[string]bool)
	}

	for key, value := range fm.Content {
		if lc.allFrontmatterValues[key] == nil {
			lc.allFrontmatterValues[key] = make(map[string]bool)
		}

		switch v := value.(type) {
		case []any:
			for _, item := range v {
				lc.allFrontmatterValues[key][fmt.Sprint(item)] = true
			}
		default:
			lc.allFrontmatterValues[key][fmt.Sprint(v)] = true
		}
	}
}

// Lint runs context assembly in dry-run mode and returns validation results.
// It skips bootstrap script execution and shell command expansion (!`cmd`), but
// otherwise performs the same file loading, parsing, and selector matching as Run().
// Fatal errors (e.g. task not found) are returned as errors; structural problems
// are collected in LintResult.Errors.
func (cc *Context) Lint(ctx context.Context, taskName string) (*LintResult, error) {
	cc.lintMode = true
	cc.lintCollector = &lintCollector{}
	cc.doBootstrap = true // ensure rule + skill discovery runs

	result, err := cc.Run(ctx, taskName)
	if err != nil {
		return nil, err
	}

	cc.validateSelectorCoverage()

	return &LintResult{
		Result:      result,
		LoadedFiles: cc.lintCollector.files,
		Errors:      cc.lintCollector.errors,
	}, nil
}

// recordLintBootstrap logs or records a bootstrap entry during lint mode without executing it.
// For frontmatter-based bootstraps, it just logs. For file-based bootstraps, it stat-checks
// the companion script and records it in LoadedFiles if it exists.
func (cc *Context) recordLintBootstrap(rulePath, frontmatterBootstrap string) {
	if frontmatterBootstrap != "" {
		cc.logger.Info("Lint mode: skipping frontmatter bootstrap", "path", rulePath)

		return
	}

	bootstrapFilePath := strings.TrimSuffix(rulePath, filepath.Ext(rulePath)) + "-bootstrap"
	if _, err := os.Stat(bootstrapFilePath); err == nil && cc.lintCollector != nil {
		cc.lintCollector.recordFile(bootstrapFilePath, LoadedFileKindBootstrap)
	}
}

// validateSelectorCoverage checks that each user-specified selector key-value pair
// appeared in at least one discovered file's frontmatter. Records a LintError for
// each unmatched pair. Auto-set keys (task_name, resume) are excluded from this check.
func (cc *Context) validateSelectorCoverage() {
	// Anchor selector errors to the task file so they appear as file annotations.
	taskPath := ""
	for _, f := range cc.lintCollector.files {
		if f.Kind == LoadedFileKindTask {
			taskPath = f.Path
			break
		}
	}

	autoKeys := map[string]bool{"task_name": true, "resume": true, "namespace": true}
	for key, values := range cc.includes {
		if autoKeys[key] {
			continue
		}

		seen := cc.lintCollector.allFrontmatterValues[key]
		for value := range values {
			if !seen[value] {
				cc.lintCollector.recordError(taskPath, LintErrorKindSelectorNoMatch,
					fmt.Sprintf("selector '%s=%s' matched no discovered files", key, value))
			}
		}
	}
}
