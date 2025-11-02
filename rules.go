package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetRuleContent is a function that returns the content of a rule as a string
type GetRuleContent func() string

// FindUserRules returns a list of GetRuleContent functions for user-level rules
// These are rules stored in the user's home directory (~/.prompts/rules)
func FindUserRules() ([]GetRuleContent, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	userRulesPath := filepath.Join(homeDir, ".prompts", "rules")
	
	// Check if the user rules directory exists
	if _, err := os.Stat(userRulesPath); os.IsNotExist(err) {
		return []GetRuleContent{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat user rules path %s: %w", userRulesPath, err)
	}

	var ruleFuncs []GetRuleContent

	err = filepath.Walk(userRulesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Only process .md files as rule files
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Create a closure that captures the path
		rulePath := path
		ruleFunc := func() string {
			content, err := os.ReadFile(rulePath)
			if err != nil {
				return ""
			}
			return string(content)
		}
		ruleFuncs = append(ruleFuncs, ruleFunc)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk user rules directory: %w", err)
	}

	return ruleFuncs, nil
}

// GetNormalizedRulePaths returns the normalized 3-level hierarchy of rule paths
// L0: System-rules (/etc/prompts/rules)
// L1: User-rules (~/.prompts/rules) 
// L2: Project-rules (.prompts/rules)
func GetNormalizedRulePaths() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	return []string{
		".prompts/rules",                             // L2: Project-rules
		filepath.Join(homeDir, ".prompts", "rules"),  // L1: User-rules
		"/etc/prompts/rules",                         // L0: System-rules
	}, nil
}

// GetNormalizedPersonaPaths returns the normalized 3-level hierarchy of persona paths
func GetNormalizedPersonaPaths() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	return []string{
		".prompts/personas",
		filepath.Join(homeDir, ".prompts", "personas"),
		"/etc/prompts/personas",
	}, nil
}

// GetNormalizedTaskPaths returns the normalized 3-level hierarchy of task paths
func GetNormalizedTaskPaths() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	return []string{
		".prompts/tasks",
		filepath.Join(homeDir, ".prompts", "tasks"),
		"/etc/prompts/tasks",
	}, nil
}
