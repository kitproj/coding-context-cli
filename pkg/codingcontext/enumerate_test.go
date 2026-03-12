package codingcontext

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// createTaskMdc creates a task file with .mdc extension.
func createTaskMdc(t *testing.T, dir, name, content string) {
	t.Helper()

	taskPath := filepath.Join(dir, ".agents", "tasks", name+".mdc")
	if err := os.MkdirAll(filepath.Dir(taskPath), 0o750); err != nil {
		t.Fatalf("failed to create task dir: %v", err)
	}

	if err := os.WriteFile(taskPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write task file %s: %v", taskPath, err)
	}
}

// ── listTasksInDir ───────────────────────────────────────────────────────────

func TestListTasksInDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T, dir string)
		taskDir   string
		namespace string
		want      []DiscoveredTask
		wantErr   bool
	}{
		{
			name:    "directory does not exist returns nil",
			taskDir: "/nonexistent/path",
			want:    nil,
		},
		{
			name: "empty directory returns nil",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				taskDir := filepath.Join(dir, ".agents", "tasks")
				if err := os.MkdirAll(taskDir, 0o750); err != nil {
					t.Fatalf("failed to create dir: %v", err)
				}
			},
			taskDir: "", // will be set from tmpDir
			want:    nil,
		},
		{
			name: "single .md task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "single", "", "Task content")
			},
			taskDir: "",
			want: []DiscoveredTask{
				{Name: "single", Path: "", Namespace: ""},
			},
		},
		{
			name: "single .mdc task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTaskMdc(t, dir, "mdc-task", "Mdc task content")
			},
			taskDir: "",
			want: []DiscoveredTask{
				{Name: "mdc-task", Path: "", Namespace: ""},
			},
		},
		{
			name: "multiple tasks",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "a", "", "A")
				createTask(t, dir, "b", "", "B")
				createTask(t, dir, "c", "", "C")
			},
			taskDir: "",
			want: []DiscoveredTask{
				{Name: "a", Path: "", Namespace: ""},
				{Name: "b", Path: "", Namespace: ""},
				{Name: "c", Path: "", Namespace: ""},
			},
		},
		{
			name: "namespace prefix applied",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createNamespaceTask(t, dir, "myteam", "build", "Build task")
			},
			taskDir:   "",
			namespace: "myteam",
			want: []DiscoveredTask{
				{Name: "myteam/build", Path: "", Namespace: "myteam"},
			},
		},
		{
			name: "ignores non-task files",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "valid", "", "Valid task")
				taskDir := filepath.Join(dir, ".agents", "tasks")
				if err := os.WriteFile(filepath.Join(taskDir, "readme.txt"), []byte("text"), 0o600); err != nil {
					t.Fatalf("failed to write txt: %v", err)
				}
				if err := os.WriteFile(filepath.Join(taskDir, "notes.json"), []byte("{}"), 0o600); err != nil {
					t.Fatalf("failed to write json: %v", err)
				}
			},
			taskDir: "",
			want: []DiscoveredTask{
				{Name: "valid", Path: "", Namespace: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			taskDir := tt.taskDir
			if tt.setup != nil {
				tt.setup(t, tmpDir)
				if taskDir == "" {
					taskDir = filepath.Join(tmpDir, ".agents", "tasks")
					// For namespace test, use the namespace task dir
					if tt.namespace != "" {
						taskDir = filepath.Join(tmpDir, ".agents", "namespaces", tt.namespace, "tasks")
					}
				}
			} else if taskDir == "" {
				taskDir = "/nonexistent/path"
			}

			got, err := listTasksInDir(taskDir, tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("listTasksInDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil {
				if got != nil {
					t.Errorf("listTasksInDir() = %v, want nil", got)
				}
				return
			}

			// Normalize: we only check Name and Namespace; Path depends on tmpdir
			if len(got) != len(tt.want) {
				t.Errorf("listTasksInDir() returned %d tasks, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i].Name != tt.want[i].Name {
					t.Errorf("task[%d].Name = %q, want %q", i, got[i].Name, tt.want[i].Name)
				}
				if got[i].Namespace != tt.want[i].Namespace {
					t.Errorf("task[%d].Namespace = %q, want %q", i, got[i].Namespace, tt.want[i].Namespace)
				}
				if got[i].Path != "" && !strings.HasSuffix(got[i].Path, "."+filepath.Ext(got[i].Path)) {
					ext := filepath.Ext(got[i].Path)
					if ext != ".md" && ext != ".mdc" {
						t.Errorf("task[%d].Path has unexpected ext %q", i, ext)
					}
				}
			}
		})
	}
}

