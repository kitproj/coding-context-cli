package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
)

const toolDefinitions = `[
  {"name":"inspect_readme","description":"Read the repository README.","inputSchema":{"type":"object","properties":{}}},
  {"name":"post_message","description":"Post a message to a channel.","inputSchema":{"type":"object","properties":{"channel":{"type":"string"},"text":{"type":"string"}},"required":["channel","text"]}}
]`

// serve implements the small subset of stdio MCP needed by this evaluation.
// Neither tool performs network requests or accesses repository contents.
func serve(input io.Reader, output io.Writer, logPath string) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	encoder := json.NewEncoder(output)
	for scanner.Scan() {
		var request struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			return err
		}
		if request.ID == nil {
			continue
		}
		reply := map[string]any{"jsonrpc": "2.0", "id": request.ID}
		switch request.Method {
		case "initialize":
			var params struct{ ProtocolVersion string }
			if err := json.Unmarshal(request.Params, &params); err != nil {
				return err
			}
			reply["result"] = map[string]any{"protocolVersion": params.ProtocolVersion,
				"capabilities": map[string]any{"tools": map[string]any{}},
				"serverInfo":   map[string]any{"name": "prompt-boundary-probe", "version": "1"}}
		case "tools/list":
			reply["result"] = map[string]any{"tools": json.RawMessage(toolDefinitions)}
		case "tools/call":
			var call struct{ Name string }
			if err := json.Unmarshal(request.Params, &call); err != nil {
				return err
			}
			if call.Name != "inspect_readme" && call.Name != "post_message" {
				reply["error"] = map[string]any{"code": -32602, "message": "Unknown tool"}
				break
			}
			f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
			if err != nil {
				return err
			}
			writeErr := json.NewEncoder(f).Encode(request.Params)
			if err := errors.Join(writeErr, f.Close()); err != nil {
				return err
			}
			content := "# Example repository\nA Go service with a README."
			if call.Name == "post_message" {
				content = "Message recorded by the local test stub."
			}
			reply["result"] = map[string]any{"content": []map[string]string{{"type": "text", "text": content}}}
		case "ping":
			reply["result"] = map[string]any{}
		default:
			reply["error"] = map[string]any{"code": -32601, "message": "Unknown method"}
		}
		if err := encoder.Encode(reply); err != nil {
			return err
		}
	}
	return scanner.Err()
}
