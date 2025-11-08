---
task_type: production
stage: implementation
language: go
---

# Go Implementation Standards

## Code Style

- Follow Go standard formatting (use `gofmt`)
- Use meaningful variable and function names
- Keep functions focused and small
- Comment exported functions and types
- Use Go idioms and patterns

## Error Handling

- Always check and handle errors explicitly
- Wrap errors with context using `fmt.Errorf`
- Return errors rather than panicking
- Use custom error types when appropriate

## Testing

- Write unit tests for all functions
- Use table-driven tests for multiple scenarios
- Aim for high test coverage (>80%)
- Use meaningful test names that describe the scenario

## Concurrency

- Use channels for communication between goroutines
- Use sync package primitives appropriately
- Avoid shared state when possible
- Document goroutine lifecycle

## Dependencies

- Minimize external dependencies
- Use standard library when possible
- Vendor dependencies for reproducible builds
- Keep dependencies up to date

## Performance

- Profile before optimizing
- Avoid premature optimization
- Use benchmarks to measure performance
- Consider memory allocations
