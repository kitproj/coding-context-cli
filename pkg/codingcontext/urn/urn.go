package urn

import (
"encoding/json"

exturn "github.com/leodido/go-urn"
)

// URN is a wrapper around github.com/leodido/go-urn.URN that adds YAML marshaling support
type URN struct {
*exturn.URN
}

// Parse wraps the external URN Parse function
func Parse(u []byte, options ...exturn.Option) (*URN, bool) {
urn, ok := exturn.Parse(u, options...)
if !ok {
return nil, false
}
return &URN{URN: urn}, true
}

// MarshalYAML implements yaml.Marshaler
func (u *URN) MarshalYAML() (interface{}, error) {
if u == nil || u.URN == nil {
return nil, nil
}
return u.URN.String(), nil
}

// UnmarshalYAML implements yaml.Unmarshaler
func (u *URN) UnmarshalYAML(unmarshal func(interface{}) error) error {
var s string
if err := unmarshal(&s); err != nil {
return err
}

parsed, ok := exturn.Parse([]byte(s))
if !ok {
return &json.UnmarshalTypeError{
Value: "string " + s,
Type:  nil,
}
}

u.URN = parsed
return nil
}

// MarshalJSON implements json.Marshaler
func (u *URN) MarshalJSON() ([]byte, error) {
if u == nil || u.URN == nil {
return []byte("null"), nil
}
return u.URN.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler
func (u *URN) UnmarshalJSON(data []byte) error {
var urn exturn.URN
if err := urn.UnmarshalJSON(data); err != nil {
return err
}
u.URN = &urn
return nil
}

// String returns the URN string representation
func (u *URN) String() string {
if u == nil || u.URN == nil {
return ""
}
return u.URN.String()
}

// Equal compares two URNs for equality
func (u *URN) Equal(other *URN) bool {
if u == nil && other == nil {
return true
}
if u == nil || other == nil {
return false
}
if u.URN == nil && other.URN == nil {
return true
}
if u.URN == nil || other.URN == nil {
return false
}
return u.URN.Equal(other.URN)
}
