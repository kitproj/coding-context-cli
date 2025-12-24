---
task_name: write-tests
selectors:
  stage: testing
---

# Write Comprehensive Tests

You are an expert developer tasked with writing comprehensive tests.

## Test Requirements

- **Component**: ${component_name}
- **File(s)**: ${files_to_test}

## Your Task

1. **Analyze the Code**
   - Understand the functionality being tested
   - Identify all code paths
   - Find edge cases and boundary conditions
   - Review existing tests for patterns

2. **Write Test Cases**
   - Cover all happy paths
   - Test error conditions
   - Test edge cases and boundary values
   - Test concurrent operations if applicable

3. **Follow Language Conventions**
   - Use meaningful test names that describe what's being tested
   - Keep tests isolated and independent
   - Follow the testing patterns in the codebase

4. **Verify Coverage**
   - Run tests to ensure they pass
   - Check code coverage metrics
   - Aim for >80% coverage
   - Ensure all critical paths are tested

## Guidelines

This task automatically includes testing-stage guidelines (via stage=testing selector).

Use `-s language=go` or similar if you need language-specific rules.

## Output

Provide:
- Well-structured test code
- Test coverage report
- Any testing utilities or fixtures needed
