---
stage: testing
---

# Testing Standards

## Test Coverage

- Write tests for all public APIs
- Test edge cases and error conditions
- Aim for >80% code coverage
- Write integration tests for critical paths

## Test Organization

- Group related tests in test files
- Use descriptive test names
- Follow AAA pattern: Arrange, Act, Assert
- Keep tests independent and isolated

## Test Data

- Use fixtures for complex test data
- Avoid hardcoding test data in tests
- Clean up test data after tests
- Use realistic test scenarios

## Mocking and Stubbing

- Mock external dependencies
- Use interfaces for testability
- Avoid over-mocking
- Test real integrations when possible

## Performance Testing

- Write benchmarks for performance-critical code
- Set performance baselines
- Monitor performance over time
- Load test under realistic conditions

## Test Automation

- All tests must run in CI/CD
- Tests should be fast (<5 minutes total)
- Flaky tests should be fixed or removed
- Failed tests should block deployment
