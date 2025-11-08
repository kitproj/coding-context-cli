---
task_name: review-code
task_type: review
---

# Code Review Task

You are tasked with conducting a thorough code review.

## Review Information

- **Pull Request**: #${pr_number}
- **Title**: ${pr_title}
- **Author**: ${pr_author}

## Your Task

1. **Code Quality Review**
   - Check code follows project conventions
   - Verify naming is clear and consistent
   - Ensure code is properly organized
   - Look for code duplication
   - Verify error handling is appropriate

2. **Functionality Review**
   - Verify implementation matches requirements
   - Check for edge cases
   - Review error handling
   - Validate input validation
   - Check for potential bugs

3. **Testing Review**
   - Verify test coverage is adequate
   - Check test quality and clarity
   - Ensure tests cover edge cases
   - Review test organization
   - Verify tests are maintainable

4. **Security Review**
   - Check for security vulnerabilities
   - Verify input sanitization
   - Review authentication/authorization
   - Check for injection vulnerabilities
   - Verify secrets are not hardcoded

5. **Documentation Review**
   - Check for necessary comments
   - Verify API documentation is updated
   - Review commit messages
   - Ensure breaking changes are documented

## Guidelines for Code Review

- **DO NOT** implement changes yourself
- **DO** provide specific, actionable feedback
- **DO** explain the reasoning behind suggestions
- **DO** distinguish between blocking issues and suggestions
- **DO** be constructive and helpful
- **DO** acknowledge good practices

## Output

Provide:
- List of issues found (categorized by severity)
- Specific suggestions for improvement
- Praise for good practices
- Overall recommendation (approve, request changes, comment)
