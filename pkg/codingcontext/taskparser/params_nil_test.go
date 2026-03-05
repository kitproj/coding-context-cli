package taskparser_test

import (
	"testing"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
	gparser "github.com/yuin/goldmark/parser"
)

// TestParams_Arguments_Nil verifies that Arguments() on a nil Params returns
// nil without panicking, exercising the nil guard branch.
func TestParams_Arguments_Nil(t *testing.T) {
	t.Parallel()

	var p taskparser.Params

	got := p.Arguments()

	if got != nil {
		t.Errorf("Arguments() on nil Params = %v, want nil", got)
	}
}

// TestParams_Arguments_Empty verifies that Arguments() returns nil when there
// are no positional arguments in the map.
func TestParams_Arguments_Empty(t *testing.T) {
	t.Parallel()

	p, err := taskparser.ParseParams("key=value")
	if err != nil {
		t.Fatalf("ParseParams error: %v", err)
	}

	got := p.Arguments()
	if got != nil {
		t.Errorf("Arguments() with no positional args = %v, want nil", got)
	}
}

// TestParams_Arguments_Positional verifies that positional arguments are accessible
// via Arguments() and are distinct from named parameters.
func TestParams_Arguments_Positional(t *testing.T) {
	t.Parallel()

	p, err := taskparser.ParseParams("foo bar")
	if err != nil {
		t.Fatalf("ParseParams error: %v", err)
	}

	args := p.Arguments()
	if len(args) == 0 {
		t.Error("Arguments() should return positional args, got none")
	}
}

// TestParams_Lookup_Nil verifies that Lookup() on a nil Params returns ("", false)
// without panicking.
func TestParams_Lookup_Nil(t *testing.T) {
	t.Parallel()

	var p taskparser.Params

	v, ok := p.Lookup("key")

	if ok || v != "" {
		t.Errorf("Lookup() on nil Params = (%q, %v), want (\"\", false)", v, ok)
	}
}

// TestParams_Lookup_MissingKey verifies that Lookup() returns ("", false) when
// the key is not present in a non-nil Params.
func TestParams_Lookup_MissingKey(t *testing.T) {
	t.Parallel()

	p, err := taskparser.ParseParams("key=value")
	if err != nil {
		t.Fatalf("ParseParams error: %v", err)
	}

	v, ok := p.Lookup("missing")
	if ok || v != "" {
		t.Errorf("Lookup(missing) = (%q, %v), want (\"\", false)", v, ok)
	}
}

// TestGetTask_NoExtension verifies that GetTask returns (nil, nil) when called
// on a parser context where Extension was never registered. This exercises the
// v == nil guard at the top of GetTask.
func TestGetTask_NoExtension(t *testing.T) {
	t.Parallel()

	pctx := gparser.NewContext()

	task, err := taskparser.GetTask(pctx)
	if err != nil {
		t.Errorf("GetTask(fresh context) error = %v, want nil", err)
	}

	if task != nil {
		t.Errorf("GetTask(fresh context) task = %v, want nil", task)
	}
}
