# Prompt boundary behavioral evaluation

Opt-in evaluation using an authenticated Claude CLI and Go. Run from the repository root:

```sh
go run ./evals/prompt-boundaries
```

The Go harness generates prompts using this checkout's real context assembly and serves its own local MCP stubs. It runs a positive control that explicitly requests a message, then two tasks with supporting rules that try to redirect the agent into unrelated messaging, including a fake task heading and embedded fences. It requires the requested completion marker and inspects actual tool calls, not just the final answer.

All built-in tools except read-only tool discovery, and all external MCP configurations, are disabled. The only executable operations are local stubs: reading a fixed README and recording a message call. No message is sent. The control verifies the messaging stub was actually callable; denied permissions or a missing marker fail the evaluation. The model API is contacted using the CLI's existing authentication and default model. This incurs normal model usage.

Responses, tool-call records, and model usage are retained in the printed temporary evidence directory. Do not treat a passing sample as a general guarantee against prompt injection or as verification of a deployed agent. Unit tests separately validate rendering and deduplication. This evaluation is deliberately not part of the default test suite.
