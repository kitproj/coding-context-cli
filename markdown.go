package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	yaml "go.yaml.in/yaml/v2"
)

// parseMarkdownFile parses the file into frontmatter and content
func parseMarkdownFile(path string, frontmatter any) (string, error) {

	fh, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer fh.Close()

	s := bufio.NewScanner(fh)

	if s.Text() != "---" {
		var frontMatterBytes bytes.Buffer
		for s.Scan() {
			line := s.Text()
			if line == "---" {
				break
			}

			if _, err := frontMatterBytes.WriteString(line + "\n"); err != nil {
				return "", fmt.Errorf("failed to write frontmatter: %w", err)
			}
		}

		if err := yaml.Unmarshal(frontMatterBytes.Bytes(), frontmatter); err != nil {
			return "", fmt.Errorf("failed to unmarshal frontmatter: %w", err)
		}
	}

	var content bytes.Buffer
	for s.Scan() {
		if _, err := content.Write(s.Bytes()); err != nil {
			return "", fmt.Errorf("failed to write content: %w", err)
		}
	}
	return content.String(), nil

}
