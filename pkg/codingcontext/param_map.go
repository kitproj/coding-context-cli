package codingcontext

import (
	"fmt"
	"strings"
)

// ParamMap stores parameter key-value pairs for substitution in task prompts
type ParamMap map[string]string

func (p *ParamMap) String() string {
	return fmt.Sprint(*p)
}

// Set parses a key=value string and adds it to the map
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
