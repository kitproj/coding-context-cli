---
task_name: implement-feature
selectors:
  language: go
  stage: implementation
---

# Implement Feature in Go

You are an expert Go developer implementing a new feature.

## Feature Information

- **Name**: ${feature_name}
- **Description**: ${feature_description}

## Your Task

1. **Design the Implementation**
   - Follow Go idioms and best practices
   - Use existing patterns in the codebase
   - Keep functions small and focused
   - Use interfaces for flexibility

2. **Write the Code**
   - Implement the feature with clean, readable Go code
   - Add proper error handling
   - Include necessary comments
   - Follow the project's code style

3. **Add Tests**
   - Write table-driven tests (project standard)
   - Test happy paths and error cases
   - Aim for >80% code coverage
   - Use meaningful test names

4. **Update Documentation**
   - Add godoc comments for exported functions
   - Update README if needed
   - Document any new behavior

## Guidelines

This task automatically includes:
- Go-specific coding rules (via language=Go selector)
- Implementation-stage guidelines (via stage=implementation selector)

Follow all included rules and best practices.

## Output

Provide:
- The implementation code
- Comprehensive tests
- Documentation updates
