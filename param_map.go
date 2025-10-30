package main

import (
	"fmt"
	"strings"
)

type paramMap map[string]string

func (p *paramMap) String() string {
	return fmt.Sprint(*p)
}

func (p *paramMap) Set(value string) error {
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
