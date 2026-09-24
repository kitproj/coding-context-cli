package codingcontext

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRun_TaskInvocationAuthorization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		frontmatter string
		opts        []Option
		wantErr     bool
	}{
		{
			name: "unrestricted task",
		},
		{
			name: "empty allowlists leave task unrestricted",
			frontmatter: `allowed_surfaces: []
allowed_requesters: []`,
		},
		{
			name:        "surface matches case insensitively",
			frontmatter: "allowed_surfaces: [jira, github]",
			opts:        []Option{WithSurface(" JIRA ")},
		},
		{
			name:        "requester username matches",
			frontmatter: "allowed_requesters: [alice]",
			opts: []Option{WithCaller(CallerIdentity{
				Username: "ALICE",
				Email:    "other@example.com",
			})},
		},
		{
			name:        "requester email matches",
			frontmatter: "allowed_requesters: [alice@example.com]",
			opts: []Option{WithCaller(CallerIdentity{
				Username: "other",
				Email:    "Alice@Example.com",
			})},
		},
		{
			name: "surface and requester both match",
			frontmatter: `allowed_surfaces: [github]
allowed_requesters: [alice, alice@example.com]`,
			opts: []Option{
				WithSurface("github"),
				WithCaller(CallerIdentity{Username: "alice"}),
			},
		},
		{
			name:        "surface mismatch",
			frontmatter: "allowed_surfaces: [jira]",
			opts:        []Option{WithSurface("github")},
			wantErr:     true,
		},
		{
			name:        "surface missing",
			frontmatter: "allowed_surfaces: [jira]",
			wantErr:     true,
		},
		{
			name:        "requester mismatch",
			frontmatter: "allowed_requesters: [alice]",
			opts:        []Option{WithCaller(CallerIdentity{Username: "bob", Email: "bob@example.com"})},
			wantErr:     true,
		},
		{
			name:        "requester missing",
			frontmatter: "allowed_requesters: [alice]",
			wantErr:     true,
		},
		{
			name: "surface matches but requester does not",
			frontmatter: `allowed_surfaces: [jira]
allowed_requesters: [alice]`,
			opts: []Option{
				WithSurface("jira"),
				WithCaller(CallerIdentity{Username: "bob"}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			createTask(t, dir, "restricted", tt.frontmatter, "Authorized task.")

			opts := append([]Option{WithSearchPaths(dir)}, tt.opts...)
			result, err := New(opts...).Run(context.Background(), "restricted")

			if tt.wantErr {
				require.ErrorIs(t, err, ErrCallerNotAllowed)
				require.Nil(t, result)

				return
			}

			require.NoError(t, err)
			require.Contains(t, result.Prompt, "Authorized task.")
		})
	}
}

func TestRun_AuthorizationPrecedesTaskRendering(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "restricted", "allowed_surfaces: [jira]", "/missing-command")

	result, err := New(
		WithSearchPaths(dir),
		WithSurface("github"),
	).Run(context.Background(), "restricted")

	require.ErrorIs(t, err, ErrCallerNotAllowed)
	require.Nil(t, result)
	require.NotErrorIs(t, err, ErrCommandNotFound)
}

func TestRun_LenientSearchPathDoesNotSuppressAuthorizationDenial(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "restricted", "allowed_surfaces: [jira]", "Restricted task.")

	result, err := New(
		WithLenientSearchPaths(dir),
		WithSurface("github"),
	).Run(context.Background(), "restricted")

	require.ErrorIs(t, err, ErrCallerNotAllowed)
	require.Nil(t, result)
}

func TestLint_RestrictedTaskDoesNotRequireInvocationIdentity(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	createTask(t, dir, "restricted", `allowed_surfaces: [jira]
allowed_requesters: [alice]`, "Lint this task.")

	result, err := New(WithSearchPaths(dir)).Lint(context.Background(), "restricted")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, []string{"jira"}, result.Task.FrontMatter.AllowedSurfaces)
	require.Equal(t, []string{"alice"}, result.Task.FrontMatter.AllowedRequesters)
}
