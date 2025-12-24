---
name: code-review
description: Perform systematic code reviews focusing on security, performance, maintainability, and best practices. Use when reviewing pull requests or code changes.
license: Apache-2.0
metadata:
  author: example-team
  version: "1.0"
---

# Code Review Skill

## When to use this skill

Use this skill when:
- Reviewing pull requests
- Conducting code audits
- Providing feedback on code changes
- Checking code quality before merging

## Review Checklist

### Security
- [ ] No hardcoded credentials or secrets
- [ ] Input validation for user-supplied data
- [ ] Proper authentication and authorization checks
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS prevention (output encoding)
- [ ] CSRF protection where applicable

### Performance
- [ ] No N+1 query problems
- [ ] Appropriate use of caching
- [ ] Efficient algorithms and data structures
- [ ] No unnecessary database calls
- [ ] Proper indexing for queries

### Maintainability
- [ ] Code follows project style guide
- [ ] Functions are small and focused
- [ ] Clear and descriptive naming
- [ ] Appropriate comments for complex logic
- [ ] No code duplication (DRY principle)
- [ ] Proper error handling

### Testing
- [ ] Unit tests for new functionality
- [ ] Edge cases covered
- [ ] No tests disabled or skipped
- [ ] Test names describe what they test

### Documentation
- [ ] API changes documented
- [ ] README updated if needed
- [ ] Breaking changes clearly noted
- [ ] Migration guide for breaking changes

## Review Process

1. **Read the description**: Understand what the change is trying to accomplish
2. **Check the tests**: Tests should clarify expected behavior
3. **Review the implementation**: Look for issues in the checklist
4. **Verify edge cases**: Think about what could go wrong
5. **Suggest improvements**: Focus on actionable, specific feedback

## Providing Feedback

Good feedback is:
- **Specific**: Point to exact lines or sections
- **Actionable**: Suggest concrete improvements
- **Educational**: Explain why something should change
- **Respectful**: Focus on the code, not the person

Example:
```
In line 42, the error is silently ignored. Consider logging it or 
returning it to the caller so failures are visible.
```

## Common Issues to Watch For

### Go
- Unchecked errors
- Goroutine leaks
- Race conditions
- Not closing resources (defer file.Close())

### Python
- Mutable default arguments
- Not using context managers for resources
- Security issues with eval() or exec()
- Not handling exceptions properly

### JavaScript/TypeScript
- Promise rejections not handled
- Memory leaks (event listeners not removed)
- Not using async/await properly
- Type safety issues (TypeScript)

## Approval Guidelines

Approve when:
- All checklist items pass
- Tests are comprehensive
- No security concerns
- Code is maintainable

Request changes when:
- Security issues present
- Tests insufficient
- Major performance concerns
- Breaking changes not documented
