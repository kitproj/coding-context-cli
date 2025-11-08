---
task_type: production
stage: implementation
---

# Production Code Standards

These standards apply when writing production code.

## Code Quality

- All code must follow project coding standards
- Use meaningful variable and function names
- Keep functions focused and small (single responsibility)
- Add comments for complex logic
- Ensure code is self-documenting

## Error Handling

- Handle all error cases explicitly
- Provide meaningful error messages
- Use appropriate error types
- Log errors with sufficient context
- Don't silently ignore errors

## Testing Requirements

- Write unit tests for all new functions
- Aim for >80% code coverage
- Test edge cases and error conditions
- Use table-driven tests where appropriate
- Add integration tests for user-facing features
- Ensure tests are deterministic and isolated

## Security

- Validate all user inputs
- Never hardcode secrets or credentials
- Use parameterized queries to prevent injection
- Follow principle of least privilege
- Sanitize outputs to prevent XSS

## Performance

- Consider performance implications
- Profile before optimizing
- Document performance requirements
- Use appropriate data structures
- Avoid premature optimization

## Documentation

- Document all public APIs
- Add inline comments for complex logic
- Update user-facing documentation
- Document breaking changes
- Include usage examples