func TestListTasksInDir_Subdirectories(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	taskDir := filepath.Join(tmpDir, ".agents", "tasks")
	if err := os.MkdirAll(taskDir, 0o750); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	// Create task in subdirectory (filepath.Walk is recursive)
	subDir := filepath.Join(taskDir, "nested")
	if err := os.MkdirAll(subDir, 0o750); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "nested-task.md"), []byte("Nested"), 0o600); err != nil {
		t.Fatalf("failed to write nested task: %v", err)
	}

	got, err := listTasksInDir(taskDir, "")
	if err != nil {
		t.Fatalf("listTasksInDir() error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 task, got %d", len(got))
	}
	if got[0].Name != "nested-task" {
		t.Errorf("Name = %q, want nested-task", got[0].Name)
	}
}

// ── ListTasks ────────────────────────────────────────────────────────────────

func TestContext_ListTasks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(t *testing.T, dir string)
		opts    []Option
		want    []DiscoveredTask
		wantErr bool
		errSub  string
	}{
		{
			name:  "empty directory returns empty list",
			setup: func(t *testing.T, _ string) { t.Helper() },
			want:  nil,
		},
		{
			name: "single global task",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "deploy", "", "Deploy content")
			},
			want: []DiscoveredTask{
				{Name: "deploy", Namespace: ""},
			},
		},
		{
			name: "multiple global tasks",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "build", "", "Build")
				createTask(t, dir, "test", "", "Test")
				createTask(t, dir, "lint", "", "Lint")
			},
			want: []DiscoveredTask{
				{Name: "build", Namespace: ""},
				{Name: "test", Namespace: ""},
				{Name: "lint", Namespace: ""},
			},
		},
		{
			name: "namespace tasks only",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createNamespaceTask(t, dir, "myteam", "release", "Release task")
				createNamespaceTask(t, dir, "myteam", "rollback", "Rollback task")
			},
			want: []DiscoveredTask{
				{Name: "myteam/release", Namespace: "myteam"},
				{Name: "myteam/rollback", Namespace: "myteam"},
			},
		},
		{
			name: "global and namespace tasks",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "global-task", "", "Global")
				createNamespaceTask(t, dir, "team-a", "ns-task", "Namespace task")
			},
			want: []DiscoveredTask{
				{Name: "global-task", Namespace: ""},
				{Name: "team-a/ns-task", Namespace: "team-a"},
			},
		},
		{
			name: "multiple namespaces",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createNamespaceTask(t, dir, "team-a", "build", "TeamA build")
				createNamespaceTask(t, dir, "team-b", "build", "TeamB build")
			},
			want: []DiscoveredTask{
				{Name: "team-a/build", Namespace: "team-a"},
				{Name: "team-b/build", Namespace: "team-b"},
			},
		},
		{
			name: "first occurrence wins with multiple search paths",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "dup", "", "First path task")

				secondDir := filepath.Join(dir, "second")
				if err := os.MkdirAll(filepath.Join(secondDir, ".agents", "tasks"), 0o750); err != nil {
					t.Fatalf("failed to create second dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(secondDir, ".agents", "tasks", "dup.md"), []byte("Second path task"), 0o600); err != nil {
					t.Fatalf("failed to write second task: %v", err)
				}
			},
			opts: nil, // overridden in test to use both dirs
			want: []DiscoveredTask{
				{Name: "dup", Namespace: ""},
			},
		},
		{
			name: "tasks have absolute paths",
			setup: func(t *testing.T, dir string) {
				t.Helper()
				createTask(t, dir, "pathcheck", "", "Content")
			},
			want: []DiscoveredTask{
				{Name: "pathcheck", Namespace: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, tmpDir)
			}

			opts := []Option{WithSearchPaths(tmpDir)}
			if tt.name == "first occurrence wins with multiple search paths" {
				opts = []Option{WithSearchPaths(tmpDir, filepath.Join(tmpDir, "second"))}
			} else if tt.opts != nil {
				opts = append(opts, tt.opts...)
			}

			cc := New(opts...)
			got, err := cc.ListTasks(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errSub != "" && err != nil && !strings.Contains(err.Error(), tt.errSub) {
				t.Errorf("ListTasks() error = %v, want substring %q", err, tt.errSub)
				return
			}

			if tt.want == nil {
				if len(got) > 0 {
					t.Errorf("ListTasks() = %v, want empty", got)
				}
				return
			}

			// Build comparable slices (ignore Path for most checks; verify it's set)
			gotNames := make([]DiscoveredTask, 0, len(got))
			for _, task := range got {
				if task.Path == "" {
					t.Errorf("task %q has empty Path", task.Name)
				}
				gotNames = append(gotNames, DiscoveredTask{Name: task.Name, Namespace: task.Namespace})
			}

			wantNames := tt.want
			if len(gotNames) != len(wantNames) {
				t.Errorf("ListTasks() returned %d tasks, want %d: got %+v", len(gotNames), len(wantNames), gotNames)
				return
			}

			// Order may vary; compare as sets
			for _, w := range wantNames {
				found := false
				for _, g := range gotNames {
					if g.Name == w.Name && g.Namespace == w.Namespace {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("ListTasks() missing task %q (ns=%q)", w.Name, w.Namespace)
				}
			}
		})
	}
}

