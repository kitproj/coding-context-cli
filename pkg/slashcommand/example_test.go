package slashcommand_test

import (
"fmt"
"github.com/kitproj/coding-context-cli/pkg/slashcommand"
)

func ExampleParseSlashCommand() {
// Parse a simple command without parameters
taskName, params, err := slashcommand.ParseSlashCommand("/fix-bug")
if err != nil {
fmt.Printf("Error: %v\n", err)
return
}
fmt.Printf("Task: %s, Params: %v\n", taskName, params)

// Parse a command with single parameter
taskName, params, err = slashcommand.ParseSlashCommand(`/fix-bug issue_number="123"`)
if err != nil {
fmt.Printf("Error: %v\n", err)
return
}
fmt.Printf("Task: %s, Issue: %s\n", taskName, params["issue_number"])

// Parse a command with multiple parameters
taskName, params, err = slashcommand.ParseSlashCommand(`/implement-feature feature_name="User Login" priority="high"`)
if err != nil {
fmt.Printf("Error: %v\n", err)
return
}
fmt.Printf("Task: %s, Feature: %s, Priority: %s\n", taskName, params["feature_name"], params["priority"])

// Output:
// Task: fix-bug, Params: map[]
// Task: fix-bug, Issue: 123
// Task: implement-feature, Feature: User Login, Priority: high
}
