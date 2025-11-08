---
language: Go
task_type: bug-fix
---

# Go Bug Fix Standards

## Testing Requirements for Bug Fixes

- **Write regression tests** - Every bug fix MUST include a test that would have caught the bug
- **Test the fix** - Verify the test fails without your fix and passes with it
- **Test edge cases** - Consider boundary conditions that might trigger similar bugs
- **Use table-driven tests** - Follow Go conventions for testing multiple scenarios

## Code Changes

- **Minimal changes only** - Don't refactor unrelated code
- **Preserve behavior** - Ensure backwards compatibility unless explicitly breaking
- **Document the fix** - Add comments explaining why the bug occurred and how it's fixed

## Example Test Structure

```go
func TestBugFix_IssueXXX(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "regression case from issue XXX",
            input: "problematic input",
            want:  "expected output",
        },
        // Additional test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := functionUnderTest(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```
