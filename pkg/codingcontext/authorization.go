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
	if err := authorizeSurface(frontMatter.AllowedSurfaces, cc.surface); err != nil {
		return err
	}

	return authorizeRequester(frontMatter.AllowedRequesters, cc.caller)
}

func authorizeSurface(allowed []string, surface string) error {
	if len(allowed) == 0 {
		return nil
	}

	surface = strings.TrimSpace(surface)
	if matchesAllowedValue(allowed, surface) {
		return nil
	}

	return fmt.Errorf("%w: surface %q is not permitted", ErrCallerNotAllowed, surface)
}

func authorizeRequester(allowed []string, caller CallerIdentity) error {
	if len(allowed) == 0 {
		return nil
	}

	if !matchesAllowedValue(allowed, caller.Username, caller.Email) {
		return fmt.Errorf("%w: requester identity is not permitted", ErrCallerNotAllowed)
	}

	return nil
}

func matchesAllowedValue(allowed []string, actualValues ...string) bool {
	for _, actual := range actualValues {
		actual = strings.TrimSpace(actual)
		if actual == "" {
			continue
		}

		for _, candidate := range allowed {
			if strings.EqualFold(actual, strings.TrimSpace(candidate)) {
				return true
			}
		}
	}

	return false
}
