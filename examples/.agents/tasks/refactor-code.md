---
task_name: refactor-code
selectors:
  language: [go, python, javascript]
  stage: refactoring
---

# Refactor Code

You are an expert developer tasked with refactoring code to improve quality, maintainability, and performance.

## Refactoring Target

- **Component**: ${component_name}
- **Files**: ${files_to_refactor}
- **Goal**: ${refactoring_goal}

## Your Task

1. **Analyze Current Code**
   - Identify code smells
   - Find duplicated code
   - Locate overly complex functions
   - Review naming and structure

2. **Plan the Refactoring**
   - Define clear refactoring goals
   - Identify potential risks
   - Plan incremental changes
   - Consider backwards compatibility

3. **Implement Changes**
   - Make small, focused changes
   - Extract functions/methods for clarity
   - Improve naming and readability
   - Reduce complexity
   - Remove duplication

4. **Ensure Tests Pass**
   - Run existing tests after each change
   - Add tests if coverage is lacking
   - Verify no behavior changes
   - Use tests to guide refactoring

## Guidelines

This task automatically includes refactoring-stage rules for Go, Python, OR JavaScript (via array selector).

The selector `language: [go, python, javascript]` means rules matching ANY of these languages will be included, along with rules for stage=refactoring.

## Output

Provide:
- Analysis of current code issues
- Refactored code
- Explanation of improvements made
- Test results showing no regressions
