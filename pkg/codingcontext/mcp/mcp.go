package mcp

import "encoding/json"

// TransportType defines the communication protocol used by the server.
// Supported by both Claude and Cursor.
type TransportType string

const (
	// TransportTypeStdio is for local processes (executables).
	TransportTypeStdio TransportType = "stdio"

	// TransportTypeSSE is for Server-Sent Events (Remote).
	// Note: Claude Code prefers HTTP over SSE, but supports it.
	TransportTypeSSE TransportType = "sse"

	// TransportTypeHTTP is for standard HTTP/POST interactions.
	TransportTypeHTTP TransportType = "http"
)

// MCPServerConfig defines the common configuration fields supported by both platforms.
// It also supports arbitrary additional fields via the Content map.
type MCPServerConfig struct {
	// Type specifies the connection protocol.
	// Values: "stdio", "sse", "http".
	Type TransportType `json:"type,omitempty" yaml:"type,omitempty"`

	// Command is the executable to run (e.g. "npx", "docker").
	// Required for "stdio" type.
	Command string `json:"command,omitempty" yaml:"command,omitempty"`

	// Args is an array of arguments for the command.
	Args []string `json:"args,omitempty" yaml:"args,omitempty"`

	// Env defines environment variables for the server process.
	Env map[string]string `json:"env,omitempty" yaml:"env,omitempty"`

	// URL is the endpoint for "http" or "sse" types.
	// Required for remote connections.
	URL string `json:"url,omitempty" yaml:"url,omitempty"`

	// Headers contains custom HTTP headers (e.g. {"Authorization": "Bearer ..."}).
	// Used for "http" and "sse" types.
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`

	// Content holds arbitrary additional fields from YAML/JSON that aren't in the struct
	Content map[string]any `json:"-" yaml:",inline"`
}

// UnmarshalJSON custom unmarshaler that populates both typed fields and Content map
func (m *MCPServerConfig) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary type to avoid infinite recursion
	type Alias MCPServerConfig
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	// Initialize Content map if needed
	if m.Content == nil {
		m.Content = make(map[string]any)
	}

	// Also unmarshal into Content map
	if err := json.Unmarshal(data, &m.Content); err != nil {
		return err
	}

	return nil
}

// MCPServerConfigs maps server names to their configurations.
type MCPServerConfigs map[string]MCPServerConfig
