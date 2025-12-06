package codingcontext

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsedTask represents the result of parsing a task with the enhanced parser
type ParsedTask struct {
	// HasSlashCommand indicates if the task contains at least one slash command
	HasSlashCommand bool
	// FirstCommandName is the name of the first slash command found (if any)
	FirstCommandName string
	// FirstCommandParams are the parameters from the first slash command
	FirstCommandParams map[string]string
	// AllText is all the text content (including text blocks and commands) concatenated
	AllText string
	// Blocks contains the parsed blocks for more advanced use cases
	Blocks []*Block
}

// ParseTaskPrompt parses a task prompt using the enhanced parser and returns
// a ParsedTask that provides both the old interface (for backward compatibility)
// and the new block-based structure.
func ParseTaskPrompt(taskPrompt string) (*ParsedTask, error) {
	// First, try the enhanced parser
	input, err := ParseTask(taskPrompt)
	
	// If the enhanced parser fails, try to fall back to the old parseSlashCommand
	// for backward compatibility with slash commands that don't have a trailing newline
	if err != nil {
		// Try the old parser as a fallback
		slashTaskName, slashParams, found, parseErr := parseSlashCommand(taskPrompt)
		if parseErr != nil {
			// If both parsers fail, treat as plain text
			return &ParsedTask{
				HasSlashCommand: false,
				AllText:         taskPrompt,
				Blocks:          nil,
			}, nil
		}
		if found {
			// Old parser found a slash command - use it for backward compatibility
			return &ParsedTask{
				HasSlashCommand:    true,
				FirstCommandName:   slashTaskName,
				FirstCommandParams: slashParams,
				AllText:            taskPrompt,
				Blocks:             nil,
			}, nil
		}
		// Neither parser found a slash command - treat as plain text
		return &ParsedTask{
			HasSlashCommand: false,
			AllText:         taskPrompt,
			Blocks:          nil,
		}, nil
	}

	result := &ParsedTask{
		Blocks: input.Blocks,
	}

	// Find the first slash command and extract its parameters
	for _, block := range input.Blocks {
		if block.SlashCommand != nil {
			result.HasSlashCommand = true
			result.FirstCommandName = block.SlashCommand.Name
			result.FirstCommandParams = buildParametersMap(block.SlashCommand.Arguments)
			break
		}
	}

	// Build AllText by concatenating all blocks
	var textParts []string
	for _, block := range input.Blocks {
		if block.SlashCommand != nil {
			// Reconstruct the slash command as text
			textParts = append(textParts, reconstructSlashCommand(block.SlashCommand))
		} else if block.Text != nil {
			// Join text content
			textParts = append(textParts, strings.Join(block.Text.Content, ""))
		}
	}
	result.AllText = strings.Join(textParts, "\n")

	return result, nil
}

// buildParametersMap converts a list of arguments into a parameters map
// compatible with the old parseSlashCommand format
func buildParametersMap(args []*Argument) map[string]string {
	params := make(map[string]string)
	var argParts []string
	positionalIndex := 1

	for _, arg := range args {
		var argStr string
		if arg.Key != nil {
			// Named parameter: key=value or key="value"
			if arg.Value.String != nil {
				argStr = fmt.Sprintf("%s=\"%s\"", *arg.Key, *arg.Value.String)
				params[*arg.Key] = *arg.Value.String
			} else if arg.Value.Term != nil {
				argStr = fmt.Sprintf("%s=%s", *arg.Key, *arg.Value.Term)
				params[*arg.Key] = *arg.Value.Term
			}
			// Named parameters are also stored as positional in the old format
			params[strconv.Itoa(positionalIndex)] = argStr
			positionalIndex++
		} else {
			// Positional parameter
			if arg.Value.String != nil {
				argStr = fmt.Sprintf("\"%s\"", *arg.Value.String)
				params[strconv.Itoa(positionalIndex)] = *arg.Value.String
			} else if arg.Value.Term != nil {
				argStr = *arg.Value.Term
				params[strconv.Itoa(positionalIndex)] = *arg.Value.Term
			}
			positionalIndex++
		}
		if argStr != "" {
			argParts = append(argParts, argStr)
		}
	}

	if len(argParts) > 0 {
		params["ARGUMENTS"] = strings.Join(argParts, " ")
	}

	return params
}

// reconstructSlashCommand reconstructs a slash command from its parsed form
func reconstructSlashCommand(cmd *SlashCommand) string {
	result := "/" + cmd.Name
	if len(cmd.Arguments) == 0 {
		return result
	}

	var args []string
	for _, arg := range cmd.Arguments {
		var argStr string
		if arg.Key != nil {
			// Named parameter
			if arg.Value.String != nil {
				argStr = fmt.Sprintf("%s=\"%s\"", *arg.Key, *arg.Value.String)
			} else if arg.Value.Term != nil {
				argStr = fmt.Sprintf("%s=%s", *arg.Key, *arg.Value.Term)
			}
		} else {
			// Positional parameter
			if arg.Value.String != nil {
				argStr = fmt.Sprintf("\"%s\"", *arg.Value.String)
			} else if arg.Value.Term != nil {
				argStr = *arg.Value.Term
			}
		}
		if argStr != "" {
			args = append(args, argStr)
		}
	}
	return result + " " + strings.Join(args, " ")
}
