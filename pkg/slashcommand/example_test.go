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

	// Parse a command with single argument
	taskName, params, err = slashcommand.ParseSlashCommand("/fix-bug 123")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Task: %s, $1: %s\n", taskName, params["1"])

	// Parse a command with multiple arguments
	taskName, params, err = slashcommand.ParseSlashCommand(`/implement-feature "User Login" high`)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Task: %s, $1: %s, $2: %s\n", taskName, params["1"], params["2"])

	// Output:
	// Task: fix-bug, Params: map[]
	// Task: fix-bug, $1: 123
	// Task: implement-feature, $1: User Login, $2: high
}
