package mcp

import (
	"encoding/json"
	"testing"

	yaml "github.com/goccy/go-yaml"
)

// assertMCPConfig checks common MCPServerConfig fields against expected values.
func assertMCPConfig(t *testing.T, got, want MCPServerConfig, err error, wantErr bool) {
	t.Helper()

	if (err != nil) != wantErr {
		t.Errorf("Unmarshal() error = %v, wantErr %v", err, wantErr)

		return
	}

	if got.Type != want.Type {
		t.Errorf("Type = %v, want %v", got.Type, want.Type)
	}

	if got.Command != want.Command {
		t.Errorf("Command = %v, want %v", got.Command, want.Command)
	}

	if got.URL != want.URL {
		t.Errorf("URL = %v, want %v", got.URL, want.URL)
	}

	for key := range want.Content {
		if _, exists := got.Content[key]; !exists {
			t.Errorf("Content missing key %q", key)
		}
	}
}

// requireConfig retrieves a named config from MCPServerConfigs or fails the test.
func requireConfig(t *testing.T, configs MCPServerConfigs, name string) MCPServerConfig {
	t.Helper()

	cfg, ok := configs[name]
	if !ok {
		t.Fatalf("%s config not found", name)
	}

	return cfg
}

func TestMCPServerConfig_YAML_ArbitraryFields(t *testing.T) {
	t.Parallel()

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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got MCPServerConfig

			err := yaml.Unmarshal([]byte(tt.yaml), &got)

			assertMCPConfig(t, got, tt.want, err, tt.wantErr)
		})
	}
}

func TestMCPServerConfig_JSON_ArbitraryFields(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			var got MCPServerConfig

			err := json.Unmarshal([]byte(tt.json), &got)

			assertMCPConfig(t, got, tt.want, err, tt.wantErr)
		})
	}
}

func TestMCPServerConfig_Marshal_YAML(t *testing.T) {
	t.Parallel()

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
					"custom_field": "custom_value",
					"max_retries":  3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
	t.Parallel()

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

	checkFilesystemConfig(t, configs)
	checkGitConfig(t, configs)
	checkAPIConfig(t, configs)
}

func checkFilesystemConfig(t *testing.T, configs MCPServerConfigs) {
	t.Helper()

	fs := requireConfig(t, configs, "filesystem")

	if fs.Type != TransportTypeStdio {
		t.Errorf("filesystem.Type = %v, want %v", fs.Type, TransportTypeStdio)
	}

	if fs.Command != "filesystem" {
		t.Errorf("filesystem.Command = %v, want %v", fs.Command, "filesystem")
	}

	if fs.Content["cache_enabled"] != true {
		t.Errorf("filesystem.Content[cache_enabled] = %v, want true", fs.Content["cache_enabled"])
	}
}

func checkGitConfig(t *testing.T, configs MCPServerConfigs) {
	t.Helper()

	git := requireConfig(t, configs, "git")

	if git.Type != TransportTypeStdio {
		t.Errorf("git.Type = %v, want %v", git.Type, TransportTypeStdio)
	}

	if _, exists := git.Content["max_depth"]; !exists {
		t.Error("git.Content[max_depth] not found")
	}
}

func checkAPIConfig(t *testing.T, configs MCPServerConfigs) {
	t.Helper()

	api := requireConfig(t, configs, "api")

	if api.Type != TransportTypeHTTP {
		t.Errorf("api.Type = %v, want %v", api.Type, TransportTypeHTTP)
	}

	if api.URL != "https://api.example.com" {
		t.Errorf("api.URL = %v, want %v", api.URL, "https://api.example.com")
	}

	if _, exists := api.Content["rate_limit"]; !exists {
		t.Error("api.Content[rate_limit] not found")
	}
}

func TestMCPServerConfig_JSON_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, cfg MCPServerConfig)
	}{
		{
			name:  "empty JSON object gives zero-value struct",
			input: `{}`,
			validate: func(t *testing.T, cfg MCPServerConfig) {
				t.Helper()

				if cfg.Type != "" {
					t.Errorf("Type = %q, want empty", cfg.Type)
				}

				if cfg.Command != "" {
					t.Errorf("Command = %q, want empty", cfg.Command)
				}

				if cfg.Content == nil {
					t.Error("Content should be non-nil empty map for {}")
				}
			},
		},
		{
			name:  "only unknown fields go to Content map",
			input: `{"foo": "bar", "num": 42, "flag": true}`,
			validate: func(t *testing.T, cfg MCPServerConfig) {
				t.Helper()

				if cfg.Type != "" {
					t.Errorf("Type = %q, want empty", cfg.Type)
				}

				if cfg.Content == nil {
					t.Fatal("Content should not be nil")
				}

				if cfg.Content["foo"] != "bar" {
					t.Errorf("Content[foo] = %v, want bar", cfg.Content["foo"])
				}

				if cfg.Content["num"] != float64(42) {
					t.Errorf("Content[num] = %v, want 42", cfg.Content["num"])
				}

				if cfg.Content["flag"] != true {
					t.Errorf("Content[flag] = %v, want true", cfg.Content["flag"])
				}
			},
		},
		{
			name:    "invalid JSON returns error",
			input:   `{not valid`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cfg MCPServerConfig

			err := json.Unmarshal([]byte(tt.input), &cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil && tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}
