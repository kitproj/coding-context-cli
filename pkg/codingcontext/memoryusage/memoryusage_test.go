package memoryusage_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/memoryusage"
)

func TestReadCurrent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileContent string
		wantBytes   int64
		wantErr     bool
	}{
		{
			name:        "valid memory value",
			fileContent: "12345678\n",
			wantBytes:   12345678,
		},
		{
			name:        "valid value without newline",
			fileContent: "999999",
			wantBytes:   999999,
		},
		{
			name:        "zero value",
			fileContent: "0\n",
			wantBytes:   0,
		},
		{
			name:        "invalid non-numeric content",
			fileContent: "abc\n",
			wantErr:     true,
		},
		{
			name:        "empty content",
			fileContent: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			memFile := filepath.Join(tmpDir, "memory.current")

			if err := os.WriteFile(memFile, []byte(tt.fileContent), 0o600); err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}

			got, err := memoryusage.ReadCurrentFromPath(memFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadCurrentFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantBytes {
				t.Errorf("ReadCurrentFromPath() = %d, want %d", got, tt.wantBytes)
			}
		})
	}
}

func TestReadCurrent_FileNotFound(t *testing.T) {
	t.Parallel()

	_, err := memoryusage.ReadCurrentFromPath("/nonexistent/path/memory.current")
	if err == nil {
		t.Error("ReadCurrentFromPath() expected error for missing file, got nil")
	}
}

func TestReadCurrent_LiveCgroup(t *testing.T) {
	t.Parallel()

	bytes, err := memoryusage.ReadCurrent()
	if err != nil {
		// On systems without cgroup v2 memory.current this is expected.
		t.Skipf("cgroup v2 memory.current not available: %v", err)
	}

	if bytes < 0 {
		t.Errorf("ReadCurrent() returned negative value %d", bytes)
	}

	t.Logf("current cgroup memory usage: %s bytes", strconv.FormatInt(bytes, 10))
}
