package mcp

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMCPServerConfig_YAML_ArbitraryFields(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    MCPServerConfig
		wantErr bool
	}{
		{
			name: "standard fields only",
			yaml: `type: stdio
command: filesystem
args: ["--verbose"]
`,
			want: MCPServerConfig{
				Type:    TransportTypeStdio,
				Command: "filesystem",
				Args:    []string{"--verbose"},
				Content: map[string]any{
					"type":    "stdio",
					"command": "filesystem",
					"args":    []any{"--verbose"},
				},
			},
		},
		{
			name: "standard fields plus arbitrary fields",
			yaml: `type: stdio
command: git
custom_field: custom_value
max_retries: 3
debug: true
`,
			want: MCPServerConfig{
				Type:    TransportTypeStdio,
				Command: "git",
				Content: map[string]any{
					"type":         "stdio",
					"command":      "git",
					"custom_field": "custom_value",
					"max_retries":  3,
					"debug":        true,
				},
			},
		},
		{
			name: "http type with custom fields",
			yaml: `type: http
url: https://api.example.com
headers:
  Authorization: Bearer token123
timeout_seconds: 30
retry_policy: exponential
`,
			want: MCPServerConfig{
				Type: TransportTypeHTTP,
				URL:  "https://api.example.com",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
				},
				Content: map[string]any{
					"type": "http",
					"url":  "https://api.example.com",
					"headers": map[string]any{
						"Authorization": "Bearer token123",
					},
					"timeout_seconds": 30,
					"retry_policy":    "exponential",
				},
			},
		},
		{
			name: "arbitrary fields with nested objects",
			yaml: `type: stdio
command: database
custom_config:
  host: localhost
  port: 5432
  ssl: true
`,
			want: MCPServerConfig{
				Type:    TransportTypeStdio,
				Command: "database",
				Content: map[string]any{
					"type":    "stdio",
					"command": "database",
					"custom_config": map[string]any{
						"host": "localhost",
						"port": 5432,
						"ssl":  true,
					},
				},
			},
		},
		{
			name: "env variables with custom fields",
			yaml: `type: stdio
command: python
args: ["-m", "server"]
env:
  PYTHON_PATH: /usr/bin/python3
  DEBUG: "true"
python_version: "3.11"
`,
			want: MCPServerConfig{
				Type:    TransportTypeStdio,
				Command: "python",
				Args:    []string{"-m", "server"},
				Env: map[string]string{
					"PYTHON_PATH": "/usr/bin/python3",
					"DEBUG":       "true",
				},
				Content: map[string]any{
					"type":    "stdio",
					"command": "python",
					"args":    []any{"-m", "server"},
					"env": map[string]any{
						"PYTHON_PATH": "/usr/bin/python3",
						"DEBUG":       "true",
					},
					"python_version": "3.11",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got MCPServerConfig
			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare standard fields
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Command != tt.want.Command {
				t.Errorf("Command = %v, want %v", got.Command, tt.want.Command)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %v, want %v", got.URL, tt.want.URL)
			}

			// Compare Content map - at least verify keys exist
			for key := range tt.want.Content {
				if _, exists := got.Content[key]; !exists {
					t.Errorf("Content missing key %q", key)
				}
			}
		})
	}
}

