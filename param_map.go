package main

import (
	"fmt"
	"strings"
)

type Params map[string]string

func (p *Params) String() string {
	return fmt.Sprint(*p)
}

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

// ParseParams parses a string in the format "key=value" and returns a Params map.
// If the string is not in the correct format, it returns an error.
func ParseParams(value string) (Params, error) {
	p := make(Params)
	if err := p.Set(value); err != nil {
		return nil, err
	}
	return p, nil
}
