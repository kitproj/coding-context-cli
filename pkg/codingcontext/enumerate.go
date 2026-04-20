package codingcontext

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiscoveredTask represents a task found during enumeration of search paths.
type DiscoveredTask struct {
	// Name is the task name as passed to Lint or Run (e.g. "my-task" or "myteam/my-task").
	Name string
	// Path is the absolute path to the task markdown file.
	Path string
	// Namespace is the namespace prefix; empty for global tasks.
	Namespace string
}

// ListTasks enumerates all available tasks from the configured search paths without
// running any of them. It resolves remote directories (same as Run) and scans both
// global (.agents/tasks/) and namespace-specific (.agents/namespaces/<ns>/tasks/)
// task directories.
//
// If the same task name appears in multiple search paths the first occurrence wins
// (consistent with how Run/Lint resolve tasks).
func (cc *Context) ListTasks(ctx context.Context) ([]DiscoveredTask, error) {
	manifestPaths, err := cc.parseManifestFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest file: %w", err)
	}

	for _, p := range manifestPaths {
		cc.searchPaths = append(cc.searchPaths, SearchPath{Path: p})
	}

	if err := cc.downloadRemoteDirectories(ctx); err != nil {
		return nil, fmt.Errorf("failed to download remote directories: %w", err)
	}

	defer cc.cleanupDownloadedDirectories()

	var tasks []DiscoveredTask

	seen := make(map[string]bool)

	for _, sp := range cc.downloadedPaths {
		// Global tasks.
		dir := sp.Path
		for _, taskDir := range taskSearchPaths(dir) {
			found, err := listTasksInDir(taskDir, "")
			if err != nil {
				return nil, fmt.Errorf("failed to list tasks in %s: %w", taskDir, err)
			}

			for _, t := range found {
				if !seen[t.Name] {
					tasks = append(tasks, t)
					seen[t.Name] = true
				}
			}
		}

		// Namespace tasks: walk .agents/namespaces/<ns>/tasks/.
		nsRootDir := filepath.Join(dir, ".agents/namespaces")

		entries, err := os.ReadDir(nsRootDir)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read namespace directory %s: %w", nsRootDir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			ns := entry.Name()
			nsTaskDir := filepath.Join(nsRootDir, ns, "tasks")

			found, err := listTasksInDir(nsTaskDir, ns)
			if err != nil {
				return nil, fmt.Errorf("failed to list namespace tasks in %s: %w", nsTaskDir, err)
			}

			for _, t := range found {
				if !seen[t.Name] {
					tasks = append(tasks, t)
					seen[t.Name] = true
				}
			}
		}
	}

	return tasks, nil
}

// listTasksInDir scans dir for .md/.mdc task files and returns a DiscoveredTask for each.
// namespace is empty for global tasks; for namespaced tasks it is prepended to the name.
func listTasksInDir(dir, namespace string) ([]DiscoveredTask, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat directory %s: %w", dir, err)
	}

	var tasks []DiscoveredTask

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".md" && ext != ".mdc" {
			return nil
		}

		baseName := strings.TrimSuffix(filepath.Base(path), ext)

		name := baseName
		if namespace != "" {
			name = namespace + "/" + baseName
		}

		tasks = append(tasks, DiscoveredTask{
			Name:      name,
			Path:      path,
			Namespace: namespace,
		})

		return nil
	})

	return tasks, err
}
