---
language: Go
task_type: poc
---

# Go POC (Proof of Concept) Guidelines

## Speed Over Quality

For proof of concepts, prioritize getting a working prototype quickly:

- **Skip comprehensive tests** - Basic manual testing is sufficient
- **Skip detailed comments** - Write self-explanatory code, but don't over-document
- **Use shortcuts** - Hard-coded values, mock data, simplified error handling are all acceptable
- **Copy-paste is fine** - Don't worry about DRY principles for a POC

## What to Focus On

- **Demonstrate the core concept** - Show that the idea works
- **Identify blockers early** - Find technical limitations or challenges
- **Validate assumptions** - Test your hypotheses about the approach
- **Document learnings** - Capture what you discovered, not how the code works

## Code Quality Bar

- Code should run without crashes
- Basic error handling for the happy path
- Readable enough for others to understand the concept
- No need for production-ready error messages or logging

## What NOT to Do

- Don't spend time on comprehensive test coverage
- Don't worry about edge cases unless they're critical to the POC
- Don't refactor for code quality - it's throwaway code
- Don't add detailed documentation or comments

## When POC is Done

Document:
- What worked
- What didn't work
- Technical risks discovered
- Estimated effort for production implementation
- Recommendations for next steps
