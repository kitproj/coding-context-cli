package main

import (
	"testing"
)

func TestStringSlice_Set(t *testing.T) {
	s := stringSlice{}

	values := []string{"first", "second", "third"}
	for _, v := range values {
		if err := s.Set(v); err != nil {
			t.Errorf("stringSlice.Set(%q) error = %v", v, err)
		}
	}

	if len(s) != len(values) {
		t.Errorf("stringSlice length = %d, want %d", len(s), len(values))
	}

	for i, want := range values {
		if s[i] != want {
			t.Errorf("stringSlice[%d] = %q, want %q", i, s[i], want)
		}
	}
}

func TestStringSlice_String(t *testing.T) {
	s := stringSlice{"value1", "value2", "value3"}
	str := s.String()
	if str == "" {
		t.Error("stringSlice.String() returned empty string")
	}
}

func TestStringSlice_SetEmpty(t *testing.T) {
	s := stringSlice{}
	
	if err := s.Set(""); err != nil {
		t.Errorf("stringSlice.Set(\"\") error = %v", err)
	}

	if len(s) != 1 {
		t.Errorf("stringSlice length = %d, want 1", len(s))
	}
	if s[0] != "" {
		t.Errorf("stringSlice[0] = %q, want empty string", s[0])
	}
}
