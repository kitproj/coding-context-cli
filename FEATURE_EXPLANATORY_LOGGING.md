# Feature: Explanatory Logging

## Overview

The CLI now clearly explains WHY each rule, task, skill, and command was included or excluded when running a task. This makes it much easier to understand the context assembly process and debug selector issues.

## What Changed

### New Functionality

1. **Inclusion Explanations**: Every included item now logs the reason it was selected
2. **Exclusion Explanations**: Every excluded item now logs why it didn't match selectors
3. **Detailed Selector Matching**: Shows exactly which selector key-value pairs matched or didn't match

### Modified Files

- `pkg/codingcontext/selectors/selectors.go`: Refactored `MatchesIncludes()` to return `(bool, string)` with reason
- `pkg/codingcontext/context.go`: Enhanced logging throughout to explain inclusion/exclusion
- `pkg/codingcontext/selectors/selectors_test.go`: Added comprehensive table-driven tests

## Usage Examples

### Example 1: Including Rules with Matching Selectors

```bash
./coding-context -C examples -s language=go -s env=dev my-task
```

**Output:**
```
time=... level=INFO msg="Including task" name=my-task reason="task name matches 'my-task'" tokens=39
time=... level=INFO msg="Including rule file" path=.agents/rules/go-dev-rule.md reason="matched selectors: language=go, env=dev" tokens=44
time=... level=INFO msg="Including rule file" path=.agents/rules/generic-rule.md reason="no selectors specified (included by default)" tokens=45
time=... level=INFO msg="Discovered skill" name=go-testing reason="matched selectors: language=go" path=/path/to/skill/SKILL.md
```

### Example 2: Skipping Non-Matching Items

```bash
./coding-context -C examples -s language=python -s env=prod my-task
```

**Output:**
```
time=... level=INFO msg="Including task" name=my-task reason="task name matches 'my-task'" tokens=39
time=... level=INFO msg="Skipping file" path=.agents/rules/go-dev-rule.md reason="selectors did not match: language=go (expected language=python), env=dev (expected env=prod)"
time=... level=INFO msg="Including rule file" path=.agents/rules/python-prod-rule.md reason="matched selectors: language=python, env=prod" tokens=41
time=... level=INFO msg="Skipping skill" name=go-testing path=/path/to/skill/SKILL.md reason="selectors did not match: language=go (expected language=python)"
```

## Log Message Format

### Inclusion Messages

- **Tasks**: `Including task` with `name` and `reason="task name matches '<task-name>'"`
- **Rules**: `Including rule file` with `path`, `reason`, and `tokens`
  - With selectors: `reason="matched selectors: key1=value1, key2=value2"`
  - Without selectors: `reason="no selectors specified (included by default)"`
- **Skills**: `Discovered skill` with `name`, `path`, and `reason`
  - With selectors: `reason="matched selectors: key1=value1, key2=value2"`
  - Without selectors: `reason="no selectors specified (included by default)"`
- **Commands**: `Including command` with `name`, `path`, and `reason="referenced by slash command '/command-name'"`

### Exclusion Messages

- **Rules**: `Skipping file` with `path` and detailed mismatch explanation
  - Single mismatch: `reason="selectors did not match: key=actual_value (expected key=expected_value)"`
  - Multiple mismatches: `reason="selectors did not match: key1=actual1 (expected key1=expected1), key2=actual2 (expected key2=expected2)"`
  - OR logic (multiple allowed values): `reason="selectors did not match: key=actual (expected key in [value1, value2, value3])"`

- **Skills**: `Skipping skill` with `name`, `path`, and detailed mismatch explanation (same format as rules)

## Implementation Details

### Refactored Method in `Selectors`

#### `MatchesIncludes(frontmatter BaseFrontMatter) (bool, string)`

Returns whether the frontmatter matches all include selectors, along with a human-readable reason explaining the result.

**Returns:**
- `bool`: true if all selectors match, false otherwise
- `string`: reason explaining why (matched selectors or mismatch details)

**Examples:**
```go
// Match case
match, reason := selectors.MatchesIncludes(frontmatter)
// match = true, reason = "matched selectors: language=go, env=dev"

// No match case
match, reason := selectors.MatchesIncludes(frontmatter)
// match = false, reason = "selectors did not match: language=python (expected language=go)"

// No selectors case
match, reason := selectors.MatchesIncludes(frontmatter)
// match = true, reason = "no selectors specified (included by default)"
```

### Selector Matching Logic

- **Missing keys**: If a selector key doesn't exist in frontmatter, it's allowed (not counted as a mismatch)
- **Matching values**: If a frontmatter value matches any selector value for that key, it matches (OR logic)
- **Non-matching values**: If a frontmatter value doesn't match any selector value for that key, it doesn't match

## Testing

### Test Coverage

- `TestSelectorMap_MatchesIncludes`: 19 test cases covering basic matching/non-matching scenarios
- `TestSelectorMap_MatchesIncludesReasons`: 9 test cases covering reason explanations with various scenarios
- All tests use table-driven test pattern (project standard)
- Tests cover: single selectors, multiple selectors, array selectors, OR logic, edge cases, both match and no-match cases

### Running Tests

```bash
# Test the selectors package
go test -v ./pkg/codingcontext/selectors/

# Test the main context package
go test -v ./pkg/codingcontext/

# Run all tests
go test -v ./...
```

## Benefits

1. **Debugging**: Instantly see why rules/skills aren't being included
2. **Transparency**: Understand exactly how selector matching works
3. **Configuration Validation**: Verify that your frontmatter and selectors are set up correctly
4. **Learning**: New users can understand the system by observing the logs
5. **Efficiency**: Single method call returns both match result and reason (no duplicate work)

## Backwards Compatibility

This feature is fully backwards compatible:
- No breaking changes to APIs or behavior
- Only adds additional logging information
- All existing tests pass
- Existing rule/task/skill files work without modification

## Future Enhancements

Potential improvements for future versions:
- Add a `--quiet` flag to suppress inclusion/exclusion logging
- Add a `--explain` flag to show even more detailed selector evaluation
- Color-code inclusion (green) vs exclusion (yellow) messages in terminal output
- Add summary statistics (e.g., "Included 5 rules, skipped 3 rules")