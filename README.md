# kmesh-mcp-prototype

Small proof of concept for [kmesh-net/kmesh#1800](https://github.com/kmesh-net/kmesh/issues/1800), the LFX proposal to build an MCP server for Kmesh.

It wraps two read only endpoints that already exist on Kmesh's status server (`pkg/status/status_server.go`, port 15200) as MCP tools:

| Tool | Wraps | Existing code |
|---|---|---|
| `get_version` | `GET /version` | `pkg/status/status_server.go`, `version` handler |
| `get_daemon_health` | `GET /debug/ready` | `pkg/status/status_server.go`, `readyProbe` handler |

Kept it to two tools on purpose. The point was to check the approach from the issue actually works end to end (thin Go wrapper around an existing Kmesh HTTP endpoint, official Go MCP SDK, stdio transport) before writing all 10.

## Run

```sh
go run . -addr localhost:15200
```

Point an MCP client (Claude Desktop, mcp-inspector, Cursor) at the binary over stdio. `-addr` should point at a real Kmesh daemon's status server, for example after `kubectl port-forward` to a kmesh-daemon pod. Defaults to `localhost:15200`.

## Test

```sh
go test ./... -v
```

No cluster needed for this. The test spins up a fake HTTP server standing in for the Kmesh status server, then drives both tools through a real in-memory MCP client/server pair (`mcp.NewInMemoryTransports`) so the whole call path gets exercised.
