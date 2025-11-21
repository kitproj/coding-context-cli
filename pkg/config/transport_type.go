package config

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
