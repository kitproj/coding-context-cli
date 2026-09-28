package codingcontext

import (
	"fmt"
	"strings"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext/markdown"
)

// CallerIdentity identifies the requester invoking a task.
type CallerIdentity struct {
	Username string
	Email    string
}

func (cc *Context) authorizeTaskInvocation(frontMatter markdown.TaskFrontMatter) error {
	if !isAllowed(frontMatter.AllowedSurfaces, cc.surface) {
		return fmt.Errorf("%w: surface is not permitted", ErrCallerNotAllowed)
	}

	if !isAllowed(frontMatter.AllowedRequesters, cc.caller.Username, cc.caller.Email) {
		return fmt.Errorf("%w: requester identity is not permitted", ErrCallerNotAllowed)
	}

	return nil
}

func isAllowed(allowlist []string, values ...string) bool {
	if len(allowlist) == 0 {
		return true
	}

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		for _, allowed := range allowlist {
			if strings.EqualFold(value, strings.TrimSpace(allowed)) {
				return true
			}
		}
	}

	return false
}
