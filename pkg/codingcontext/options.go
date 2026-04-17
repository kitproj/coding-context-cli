package codingcontext

import (
	"log/slog"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/selectors"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
)

// Option is a functional option for configuring a Context.
type Option func(*Context)

// WithParams sets the parameters.
func WithParams(params taskparser.Params) Option {
	return func(c *Context) {
		c.params = params
	}
}

// WithSelectors sets the selectors.
func WithSelectors(selectors selectors.Selectors) Option {
	return func(c *Context) {
		c.includes = selectors
	}
}

// WithManifestURL sets the manifest URL.
func WithManifestURL(manifestURL string) Option {
	return func(c *Context) {
		c.manifestURL = manifestURL
	}
}

// WithSearchPaths adds one or more strict search paths.
// Errors encountered while processing files from these paths are treated as fatal.
func WithSearchPaths(paths ...string) Option {
	return func(c *Context) {
		for _, p := range paths {
			c.searchPaths = append(c.searchPaths, SearchPath{Path: p})
		}
	}
}

// WithLenientSearchPaths adds one or more lenient search paths.
// Errors encountered while processing files from these paths are logged as warnings
// and the problematic files are skipped rather than causing a fatal error.
// For skills with a missing name, the name is inferred from the directory name.
func WithLenientSearchPaths(paths ...string) Option {
	return func(c *Context) {
		for _, p := range paths {
			c.searchPaths = append(c.searchPaths, SearchPath{Path: p, Lenient: true})
		}
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Context) {
		c.logger = logger
	}
}

// WithResume sets the resume selector to "true", which can be used to filter tasks
// by their frontmatter resume field. This does not affect rule/skill discovery or bootstrap scripts.
func WithResume(resume bool) Option {
	return func(c *Context) {
		c.resume = resume
	}
}

// WithBootstrap controls whether to discover rules, skills, and run bootstrap scripts.
// When set to false, rule discovery, skill discovery, and bootstrap script execution are skipped.
func WithBootstrap(doBootstrap bool) Option {
	return func(c *Context) {
		c.doBootstrap = doBootstrap
	}
}

// WithAgent sets the target agent, which excludes that agent's own rules.
// Agent-specific paths are treated as strict (errors are fatal).
// Mutually exclusive with WithLenientAgent.
func WithAgent(agent Agent) Option {
	return func(c *Context) {
		c.agent = agent
		c.strictAgent = agent.IsSet()
	}
}

// WithLenientAgent sets the target agent with lenient error handling.
// Agent-specific paths are treated as lenient: errors are logged as warnings
// and problematic files are skipped rather than causing a fatal error.
// Mutually exclusive with WithAgent.
func WithLenientAgent(agent Agent) Option {
	return func(c *Context) {
		c.agent = agent
		c.lenientAgent = true
	}
}

// WithUserPrompt sets the user prompt to append to the task.
func WithUserPrompt(userPrompt string) Option {
	return func(c *Context) {
		c.userPrompt = userPrompt
	}
}

// WithLint enables lint mode: skips bootstrap script execution and shell command
// expansion (!`cmd`). File access is tracked and non-fatal structural errors are
// collected in LintResult. Use Lint() instead of Run() to retrieve results.
func WithLint(lint bool) Option {
	return func(c *Context) {
		c.lintMode = lint
	}
}
