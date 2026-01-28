package codingcontext

import (
	"log/slog"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/selectors"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
)

// Option is a functional option for configuring a Context
type Option func(*Context)

// WithParams sets the parameters
func WithParams(params taskparser.Params) Option {
	return func(c *Context) {
		c.params = params
	}
}

// WithSelectors sets the selectors
func WithSelectors(selectors selectors.Selectors) Option {
	return func(c *Context) {
		c.includes = selectors
	}
}

// WithManifestURL sets the manifest URL
func WithManifestURL(manifestURL string) Option {
	return func(c *Context) {
		c.manifestURL = manifestURL
	}
}

// WithSearchPaths adds one or more search paths
func WithSearchPaths(paths ...string) Option {
	return func(c *Context) {
		c.searchPaths = append(c.searchPaths, paths...)
	}
}

// WithLogger sets the logger
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

// WithAgent sets the target agent, which excludes that agent's own rules
func WithAgent(agent Agent) Option {
	return func(c *Context) {
		c.agent = agent
	}
}

// WithUserPrompt sets the user prompt to append to the task
func WithUserPrompt(userPrompt string) Option {
	return func(c *Context) {
		c.userPrompt = userPrompt
	}
}
