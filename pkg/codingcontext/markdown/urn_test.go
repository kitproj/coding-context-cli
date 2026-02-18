package markdown

import (
	"github.com/leodido/go-urn"
)

func mustParseURN(s string) *urn.URN {
	u, ok := urn.Parse([]byte(s))
	if !ok {
		panic("invalid urn: " + s)
	}
	return u
}

func urnEqual(a, b *urn.URN) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(b)
}

func urnString(u *urn.URN) string {
	if u == nil {
		return ""
	}
	return u.String()
}
