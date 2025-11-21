package config

import (
	"encoding/json"
	"testing"
)

func TestConfig_Marshal(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{
			name: "empty config",
			config: Config{
				MCPServers: map[string]MCPServerConfig{},
			},
			want: `{"mcpServers":{}}`,
		},
		{
			name: "stdio server",
			config: Config{
				MCPServers: map[string]MCPServerConfig{
					"filesystem": {
						Type:    TransportTypeStdio,
						Command: "npx",
						Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
					},
				},
			},
			want: `{"mcpServers":{"filesystem":{"type":"stdio","command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","/tmp"]}}}`,
		},
		{
			name: "http server with headers",
			config: Config{
				MCPServers: map[string]MCPServerConfig{
					"api": {
						Type: TransportTypeHTTP,
						URL:  "https://api.example.com/mcp",
						Headers: map[string]string{
							"Authorization": "Bearer token123",
						},
					},
				},
			},
			want: `{"mcpServers":{"api":{"type":"http","url":"https://api.example.com/mcp","headers":{"Authorization":"Bearer token123"}}}}`,
		},
		{
			name: "multiple servers",
			config: Config{
				MCPServers: map[string]MCPServerConfig{
					"filesystem": {
						Type:    TransportTypeStdio,
						Command: "npx",
						Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
					},
					"git": {
						Type:    TransportTypeStdio,
						Command: "npx",
						Args:    []string{"-y", "@modelcontextprotocol/server-git"},
					},
				},
			},
			// Note: map order is not guaranteed in JSON, so we just check it's valid
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(&tt.config)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if tt.want != "" && string(got) != tt.want {
				t.Errorf("Marshal() = %s, want %s", string(got), tt.want)
			}
			// Verify it's valid JSON at least
			if !json.Valid(got) {
				t.Errorf("Marshal() produced invalid JSON: %s", string(got))
			}
		})
	}
}

func TestConfig_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "claude desktop stdio config",
			json: `{
				"mcpServers": {
					"filesystem": {
						"type": "stdio",
						"command": "npx",
						"args": ["-y", "@modelcontextprotocol/server-filesystem", "/Users/username"]
					}
				}
			}`,
			wantErr: false,
		},
		{
			name: "cursor config with env vars",
			json: `{
				"mcpServers": {
					"postgres": {
						"type": "stdio",
						"command": "docker",
						"args": ["run", "-i", "--rm", "mcp-server-postgres"],
						"env": {
							"DB_HOST": "localhost",
							"DB_PORT": "5432"
						}
					}
				}
			}`,
			wantErr: false,
		},
		{
			name: "http server",
			json: `{
				"mcpServers": {
					"remote": {
						"type": "http",
						"url": "https://api.example.com/mcp",
						"headers": {
							"Authorization": "Bearer secret"
						}
					}
				}
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Config
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			// Verify we got some servers
			if len(got.MCPServers) == 0 {
				t.Error("Unmarshal() resulted in empty MCPServers")
			}
		})
	}
}

func TestMCPServerConfig_Marshal(t *testing.T) {
	tests := []struct {
		name   string
		config MCPServerConfig
		want   string
	}{
		{
			name: "stdio with env",
			config: MCPServerConfig{
				Type:    TransportTypeStdio,
				Command: "python",
				Args:    []string{"-m", "mcp_server"},
				Env: map[string]string{
					"API_KEY": "secret",
				},
			},
			want: `{"type":"stdio","command":"python","args":["-m","mcp_server"],"env":{"API_KEY":"secret"}}`,
		},
		{
			name: "minimal http",
			config: MCPServerConfig{
				Type: TransportTypeHTTP,
				URL:  "http://localhost:8080",
			},
			want: `{"type":"http","url":"http://localhost:8080"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(&tt.config)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

func TestTransportType_Values(t *testing.T) {
	tests := []struct {
		name  string
		value TransportType
		want  string
	}{
		{"stdio", TransportTypeStdio, "stdio"},
		{"sse", TransportTypeSSE, "sse"},
		{"http", TransportTypeHTTP, "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.want {
				t.Errorf("TransportType value = %s, want %s", tt.value, tt.want)
			}
		})
	}
}
