package taskparser

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrPathContainsNullByte is returned when a path expansion contains a null byte.
var ErrPathContainsNullByte = errors.New("path contains null byte")

// ExpandOptions controls optional behavior of content expansion.
type ExpandOptions struct {
	// SkipCommands disables !`cmd` shell execution; the literal !`cmd` text is preserved in output.
	SkipCommands bool
	// PathRefs, if non-nil, is appended with each successfully resolved @path reference.
	PathRefs *[]string
}

// Expand performs all types of expansion on the content in a single pass:
// 1. Parameter expansion: ${param_name}
// 2. Command expansion: !`command`
// 3. Path expansion: @path
// SECURITY: Processes rune-by-rune to prevent injection attacks where expanded
// content contains further expansion sequences (e.g., command output with ${param}).
func (p Params) Expand(content string) (string, error) {
	return p.ExpandWith(content, ExpandOptions{})
}

// ExpandWith is like Expand but accepts options for lint/dry-run mode.
func (p Params) ExpandWith(content string, opts ExpandOptions) (string, error) {
	var result strings.Builder

	result.Grow(len(content))

	runes := []rune(content)

	for i := 0; i < len(runes); {
		if val, newI, ok := tryExpandParam(runes, i, p); ok {
			result.WriteString(val)

			i = newI

			continue
		}

		if opts.SkipCommands {
			if val, newI, ok := trySkipCommand(runes, i); ok {
				result.Write(val)

				i = newI

				continue
			}
		} else {
			if val, newI, ok := tryExpandCommand(runes, i); ok {
				result.Write(val)

				i = newI

				continue
			}
		}

		if runes[i] == '@' && (i == 0 || isWhitespaceRune(runes[i-1])) {
			if pathContent, newI, ok := tryExpandPath(runes, i, opts.PathRefs); ok {
				result.Write(pathContent)

				i = newI

				continue
			}
		}

		result.WriteRune(runes[i])
		i++
	}

	return result.String(), nil
}

// tryExpandParam attempts ${param} expansion at position i.
// Returns the expanded value, the new index past the expansion, and true if matched.
func tryExpandParam(runes []rune, i int, p Params) (string, int, bool) {
	const prefixLen = 2
	if i+prefixLen >= len(runes) || runes[i] != '$' || runes[i+1] != '{' {
		return "", 0, false
	}

	end := i + prefixLen
	for end < len(runes) && runes[end] != '}' {
		end++
	}

	if end >= len(runes) {
		return "", 0, false
	}

	paramName := string(runes[i+2 : end])
	if val, ok := p.Lookup(paramName); ok {
		return val, end + 1, true
	}

	return string(runes[i : end+1]), end + 1, true
}

// tryExpandCommand attempts !`command` expansion at position i.
// Returns the command output, the new index past the expansion, and true if matched.
func tryExpandCommand(runes []rune, i int) ([]byte, int, bool) {
	const prefixLen = 2
	if i+prefixLen >= len(runes) || runes[i] != '!' || runes[i+1] != '`' {
		return nil, 0, false
	}

	end := i + prefixLen
	for end < len(runes) && runes[end] != '`' {
		end++
	}

	if end >= len(runes) {
		return nil, 0, false
	}

	command := string(runes[i+prefixLen : end])
	// #nosec G204 -- slash command expansion is an intentional feature; commands come from task content
	//nolint:noctx // Expand has no context; command output is best-effort
	cmd := exec.Command("sh", "-c", command)
	output, _ := cmd.CombinedOutput()

	return output, end + 1, true
}

