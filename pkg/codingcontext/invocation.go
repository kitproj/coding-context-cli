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
	if len(frontMatter.AllowedSurfaces) > 0 {
		if strings.TrimSpace(cc.surface) == "" {
			return fmt.Errorf("%w: surface is required by the task allowlist", ErrCallerNotAllowed)
		}

		if !matchesAllowedValue(frontMatter.AllowedSurfaces, cc.surface) {
			return fmt.Errorf("%w: surface %q is not permitted", ErrCallerNotAllowed, cc.surface)
		}
	}

	if len(frontMatter.AllowedRequesters) == 0 {
		return nil
	}

	if strings.TrimSpace(cc.caller.Username) == "" && strings.TrimSpace(cc.caller.Email) == "" {
		return fmt.Errorf("%w: requester identity is required by the task allowlist", ErrCallerNotAllowed)
	}

	if matchesAllowedValue(frontMatter.AllowedRequesters, cc.caller.Username, cc.caller.Email) {
		return nil
	}

	return fmt.Errorf("%w: requester %s is not permitted", ErrCallerNotAllowed, cc.caller.description())
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

func (c CallerIdentity) description() string {
	username := strings.TrimSpace(c.Username)
	email := strings.TrimSpace(c.Email)

	switch {
	case username != "" && email != "":
		return fmt.Sprintf("%q <%s>", username, email)
	case username != "":
		return fmt.Sprintf("%q", username)
	default:
		return fmt.Sprintf("%q", email)
	}
}
