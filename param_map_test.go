package main

import (
	"testing"
)

func TestParamMap_Set(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantKey string
		wantVal string
		wantErr bool
	}{
		{
			name:    "valid key=value",
			value:   "key=value",
			wantKey: "key",
			wantVal: "value",
			wantErr: false,
		},
		{
			name:    "key=value with equals in value",
			value:   "key=value=with=equals",
			wantKey: "key",
			wantVal: "value=with=equals",
			wantErr: false,
		},
		{
			name:    "empty value",
			value:   "key=",
			wantKey: "key",
			wantVal: "",
			wantErr: false,
		},
		{
			name:    "invalid format - no equals",
			value:   "keyvalue",
			wantErr: true,
		},
		{
			name:    "invalid format - only key",
			value:   "key",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := paramMap{}
			err := p.Set(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("paramMap.Set() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if p[tt.wantKey] != tt.wantVal {
					t.Errorf("paramMap[%q] = %q, want %q", tt.wantKey, p[tt.wantKey], tt.wantVal)
				}
			}
		})
	}
}

func TestParamMap_String(t *testing.T) {
	p := paramMap{
		"key1": "value1",
		"key2": "value2",
	}
	s := p.String()
	if s == "" {
		t.Error("paramMap.String() returned empty string")
	}
}

func TestParamMap_SetMultiple(t *testing.T) {
	p := paramMap{}

	if err := p.Set("key1=value1"); err != nil {
		t.Fatalf("paramMap.Set() failed: %v", err)
	}
	if err := p.Set("key2=value2"); err != nil {
		t.Fatalf("paramMap.Set() failed: %v", err)
	}

	if len(p) != 2 {
		t.Errorf("paramMap length = %d, want 2", len(p))
	}
	if p["key1"] != "value1" {
		t.Errorf("paramMap[key1] = %q, want %q", p["key1"], "value1")
	}
	if p["key2"] != "value2" {
		t.Errorf("paramMap[key2] = %q, want %q", p["key2"], "value2")
	}
}
