package codingcontext

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestContext_ListTasks(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "list-tasks-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create task directories
	tasksDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		t.Fatalf("Failed to create tasks dir: %v", err)
	}

	tests := []struct {
		name     string
		files    map[string]string
		expected []TaskInfo
	}{
		{
			name: "single task",
			files: map[string]string{
				"fix-bug.md": `---
task_name: fix-bug
---
# Fix Bug Task
This is a bug fix task.`,
			},
			expected: []TaskInfo{
				{
					TaskName:    "fix-bug",
					Description: "Fix Bug Task",
					Selectors:   nil,
					Resume:      false,
				},
			},
		},
		{
			name: "multiple tasks",
			files: map[string]string{
				"fix-bug.md": `---
task_name: fix-bug
---
# Fix Bug Task`,
				"plan-feature.md": `---
task_name: plan-feature
---
# Feature Planning`,
			},
			expected: []TaskInfo{
				{
					TaskName:    "fix-bug",
					Description: "Fix Bug Task",
					Selectors:   nil,
					Resume:      false,
				},
				{
					TaskName:    "plan-feature",
					Description: "Feature Planning",
					Selectors:   nil,
					Resume:      false,
				},
			},
		},
		{
			name: "task with selectors",
			files: map[string]string{
				"implement-feature.md": `---
task_name: implement-feature
selectors:
  language: go
  stage: implementation
---
# Implement Feature in Go`,
			},
			expected: []TaskInfo{
				{
					TaskName:    "implement-feature",
					Description: "Implement Feature in Go",
					Selectors: map[string]interface{}{
						"language": "go",
						"stage":    "implementation",
					},
					Resume: false,
				},
			},
		},
		{
			name: "task with resume variant",
			files: map[string]string{
				"fix-bug.md": `---
task_name: fix-bug
resume: false
---
# Fix Bug Task`,
				"fix-bug-resume.md": `---
task_name: fix-bug
resume: true
---
# Fix Bug - Resume`,
			},
			expected: []TaskInfo{
				{
					TaskName:    "fix-bug",
					Description: "Fix Bug Task",
					Selectors:   nil,
					Resume:      false,
				},
				{
					TaskName:    "fix-bug",
					Description: "Fix Bug - Resume",
					Selectors:   nil,
					Resume:      true,
				},
			},
		},
		{
			name: "file without task_name is skipped",
			files: map[string]string{
				"fix-bug.md": `---
task_name: fix-bug
---
# Fix Bug`,
				"no-task-name.md": `---
language: go
---
# Some content without task_name`,
			},
			expected: []TaskInfo{
				{
					TaskName:    "fix-bug",
					Description: "Fix Bug",
					Selectors:   nil,
					Resume:      false,
				},
			},
		},
		{
			name: "description from paragraph",
			files: map[string]string{
				"test-task.md": `---
task_name: test-task
---

This is a description paragraph that should be extracted.

More content here.`,
			},
			expected: []TaskInfo{
				{
					TaskName:    "test-task",
					Description: "This is a description paragraph that should be extracted.",
					Selectors:   nil,
					Resume:      false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean the tasks directory
			os.RemoveAll(tasksDir)
			os.MkdirAll(tasksDir, 0755)

			// Create test files
			for filename, content := range tt.files {
				path := filepath.Join(tasksDir, filename)
				if err := os.WriteFile(path, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write test file %s: %v", filename, err)
				}
			}

			// Create context
			cc := New(WithWorkDir(tmpDir))

			// List tasks
			tasks, err := cc.ListTasks(context.Background())
			if err != nil {
				t.Fatalf("ListTasks failed: %v", err)
			}

			// Compare results
			if len(tasks) != len(tt.expected) {
				t.Fatalf("Expected %d tasks, got %d", len(tt.expected), len(tasks))
			}

			for i, expected := range tt.expected {
				actual := tasks[i]
				if actual.TaskName != expected.TaskName {
					t.Errorf("Task %d: expected task_name=%q, got %q", i, expected.TaskName, actual.TaskName)
				}
				if actual.Description != expected.Description {
					t.Errorf("Task %d: expected description=%q, got %q", i, expected.Description, actual.Description)
				}
				if actual.Resume != expected.Resume {
					t.Errorf("Task %d: expected resume=%v, got %v", i, expected.Resume, actual.Resume)
				}
				if expected.Selectors != nil {
					if len(actual.Selectors) != len(expected.Selectors) {
						t.Errorf("Task %d: expected %d selectors, got %d", i, len(expected.Selectors), len(actual.Selectors))
					}
					for k, v := range expected.Selectors {
						if actual.Selectors[k] != v {
							t.Errorf("Task %d: expected selector %s=%v, got %v", i, k, v, actual.Selectors[k])
						}
					}
				}
			}
		})
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "heading",
			content:  "# Fix Bug Task\n\nSome description",
			expected: "Fix Bug Task",
		},
		{
			name:     "paragraph",
			content:  "This is a description.",
			expected: "This is a description.",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "content with leading newlines",
			content:  "\n\n# Heading\n\nContent",
			expected: "Heading",
		},
		{
			name:     "long description truncated",
			content:  "This is a very long description that exceeds one hundred characters and should be truncated appropriately to fit within the limit",
			expected: "This is a very long description that exceeds one hundred characters and should be truncated appro...",
		},
		{
			name: "skip code blocks",
			content: `# Task Name

` + "```" + `
code here
` + "```" + `

Description here.`,
			expected: "Task Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDescription(tt.content)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
