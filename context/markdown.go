package context

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	yaml "go.yaml.in/yaml/v2"
)

// ParseMarkdownFile parses the file into frontmatter and content
func ParseMarkdownFile(path string, frontmatter any) (string, error) {

	fh, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer fh.Close()

	s := bufio.NewScanner(fh)

	var content bytes.Buffer
	var frontMatterBytes bytes.Buffer

	// State machine: 0 = unknown, 1 = scanning frontmatter, 2 = scanning content
	state := 0

	for s.Scan() {
		line := s.Text()

		switch state {
		case 0: // State unknown - first line
			if line == "---" {
				state = 1 // Start scanning frontmatter
			} else {
				state = 2 // No frontmatter, start scanning content
				if _, err := content.WriteString(line + "\n"); err != nil {
					return "", fmt.Errorf("failed to write content: %w", err)
				}
			}
		case 1: // Scanning frontmatter
			if line == "---" {
				state = 2 // End of frontmatter, start scanning content
			} else {
				if _, err := frontMatterBytes.WriteString(line + "\n"); err != nil {
					return "", fmt.Errorf("failed to write frontmatter: %w", err)
				}
			}
		case 2: // Scanning content
			if _, err := content.WriteString(line + "\n"); err != nil {
				return "", fmt.Errorf("failed to write content: %w", err)
			}
		}
	}

	if err := s.Err(); err != nil {
		return "", fmt.Errorf("failed to scan file: %w", err)
	}

	// Parse frontmatter if we collected any
	if frontMatterBytes.Len() > 0 {
		if err := yaml.Unmarshal(frontMatterBytes.Bytes(), frontmatter); err != nil {
			return "", fmt.Errorf("failed to unmarshal frontmatter: %w", err)
		}
	}

	return content.String(), nil
}
