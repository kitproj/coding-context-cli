// Package memoryusage provides memory usage reading from cgroup v2.
package memoryusage

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const cgroupV2MemoryCurrentPath = "/sys/fs/cgroup/memory.current"

// ReadCurrent reads the current memory usage in bytes from the cgroup v2
// memory.current file at the default path. Returns an error if the file
// cannot be read or parsed.
func ReadCurrent() (int64, error) {
	return ReadCurrentFromPath(cgroupV2MemoryCurrentPath)
}

// ReadCurrentFromPath reads the current memory usage in bytes from the
// provided cgroup v2 memory.current file path. Returns an error if the
// file cannot be read or parsed.
func ReadCurrentFromPath(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("reading %s: %w", path, err)
	}

	val, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing %s: %w", path, err)
	}

	return val, nil
}