// tryExpandPathAt attempts to expand @path at the given index.
// Returns the content to write (either file content or original @path), the new index,
// and true if expansion was attempted.
func tryExpandPathAt(runes []rune, i int) ([]byte, int, bool) {
	pathStart := i + 1
	pathEnd := pathStart

pathScan:
	for pathEnd < len(runes) {
		switch {
		case pathEnd+1 < len(runes) && runes[pathEnd] == '\\' && runes[pathEnd+1] == ' ':
			pathEnd += 2
		case isWhitespaceRune(runes[pathEnd]):
			break pathScan
		default:
			pathEnd++
		}
	}

	if pathEnd <= pathStart {
		return nil, 0, false
	}

	path := unescapePath(string(runes[pathStart:pathEnd]))
	if err := ValidatePath(path); err != nil {
		return []byte(string(runes[i:pathEnd])), pathEnd, true
	}

	cleanPath := filepath.Clean(path)

	fileContent, err := os.ReadFile(cleanPath)
	if err != nil {
		return []byte(string(runes[i:pathEnd])), pathEnd, true
	}

	return fileContent, pathEnd, true
}

// trySkipCommand detects !`command` and returns the original literal bytes unchanged (no exec).
// Mirrors the detection logic of tryExpandCommand without executing anything.
func trySkipCommand(runes []rune, i int) ([]byte, int, bool) {
	const prefixLen = 2
	if i+prefixLen >= len(runes) || runes[i] != '!' || runes[i+1] != '`' {
		return nil, 0, false
	}

	end := i + prefixLen
	for end < len(runes) && runes[end] != '`' {
		end++
	}

	if end >= len(runes) {
		return nil, 0, false
	}

	return []byte(string(runes[i : end+1])), end + 1, true
}

// tryExpandPath expands @path at position i, optionally tracking resolved paths.
// If pathRefs is non-nil, the successfully resolved path is appended to it.
func tryExpandPath(runes []rune, i int, pathRefs *[]string) ([]byte, int, bool) {
	if pathRefs == nil {
		return tryExpandPathAt(runes, i)
	}

	fileContent, newI, ok, resolved := tryExpandPathAtTracked(runes, i)
	if ok && resolved != "" {
		*pathRefs = append(*pathRefs, resolved)
	}

	return fileContent, newI, ok
}

// tryExpandPathAtTracked is identical to tryExpandPathAt but also returns the resolved
// cleanPath so callers can record it as a loaded file. resolvedPath is empty if the
// file could not be read (the original @path text is returned as content in that case).
func tryExpandPathAtTracked(runes []rune, i int) ([]byte, int, bool, string) {
	pathStart := i + 1
	pathEnd := pathStart

pathScan:
	for pathEnd < len(runes) {
		switch {
		case pathEnd+1 < len(runes) && runes[pathEnd] == '\\' && runes[pathEnd+1] == ' ':
			pathEnd += 2
		case isWhitespaceRune(runes[pathEnd]):
			break pathScan
		default:
			pathEnd++
		}
	}

	if pathEnd <= pathStart {
		return nil, 0, false, ""
	}

	path := unescapePath(string(runes[pathStart:pathEnd]))
	if err := ValidatePath(path); err != nil {
		return []byte(string(runes[i:pathEnd])), pathEnd, true, ""
	}

	cleanPath := filepath.Clean(path)

	fileContent, err := os.ReadFile(cleanPath)
	if err != nil {
		return []byte(string(runes[i:pathEnd])), pathEnd, true, ""
	}

	return fileContent, pathEnd, true, cleanPath
}

// isWhitespaceRune checks if a rune is whitespace (space, tab, newline, carriage return).
func isWhitespaceRune(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// unescapePath removes escape sequences from a path (specifically \<space>).
func unescapePath(path string) string {
	return strings.ReplaceAll(path, "\\ ", " ")
}

// ValidatePath validates a file path for basic safety checks.
// Note: This tool is designed to work with user-created markdown files in their
// workspace and grants read access to files the user can read. The primary
// defense is that users should only use trusted markdown files.
func ValidatePath(path string) error {
	// Check for null bytes which are never valid in file paths
	if strings.Contains(path, "\x00") {
		return ErrPathContainsNullByte
	}

	// We intentionally allow paths with .. components as they may be
	// legitimate references to files in parent directories within the
	// user's workspace. The security model is that users should only
	// use trusted markdown files, similar to running trusted shell scripts.

	return nil
}
