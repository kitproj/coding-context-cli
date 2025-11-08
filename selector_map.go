package main

import (
	"fmt"

	"github.com/google/cel-go/cel"
)

// selector holds a CEL expression for filtering frontmatter
type selector struct {
	expression string
	program    cel.Program
}

func (s *selector) String() string {
	return s.expression
}

func (s *selector) Set(value string) error {
	// If empty, allow it (means no filtering)
	if value == "" {
		s.expression = ""
		s.program = nil
		return nil
	}

	// Create a CEL environment with dynamic types for frontmatter
	env, err := cel.NewEnv(
		cel.Variable("frontmatter", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return fmt.Errorf("failed to create CEL environment: %w", err)
	}

	// Parse the expression
	ast, issues := env.Compile(value)
	if issues != nil && issues.Err() != nil {
		return fmt.Errorf("failed to compile CEL expression: %w", issues.Err())
	}

	// Check that the expression returns a boolean
	if !ast.OutputType().IsExactType(cel.BoolType) {
		return fmt.Errorf("CEL expression must return a boolean value")
	}

	// Create a program
	prg, err := env.Program(ast)
	if err != nil {
		return fmt.Errorf("failed to create CEL program: %w", err)
	}

	s.expression = value
	s.program = prg
	return nil
}

// matchesIncludes returns true if the frontmatter matches the CEL selector expression
func (s *selector) matchesIncludes(frontmatter frontMatter) bool {
	// If no expression is set, match everything
	if s.program == nil {
		return true
	}

	// Convert frontmatter to map[string]any for CEL evaluation
	fmMap := make(map[string]any)
	for k, v := range frontmatter {
		fmMap[k] = v
	}

	// Evaluate the expression
	result, _, err := s.program.Eval(map[string]any{
		"frontmatter": fmMap,
	})
	if err != nil {
		// If evaluation fails (e.g., due to missing field), match it
		// This maintains backward compatibility where missing keys were allowed
		return true
	}

	// Check if result is true
	boolResult, ok := result.Value().(bool)
	if !ok {
		return false
	}

	return boolResult
}
