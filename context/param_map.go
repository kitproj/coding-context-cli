package context

import (
	"fmt"
	"strings"
)

// ParamMap represents a map of parameters for substitution in task prompts
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