func TestContext_ListTasks_FileProtocol(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createTask(t, tmpDir, "file-task", "", "Content")

	cc := New(WithSearchPaths("file://" + tmpDir))
	got, err := cc.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks() error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 task, got %d", len(got))
	}
	if got[0].Name != "file-task" {
		t.Errorf("Name = %q, want file-task", got[0].Name)
	}
}

func TestContext_ListTasks_FirstOccurrenceWins(t *testing.T) {
	t.Parallel()

	dir1 := t.TempDir()
	dir2 := t.TempDir()

	createTask(t, dir1, "same", "", "First")
	createTask(t, dir2, "same", "", "Second")

	cc := New(WithSearchPaths(dir1, dir2))
	got, err := cc.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks() error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 task (first wins), got %d", len(got))
	}
	if got[0].Name != "same" {
		t.Errorf("Name = %q, want same", got[0].Name)
	}
	// Path should point to first search path
	if !strings.Contains(got[0].Path, dir1) {
		t.Errorf("Path should be from first search path, got %q", got[0].Path)
	}
}

func TestDiscoveredTask_Fields(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	createTask(t, tmpDir, "fields-test", "", "Content")
	createNamespaceTask(t, tmpDir, "team", "ns-task", "NS content")

	cc := New(WithSearchPaths(tmpDir))
	tasks, err := cc.ListTasks(context.Background())
	if err != nil {
		t.Fatalf("ListTasks() error: %v", err)
	}

	// Sort for deterministic comparison
	slices.SortFunc(tasks, func(a, b DiscoveredTask) int {
		return strings.Compare(a.Name, b.Name)
	})

	var global, namespaced *DiscoveredTask
	for i := range tasks {
		if tasks[i].Namespace == "" {
			global = &tasks[i]
		} else {
			namespaced = &tasks[i]
		}
	}

	if global == nil || namespaced == nil {
		t.Fatalf("expected both global and namespaced tasks")
	}

	// Global task
	if global.Name != "fields-test" {
		t.Errorf("global Name = %q, want fields-test", global.Name)
	}
	if global.Namespace != "" {
		t.Errorf("global Namespace = %q, want empty", global.Namespace)
	}
	if global.Path == "" || !strings.HasSuffix(global.Path, "fields-test.md") {
		t.Errorf("global Path = %q, want *fields-test.md", global.Path)
	}

	// Namespaced task
	if namespaced.Name != "team/ns-task" {
		t.Errorf("namespaced Name = %q, want team/ns-task", namespaced.Name)
	}
	if namespaced.Namespace != "team" {
		t.Errorf("namespaced Namespace = %q, want team", namespaced.Namespace)
	}
	if namespaced.Path == "" || !strings.HasSuffix(namespaced.Path, "ns-task.md") {
		t.Errorf("namespaced Path = %q, want *ns-task.md", namespaced.Path)
	}
}

