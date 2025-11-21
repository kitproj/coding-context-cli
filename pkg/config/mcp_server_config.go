package config

// MCPServerConfig defines the common configuration fields supported by both platforms.
type MCPServerConfig struct {
	// Type specifies the connection protocol.
	// Values: "stdio", "sse", "http".
	Type TransportType `json:"type,omitempty"`

	// Command is the executable to run (e.g. "npx", "docker").
	// Required for "stdio" type.
	Command string `json:"command,omitempty"`

	// Args is an array of arguments for the command.
	Args []string `json:"args,omitempty"`

	// Env defines environment variables for the server process.
	Env map[string]string `json:"env,omitempty"`

	// URL is the endpoint for "http" or "sse" types.
	// Required for remote connections.
	URL string `json:"url,omitempty"`

	// Headers contains custom HTTP headers (e.g. {"Authorization": "Bearer ..."}).
	// Used for "http" and "sse" types.
	Headers map[string]string `json:"headers,omitempty"`
}
