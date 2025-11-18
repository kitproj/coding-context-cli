package context

import (
	"fmt"
	"strings"
)

// ParamMap stores parameter key-value pairs for template substitution
type ParamMap map[string]string

func (p *ParamMap) String() string {
	return fmt.Sprint(*p)
}

func (p *ParamMap) Set(value string) error {
	kv := strings.SplitN(value, "=", 2)
	if len(kv) != 2 {
		return fmt.Errorf("invalid parameter format: %s", value)
	}
	if *p == nil {
		*p = make(map[string]string)
	}
	(*p)[kv[0]] = kv[1]
	return nil
}