func TestMCPServerConfig_JSON_ArbitraryFields(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    MCPServerConfig
		wantErr bool
	}{
		{
			name: "standard fields only",
			json: `{
				"type": "stdio",
				"command": "filesystem"
			}`,
			want: MCPServerConfig{
				Type:    TransportTypeStdio,
				Command: "filesystem",
				Content: map[string]any{
					"type":    "stdio",
					"command": "filesystem",
				},
			},
		},
		{
			name: "standard fields plus arbitrary fields",
			json: `{
				"type": "stdio",
				"command": "git",
				"custom_field": "custom_value",
				"max_retries": 3
			}`,
			want: MCPServerConfig{
				Type:    TransportTypeStdio,
				Command: "git",
				Content: map[string]any{
					"type":         "stdio",
					"command":      "git",
					"custom_field": "custom_value",
					"max_retries":  float64(3), // JSON numbers unmarshal as float64
				},
			},
		},
		{
			name: "http type with custom metadata",
			json: `{
				"type": "http",
				"url": "https://api.example.com",
				"api_version": "v1",
				"timeout_ms": 5000
			}`,
			want: MCPServerConfig{
				Type: TransportTypeHTTP,
				URL:  "https://api.example.com",
				Content: map[string]any{
					"type":        "http",
					"url":         "https://api.example.com",
					"api_version": "v1",
					"timeout_ms":  float64(5000),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got MCPServerConfig
			err := json.Unmarshal([]byte(tt.json), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Compare standard fields
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Command != tt.want.Command {
				t.Errorf("Command = %v, want %v", got.Command, tt.want.Command)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %v, want %v", got.URL, tt.want.URL)
			}

			// Compare Content map - at least verify keys exist
			for key := range tt.want.Content {
				if _, exists := got.Content[key]; !exists {
					t.Errorf("Content missing key %q", key)
				}
			}
		})
	}
}

func TestMCPServerConfig_Marshal_YAML(t *testing.T) {
	tests := []struct {
		name   string
		config MCPServerConfig
	}{
		{
			name: "standard fields only",
			config: MCPServerConfig{
				Type:    TransportTypeStdio,
				Command: "filesystem",
			},
		},
		{
			name: "with arbitrary fields in Content",
			config: MCPServerConfig{
				Type:    TransportTypeStdio,
				Command: "git",
				Content: map[string]any{
					"type":         "stdio",
					"command":      "git",
					"custom_field": "custom_value",
					"max_retries":  3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.config)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			// Unmarshal to verify round-trip
			var got MCPServerConfig
			if err := yaml.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			// Verify standard fields match
			if got.Type != tt.config.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.config.Type)
			}
			if got.Command != tt.config.Command {
				t.Errorf("Command = %v, want %v", got.Command, tt.config.Command)
			}
		})
	}
}

func TestMCPServerConfigs_WithArbitraryFields(t *testing.T) {
	yamlContent := `
filesystem:
  type: stdio
  command: filesystem
  cache_enabled: true
git:
  type: stdio
  command: git
  max_depth: 10
api:
  type: http
  url: https://api.example.com
  rate_limit: 100
`

	var configs MCPServerConfigs
	err := yaml.Unmarshal([]byte(yamlContent), &configs)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(configs) != 3 {
		t.Errorf("Expected 3 configs, got %d", len(configs))
	}

	// Check filesystem config
	if fs, ok := configs["filesystem"]; ok {
		if fs.Type != TransportTypeStdio {
			t.Errorf("filesystem.Type = %v, want %v", fs.Type, TransportTypeStdio)
		}
		if fs.Command != "filesystem" {
			t.Errorf("filesystem.Command = %v, want %v", fs.Command, "filesystem")
		}
		if fs.Content["cache_enabled"] != true {
			t.Errorf("filesystem.Content[cache_enabled] = %v, want true", fs.Content["cache_enabled"])
		}
	} else {
		t.Error("filesystem config not found")
	}

	// Check git config
	if git, ok := configs["git"]; ok {
		if git.Type != TransportTypeStdio {
			t.Errorf("git.Type = %v, want %v", git.Type, TransportTypeStdio)
		}
		// Verify custom field exists
		if _, exists := git.Content["max_depth"]; !exists {
			t.Error("git.Content[max_depth] not found")
		}
	} else {
		t.Error("git config not found")
	}

	// Check api config
	if api, ok := configs["api"]; ok {
		if api.Type != TransportTypeHTTP {
			t.Errorf("api.Type = %v, want %v", api.Type, TransportTypeHTTP)
		}
		if api.URL != "https://api.example.com" {
			t.Errorf("api.URL = %v, want %v", api.URL, "https://api.example.com")
		}
		// Verify custom field exists
		if _, exists := api.Content["rate_limit"]; !exists {
			t.Error("api.Content[rate_limit] not found")
		}
	} else {
		t.Error("api config not found")
	}
}
