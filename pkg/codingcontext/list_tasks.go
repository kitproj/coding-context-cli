package codingcontext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	maxDescriptionLength = 100
	truncationSuffix     = "..."
)

// TaskInfo contains metadata about a discovered task
type TaskInfo struct {
	TaskName    string
	Path        string
	Description string
	Selectors   map[string]interface{}
	Agent       string
	Language    string
	Resume      bool
}

// ListTasks discovers and returns all available tasks from the standard search paths
func (cc *Context) ListTasks(ctx context.Context) ([]TaskInfo, error) {
	if err := cc.downloadRemoteDirectories(ctx); err != nil {
		return nil, fmt.Errorf("failed to download remote directories: %w", err)
	}
	defer cc.cleanupDownloadedDirectories()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Build the list of task search paths (local and remote)
	taskSearchDirs := AllTaskSearchPaths(cc.workDir, homeDir)

	// Add downloaded remote directories to task search paths
	for _, dir := range cc.downloadedDirs {
		taskSearchDirs = append(taskSearchDirs, DownloadedTaskSearchPaths(dir)...)
	}

	// Map to track unique tasks by task_name
	taskMap := make(map[string]TaskInfo)

	for _, dir := range taskSearchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("failed to stat task dir %s: %w", dir, err)
		}

		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if filepath.Ext(path) != ".md" {
				return nil
			}

			// Parse frontmatter
			var frontmatter TaskFrontMatter
			content, err := ParseMarkdownFile(path, &frontmatter)
			if err != nil {
				cc.logger.Warn("Failed to parse task file", "path", path, "error", err)
				return nil // Skip this file but continue
			}

			// Extract task_name from frontmatter Content map
			taskNameVal, hasTaskName := frontmatter.Content["task_name"]
			if !hasTaskName {
				return nil // Skip files without task_name
			}

			taskName, ok := taskNameVal.(string)
			if !ok {
				return nil // Skip if task_name is not a string
			}

			// Extract description from content (first paragraph or heading)
			markdownText := content.Content
			description := extractDescription(markdownText)

			// Extract language from frontmatter (can be in Content or Languages field)
			var language string
			if len(frontmatter.Languages) > 0 {
				language = frontmatter.Languages[0]
			} else if langVal, ok := frontmatter.Content["language"]; ok {
				if langStr, ok := langVal.(string); ok {
					language = langStr
				}
			}

			// Create a unique key that includes selectors for variant tasks
			taskKey := taskName
			if len(frontmatter.Selectors) > 0 || frontmatter.Resume {
				// Add selectors to make key unique for variants
				selectorKeys := make([]string, 0, len(frontmatter.Selectors))
				for k := range frontmatter.Selectors {
					selectorKeys = append(selectorKeys, k)
				}
				sort.Strings(selectorKeys)

				parts := []string{taskName}
				for _, k := range selectorKeys {
					parts = append(parts, fmt.Sprintf("%s=%v", k, frontmatter.Selectors[k]))
				}
				if frontmatter.Resume {
					parts = append(parts, "resume=true")
				}
				taskKey = strings.Join(parts, ",")
			}

			// Store task info (will overwrite if duplicate, keeping last one found)
			taskMap[taskKey] = TaskInfo{
				TaskName:    taskName,
				Path:        path,
				Description: description,
				Selectors:   frontmatter.Selectors,
				Agent:       frontmatter.Agent,
				Language:    language,
				Resume:      frontmatter.Resume,
			}

			return nil
		}); err != nil {
			return nil, fmt.Errorf("failed to walk task dir %s: %w", dir, err)
		}
	}

	// Convert map to sorted slice
	tasks := make([]TaskInfo, 0, len(taskMap))
	for _, task := range taskMap {
		tasks = append(tasks, task)
	}

	// Sort by task name, then by selectors
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].TaskName != tasks[j].TaskName {
			return tasks[i].TaskName < tasks[j].TaskName
		}
		// If same task name, sort by whether it's a resume variant
		if tasks[i].Resume != tasks[j].Resume {
			return !tasks[i].Resume // non-resume comes first
		}
		// Then by number of selectors (simpler variants first)
		return len(tasks[i].Selectors) < len(tasks[j].Selectors)
	})

	return tasks, nil
}

// extractDescription extracts a brief description from the task content
// It looks for the first paragraph or heading after frontmatter
func extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	var description strings.Builder
	inCodeBlock := false
	foundContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines at the start
		if !foundContent && trimmed == "" {
			continue
		}

		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		// If we hit a heading, extract it and stop
		if strings.HasPrefix(trimmed, "#") {
			heading := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
			return heading
		}

		// If we have content, accumulate until we hit an empty line
		if trimmed != "" {
			if foundContent {
				description.WriteString(" ")
			}
			description.WriteString(trimmed)
			foundContent = true
		} else if foundContent {
			// Stop at first empty line after finding content
			break
		}
	}

	result := description.String()
	// Truncate if too long
	if len(result) > maxDescriptionLength {
		truncateAt := maxDescriptionLength - len(truncationSuffix)
		result = result[:truncateAt] + truncationSuffix
	}
	return result
}
