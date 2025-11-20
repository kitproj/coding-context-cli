package codingcontext

import (
	"fmt"
	"strings"
)

// Params is a map of parameter key-value pairs for template substitution
type Params map[string]string

// String implements the fmt.Stringer interface for Params
func (p *Params) String() string {
	return fmt.Sprint(*p)
}

// Set implements the flag.Value interface for Params
func (p *Params) Set(value string) error {
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
