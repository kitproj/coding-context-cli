---
task_type: review
---

# Code Review Guidelines

These guidelines apply when reviewing code.

## Review Mindset

- Be constructive and helpful
- Assume good intent
- Focus on the code, not the person
- Provide specific, actionable feedback
- Explain your reasoning
- Acknowledge good practices

## What to Review

### Code Quality
- Does code follow project conventions?
- Is naming clear and consistent?
- Is code properly organized?
- Is there unnecessary duplication?
- Is complexity justified?

### Functionality
- Does implementation match requirements?
- Are edge cases handled?
- Is error handling appropriate?
- Are inputs validated?
- Could this introduce bugs?

### Testing
- Is test coverage adequate?
- Do tests cover edge cases?
- Are tests clear and maintainable?
- Do tests verify the right things?

### Security
- Are there security vulnerabilities?
- Is user input sanitized?
- Are secrets handled properly?
- Is authentication/authorization correct?
- Are there injection risks?

### Documentation
- Are complex parts commented?
- Is API documentation updated?
- Are breaking changes documented?
- Are commit messages clear?

## Feedback Categories

Categorize feedback by severity:

- **🔴 Blocking**: Must be fixed before merge
  - Security vulnerabilities
  - Bugs or incorrect behavior
  - Breaking changes without migration path

- **🟡 Important**: Should be addressed
  - Missing tests
  - Poor error handling
  - Performance concerns
  - Maintainability issues

- **🟢 Suggestion**: Nice to have
  - Code style improvements
  - Better naming
  - Additional tests
  - Refactoring opportunities

## Providing Feedback

- Be specific about what to change
- Explain why the change is needed
- Suggest alternatives when possible
- Link to relevant documentation
- Provide code examples if helpful

## Example Feedback

Good:
> "This function could throw a NullPointerException if `user` is null on line 42. Consider adding a null check or using Optional<User>."

Bad:
> "This code is wrong."

## Final Recommendation

Choose one:
- ✅ **Approve**: Ready to merge
- 🔄 **Request Changes**: Issues must be addressed
- 💬 **Comment**: Feedback provided, author decides
