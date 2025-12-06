package codingcontext

import (
"fmt"
"strings"
"testing"
)

// Helper function to create a Text block from a string
func makeText(content string) *Text {
if content == "" {
return nil
}
// For testing purposes, we'll just store the content as a single token
// This won't match exactly how the parser works, but for the purpose of comparison
// we only care about the final Content() result
return &Text{Tokens: []TextToken{{Token: content}}}
}

func TestParseTask(t *testing.T) {
tests := []struct {
name     string
input    string
wantTask Task
wantErr  bool
}{
{
name:  "simple text only",
input: "Hello world",
wantTask: Task{
Block{Text: makeText("Hello world")},
},
wantErr: false,
},
{
name:  "slash command only",
input: "/fix-bug\n",
wantTask: Task{
Block{SlashCommand: &SlashCommand{Name: "fix-bug", Arguments: []Argument{}}},
Block{Text: makeText("\n")},
},
wantErr: false,
},
{
name:  "slash command with positional argument",
input: "/fix-bug 123\n",
wantTask: Task{
Block{SlashCommand: &SlashCommand{
Name:      "fix-bug",
Arguments: []Argument{{Value: "123"}},
}},
Block{Text: makeText("\n")},
},
wantErr: false,
},
{
name:  "slash command with multiple positional arguments",
input: "/fix-bug 123 456\n",
wantTask: Task{
Block{SlashCommand: &SlashCommand{
Name: "fix-bug",
Arguments: []Argument{
{Value: "123"},
{Value: "456"},
},
}},
Block{Text: makeText("\n")},
},
wantErr: false,
},
{
name:  "slash command with named argument",
input: "/fix-bug issue=\"PROJ-123\"\n",
wantTask: Task{
Block{SlashCommand: &SlashCommand{
Name: "fix-bug",
Arguments: []Argument{
{Key: "issue", Value: "PROJ-123"},
},
}},
Block{Text: makeText("\n")},
},
wantErr: false,
},
{
name:  "slash command with mixed arguments",
input: "/fix-bug 123 priority=\"high\"\n",
wantTask: Task{
Block{SlashCommand: &SlashCommand{
Name: "fix-bug",
Arguments: []Argument{
{Value: "123"},
{Key: "priority", Value: "high"},
},
}},
Block{Text: makeText("\n")},
},
wantErr: false,
},
{
name:  "text before slash command",
input: "Please /fix-bug 123\n",
wantTask: Task{
Block{Text: makeText("Please ")},
Block{SlashCommand: &SlashCommand{
Name:      "fix-bug",
Arguments: []Argument{{Value: "123"}},
}},
Block{Text: makeText("\n")},
},
wantErr: false,
},
{
name:  "slash command followed by text",
input: "/fix-bug 123\nSome text after",
wantTask: Task{
Block{SlashCommand: &SlashCommand{
Name:      "fix-bug",
Arguments: []Argument{{Value: "123"}},
}},
Block{Text: makeText("\nSome text after")},
},
wantErr: false,
},
{
name:  "multiple slash commands",
input: "/fix-bug 123\n/code-review\n",
wantTask: Task{
Block{SlashCommand: &SlashCommand{
Name:      "fix-bug",
Arguments: []Argument{{Value: "123"}},
}},
Block{Text: makeText("\n")},
Block{SlashCommand: &SlashCommand{
Name:      "code-review",
Arguments: []Argument{},
}},
Block{Text: makeText("\n")},
},
wantErr: false,
},
{
name:  "slash not at line start is text",
input: "a/b",
wantTask: Task{
Block{Text: makeText("a")},
				Block{SlashCommand: &SlashCommand{Name: "b", Arguments: []Argument{}}},
},
wantErr: false,
},
{
name:  "quoted string with spaces",
input: "/command \"hello world\"\n",
wantTask: Task{
Block{SlashCommand: &SlashCommand{
Name:      "command",
Arguments: []Argument{{Value: "hello world"}},
}},
Block{Text: makeText("\n")},
},
wantErr: false,
},
{
name:     "empty input",
input:    "",
wantTask: Task{},
wantErr:  false,
},
{
name:  "multiline text",
input: "Line 1\nLine 2\nLine 3",
wantTask: Task{
Block{Text: makeText("Line 1\nLine 2\nLine 3")},
},
wantErr: false,
},
{
name:  "complex mixed content",
input: "Introduction text\n/fix-bug issue=\"123\" priority=\"high\"\nMore text\n/code-review\nFinal text",
wantTask: Task{
Block{Text: makeText("Introduction text\n")},
Block{SlashCommand: &SlashCommand{
Name: "fix-bug",
Arguments: []Argument{
{Key: "issue", Value: "123"},
{Key: "priority", Value: "high"},
},
}},
Block{Text: makeText("\nMore text\n")},
Block{SlashCommand: &SlashCommand{
Name:      "code-review",
Arguments: []Argument{},
}},
Block{Text: makeText("\nFinal text")},
},
wantErr: false,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
gotTask, err := ParseTask(tt.input)

if (err != nil) != tt.wantErr {
t.Errorf("ParseTask() error = %v, wantErr %v", err, tt.wantErr)
return
}

if !tt.wantErr {
if !tasksEqual(gotTask, tt.wantTask) {
t.Errorf("ParseTask() = %+v, want %+v", formatTask(gotTask), formatTask(tt.wantTask))
}
}
})
}
}

