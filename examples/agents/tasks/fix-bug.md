---
task_name: fix-bug
---

# Bug Fix Task

You are an expert software developer tasked with fixing a bug.

## Issue Information

- **Issue Number**: #${issue_number}
- **Title**: ${issue_title}
- **URL**: ${issue_url}
- **Description**:
${issue_body}

## Your Task

1. **Analyze the Issue**
   - Understand the reported problem
   - Identify the root cause
   - Determine affected components

2. **Implement the Fix**
   - Write minimal code changes to fix the bug
   - Follow existing code patterns
   - Ensure backwards compatibility
   - Handle edge cases

3. **Add Tests**
   - Write regression tests
   - Verify the fix works
   - Test edge cases
   - Ensure no new bugs introduced

4. **Document Changes**
   - Update relevant documentation
   - Add code comments if needed
   - Document any behavior changes

## Guidelines

- Make minimal changes - only fix the specific issue
- Don't refactor unrelated code
- Ensure all existing tests still pass
- Add clear commit messages
- Consider if this bug might exist elsewhere

## Output

Provide:
- Root cause analysis
- The fix implementation
- Test cases
- Any documentation updates
