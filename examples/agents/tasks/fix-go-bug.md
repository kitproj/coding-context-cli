---
task_name: fix-go-bug
selector:
  language: Go
  task_type: bug-fix
---

# Fix Go Bug

You are tasked with fixing a bug in the Go codebase.

## Issue Information

- **Issue Number**: #${issue_number}
- **Title**: ${issue_title}
- **Description**: ${issue_description}

## Your Task

1. **Analyze the Issue**
   - Understand the reported problem
   - Identify the root cause in the Go code
   - Determine affected components

2. **Implement the Fix**
   - Write minimal code changes to fix the bug
   - Follow Go coding standards
   - Ensure backwards compatibility

3. **Add Tests**
   - Write regression tests using Go testing framework
   - Verify the fix works as expected
   - Test edge cases

4. **Document Changes**
   - Update relevant documentation
   - Add code comments if needed

## Guidelines

- Make minimal changes - only fix the specific issue
- Follow the Go testing standards for bug fixes
- Ensure all existing tests still pass
