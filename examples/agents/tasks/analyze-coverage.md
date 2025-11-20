---
task_name: analyze-coverage
description: Analyze test coverage
---

# Test Coverage Analysis

Here are the current test results:
!`go test -cover ./... 2>&1 | grep -E "(^ok|coverage:)" | head -10`

Based on these results, suggest improvements to increase coverage.

## Guidelines

- Focus on areas with low coverage
- Identify untested edge cases
- Suggest additional test scenarios
- Prioritize critical code paths