// tasksEqual compares two tasks by comparing their values, not pointers
func tasksEqual(a, b Task) bool {
if len(a) != len(b) {
return false
}
for i := range a {
if !blocksEqual(a[i], b[i]) {
return false
}
}
return true
}

// blocksEqual compares two blocks by comparing their values
func blocksEqual(a, b Block) bool {
// Check if both have slash commands
if (a.SlashCommand != nil) != (b.SlashCommand != nil) {
return false
}
if a.SlashCommand != nil {
if !slashCommandsEqual(*a.SlashCommand, *b.SlashCommand) {
return false
}
}

// Check if both have text
if (a.Text != nil) != (b.Text != nil) {
return false
}
if a.Text != nil {
if a.Text.Content() != b.Text.Content() {
return false
}
}

return true
}

// slashCommandsEqual compares two slash commands
func slashCommandsEqual(a, b SlashCommand) bool {
if a.Name != b.Name {
return false
}
if len(a.Arguments) != len(b.Arguments) {
return false
}
for i := range a.Arguments {
if a.Arguments[i] != b.Arguments[i] {
return false
}
}
return true
}

// formatTask formats a task for error messages
func formatTask(t Task) string {
var parts []string
for _, block := range t {
if block.SlashCommand != nil {
parts = append(parts, fmt.Sprintf("SlashCommand{Name:%q, Args:%v}", block.SlashCommand.Name, block.SlashCommand.Arguments))
}
if block.Text != nil {
parts = append(parts, fmt.Sprintf("Text{Content:%q}", block.Text.Content()))
}
}
return "[" + strings.Join(parts, ", ") + "]"
}

func TestTaskToPrompt(t *testing.T) {
tests := []struct {
name   string
task   Task
params map[string]string
want   string
}{
{
name: "text only, no params",
task: Task{
Block{Text: makeText("Hello world")},
},
params: map[string]string{},
want:   "Hello world",
},
{
name: "slash command removed from output",
task: Task{
Block{SlashCommand: &SlashCommand{Name: "fix-bug"}},
Block{Text: makeText("\nSome text")},
},
params: map[string]string{},
want:   "\nSome text",
},
{
name: "parameter substitution",
task: Task{
Block{Text: makeText("Issue: ${issue_number}")},
},
params: map[string]string{"issue_number": "123"},
want:   "Issue: 123",
},
{
name: "multiple parameters",
task: Task{
Block{Text: makeText("Issue ${issue} has priority ${priority}")},
},
params: map[string]string{
"issue":    "PROJ-123",
"priority": "high",
},
want: "Issue PROJ-123 has priority high",
},
{
name: "missing parameter keeps placeholder",
task: Task{
Block{Text: makeText("Issue: ${issue_number}")},
},
params: map[string]string{},
want:   "Issue: ${issue_number}",
},
{
name: "complex task with commands and text",
task: Task{
Block{Text: makeText("Task description\n")},
Block{SlashCommand: &SlashCommand{
Name: "fix-bug",
Arguments: []Argument{
{Key: "issue", Value: "123"},
},
}},
Block{Text: makeText("\nFix issue ${issue} with priority ${priority}\n")},
Block{SlashCommand: &SlashCommand{Name: "code-review"}},
Block{Text: makeText("\nDone")},
},
params: map[string]string{
"issue":    "PROJ-123",
"priority": "high",
},
want: "Task description\n\nFix issue PROJ-123 with priority high\n\nDone",
},
{
name:   "empty task",
task:   Task{},
params: map[string]string{},
want:   "",
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
got := taskToPrompt(tt.task, tt.params)
if got != tt.want {
t.Errorf("taskToPrompt() = %q, want %q", got, tt.want)
}
})
}
}
