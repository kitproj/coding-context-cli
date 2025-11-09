package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileSystem is an abstraction for accessing files from various sources
type FileSystem interface {
	// Open opens a file for reading
	Open(name string) (fs.File, error)
	// Stat returns file info
	Stat(name string) (fs.FileInfo, error)
	// Walk traverses a directory tree
	Walk(root string, fn filepath.WalkFunc) error
	// IsRemote returns true if this is a remote file system
	IsRemote() bool
}

// LocalFileSystem implements FileSystem for local files
type LocalFileSystem struct{}

func (lfs *LocalFileSystem) Open(name string) (fs.File, error) {
	return os.Open(name)
}

func (lfs *LocalFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (lfs *LocalFileSystem) Walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, fn)
}

func (lfs *LocalFileSystem) IsRemote() bool {
	return false
}

// HTTPFileSystem implements FileSystem for HTTP/HTTPS URLs
type HTTPFileSystem struct {
	baseURL string
	client  *http.Client
}

func NewHTTPFileSystem(baseURL string) *HTTPFileSystem {
	return &HTTPFileSystem{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (hfs *HTTPFileSystem) Open(name string) (fs.File, error) {
	// Convert local path to URL path
	urlPath := filepath.ToSlash(name)
	fullURL := hfs.baseURL + "/" + strings.TrimPrefix(urlPath, "/")

	resp, err := hfs.client.Get(fullURL)
	if err != nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: err}
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, &fs.PathError{Op: "open", Path: name, Err: fmt.Errorf("HTTP %d", resp.StatusCode)}
	}

	return &httpFile{
		name:   name,
		reader: resp.Body,
		size:   resp.ContentLength,
	}, nil
}

func (hfs *HTTPFileSystem) Stat(name string) (fs.FileInfo, error) {
	// For HTTP, we use HEAD request to check if file exists
	urlPath := filepath.ToSlash(name)
	fullURL := hfs.baseURL + "/" + strings.TrimPrefix(urlPath, "/")

	resp, err := hfs.client.Head(fullURL)
	if err != nil {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fmt.Errorf("HTTP %d", resp.StatusCode)}
	}

	return &httpFileInfo{
		name: filepath.Base(name),
		size: resp.ContentLength,
		mode: 0644,
	}, nil
}

func (hfs *HTTPFileSystem) Walk(root string, fn filepath.WalkFunc) error {
	// For HTTP, we can't really walk directories without a directory listing API
	// We'll try to open the root as a file directly
	info, err := hfs.Stat(root)
	if err != nil {
		if pathErr, ok := err.(*fs.PathError); ok && pathErr.Err == fs.ErrNotExist {
			// File doesn't exist, skip it
			return nil
		}
		return err
	}

	// Reconstruct the full URL for the walk function
	urlPath := filepath.ToSlash(root)
	fullURL := hfs.baseURL + "/" + strings.TrimPrefix(urlPath, "/")
	
	// Call the walk function with the full URL so it can be opened later
	return fn(fullURL, info, nil)
}

func (hfs *HTTPFileSystem) IsRemote() bool {
	return true
}

// httpFile implements fs.File for HTTP responses
type httpFile struct {
	name   string
	reader io.ReadCloser
	size   int64
}

func (f *httpFile) Stat() (fs.FileInfo, error) {
	return &httpFileInfo{
		name: filepath.Base(f.name),
		size: f.size,
		mode: 0644,
	}, nil
}

func (f *httpFile) Read(b []byte) (int, error) {
	return f.reader.Read(b)
}

func (f *httpFile) Close() error {
	return f.reader.Close()
}

// httpFileInfo implements fs.FileInfo for HTTP files
type httpFileInfo struct {
	name string
	size int64
	mode fs.FileMode
}

func (fi *httpFileInfo) Name() string       { return fi.name }
func (fi *httpFileInfo) Size() int64        { return fi.size }
func (fi *httpFileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *httpFileInfo) ModTime() time.Time { return time.Time{} }
func (fi *httpFileInfo) IsDir() bool        { return false }
func (fi *httpFileInfo) Sys() interface{}   { return nil }

// NewFileSystem creates a FileSystem based on the path
// If path is a URL (http:// or https://), returns HTTPFileSystem
// Otherwise returns LocalFileSystem
func NewFileSystem(pathOrURL string) (FileSystem, string, error) {
	// Check if it's a URL
	if strings.HasPrefix(pathOrURL, "http://") || strings.HasPrefix(pathOrURL, "https://") {
		u, err := url.Parse(pathOrURL)
		if err != nil {
			return nil, "", fmt.Errorf("invalid URL %s: %w", pathOrURL, err)
		}

		// Extract base URL and path
		baseURL := u.Scheme + "://" + u.Host
		filePath := strings.TrimPrefix(u.Path, "/")

		return NewHTTPFileSystem(baseURL), filePath, nil
	}

	// It's a local path
	return &LocalFileSystem{}, pathOrURL, nil
}

// resolveRulePath handles both local paths and URLs
// For local paths starting with ~, expands to home directory
// For URLs, returns them as-is
// For relative paths, joins with base directory
func resolveRulePath(rule, homeDir string) string {
	// If it's already a URL, return as-is
	if strings.HasPrefix(rule, "http://") || strings.HasPrefix(rule, "https://") {
		return rule
	}

	// Handle home directory expansion
	if strings.HasPrefix(rule, "~/") {
		return filepath.Join(homeDir, rule[2:])
	}

	// If it's an absolute path or starts with special directories, return as-is
	if filepath.IsAbs(rule) || strings.HasPrefix(rule, "/") {
		return rule
	}

	// For relative paths, return as-is (they'll be resolved relative to current dir)
	return rule
}

// openFile is a helper to open a file using the appropriate file system
func openFile(pathOrURL string) (fs.File, FileSystem, error) {
	fsys, resolvedPath, err := NewFileSystem(pathOrURL)
	if err != nil {
		return nil, nil, err
	}

	file, err := fsys.Open(resolvedPath)
	if err != nil {
		return nil, nil, err
	}

	return file, fsys, nil
}

// statFile is a helper to stat a file using the appropriate file system
func statFile(pathOrURL string) (fs.FileInfo, FileSystem, error) {
	fsys, resolvedPath, err := NewFileSystem(pathOrURL)
	if err != nil {
		return nil, nil, err
	}

	info, err := fsys.Stat(resolvedPath)
	if err != nil {
		return nil, nil, err
	}

	return info, fsys, nil
}

// walkPath is a helper to walk a path using the appropriate file system
func walkPath(pathOrURL string, fn filepath.WalkFunc) error {
	fsys, resolvedPath, err := NewFileSystem(pathOrURL)
	if err != nil {
		return err
	}

	return fsys.Walk(resolvedPath, fn)
}
