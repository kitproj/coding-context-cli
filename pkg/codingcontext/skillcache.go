package codingcontext

import (
	"context"
	"errors"
	"sync"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
)

// skillParseResult is the cached outcome of parsing a single SKILL.md file:
// either the parsed frontmatter, or the error that rejected it. Exactly one
// of the two is populated, matched by err == nil — on error the frontmatter
// is left zero rather than stored half-built, since parsing fills it field by
// field and can fail partway.
//
// Only deterministic outcomes reach the cache; see isDeterministicParseError.
//
// This is captured immediately after the parse call in loadSkillEntry —
// before validateAndAddSkill's lenient-mode name inference (which derives a
// name from skillFile's directory when frontmatter.Name is empty) — so what
// gets cached is the raw, mode-independent parse result. lenient does not
// affect ParseMarkdownFileWithLogger itself (it takes no such argument), so
// the same skillFile always parses to the same result regardless of which
// call's lenient value happens to populate the cache first.
type skillParseResult struct {
	frontmatter markdown.SkillFrontMatter
	err         error
}

// SkillCache memoizes SKILL.md parse results across repeated skill discovery
// runs, keyed by file path. Safe for concurrent use.
//
// discoverSkills (via New/Run/Lint) reruns a full parse of every skill file
// on every call — and codegen-runner's lint-all-tasks loop constructs a new
// *Context (and so triggers a fresh discoverSkills) once per task, re-parsing
// the same, unchanged skill tree from scratch each time. A SkillCache shared
// across those Context instances, via context.Context (see
// NewContextWithSkillCache), turns that into "parse once across the whole
// run" without changing any single Context's own behavior.
//
// Frontmatter values are safe to share this way: markdown.SkillFrontMatter's
// only reference fields (Content, Metadata maps) are read-only in every
// production code path once loadSkillEntry returns them — the maps are never
// written after that point, so sharing the map underneath a value-copied
// struct across goroutines/tasks is safe without a deep copy.
type SkillCache struct {
	mu      sync.RWMutex
	results map[string]skillParseResult
}

// NewSkillCache returns an empty, ready-to-use SkillCache.
func NewSkillCache() *SkillCache {
	return &SkillCache{results: make(map[string]skillParseResult)}
}

// get returns the cached parse result for skillFile, if any.
func (c *SkillCache) get(skillFile string) (skillParseResult, bool) {
	if c == nil {
		return skillParseResult{}, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	result, ok := c.results[skillFile]

	return result, ok
}

// set stores the parse result for skillFile, replacing any existing entry.
func (c *SkillCache) set(skillFile string, result skillParseResult) {
	if c == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.results[skillFile] = result
}

// isDeterministicParseError reports whether err is a verdict on a file's
// content rather than a report about the I/O that tried to reach it.
//
// A *markdown.ParseError means the bytes were read and rejected, so re-reading
// the same file yields the same verdict and the result is safe to memoize.
// Everything else — permission denied, too many open files, a network
// filesystem hiccup, a file that vanished mid-walk — describes one attempt,
// not the file, and re-reading may well succeed. Caching those would let a
// single transient failure early in a run persist for every later caller.
func isDeterministicParseError(err error) bool {
	var parseErr *markdown.ParseError

	return errors.As(err, &parseErr)
}

// skillCacheContextKey is an unexported type so this package's context key
// cannot collide with keys from other packages using the same value.
type skillCacheContextKey struct{}

// NewContextWithSkillCache returns a child context carrying cache for
// discoverSkills to use. Callers that construct multiple *Context instances
// against the same skill tree (e.g. once per task in a lint-all loop) should
// create one SkillCache with NewSkillCache, attach it once via this
// function, and pass the resulting context.Context into every Run/Lint call
// to share parsed skill files across all of them.
//
// Passing a context without an attached SkillCache (e.g. context.Background()
// unmodified) is fully supported: discoverSkills falls back to parsing every
// call, exactly as before this cache existed.
func NewContextWithSkillCache(ctx context.Context, cache *SkillCache) context.Context {
	return context.WithValue(ctx, skillCacheContextKey{}, cache)
}

// skillCacheFromContext extracts the SkillCache attached via
// NewContextWithSkillCache, or nil if none was attached. All SkillCache
// methods are nil-receiver-safe (get returns !ok, set is a no-op), so callers
// never need to nil-check the result themselves.
func skillCacheFromContext(ctx context.Context) *SkillCache {
	cache, _ := ctx.Value(skillCacheContextKey{}).(*SkillCache)

	return cache
}
