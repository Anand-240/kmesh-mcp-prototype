// MCP server prototype for kmesh-net/kmesh#1800. Wraps the /version and
// /debug/ready endpoints from Kmesh's status server as MCP tools.
// See README for how to run it.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var daemonAddr = flag.String("addr", "localhost:15200", "Kmesh daemon status-server address")

var httpClient = &http.Client{Timeout: 5 * time.Second}

type emptyInput struct{}

type versionOutput struct {
	Raw string `json:"raw" jsonschema:"raw JSON response from the Kmesh /version endpoint"`
}

func getVersion(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, versionOutput, error) {
	body, err := fetch(ctx, "/version")
	if err != nil {
		return nil, versionOutput{}, err
	}
	return nil, versionOutput{Raw: body}, nil
}

type healthOutput struct {
	Ready bool   `json:"ready" jsonschema:"whether the Kmesh daemon reported itself ready"`
	Raw   string `json:"raw" jsonschema:"raw response body from /debug/ready"`
}

func getDaemonHealth(ctx context.Context, req *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, healthOutput, error) {
	body, err := fetch(ctx, "/debug/ready")
	if err != nil {
		return nil, healthOutput{}, err
	}
	return nil, healthOutput{Ready: body == "OK", Raw: body}, nil
}

func fetch(ctx context.Context, path string) (string, error) {
	url := fmt.Sprintf("http://%s%s", *daemonAddr, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling kmesh daemon at %s: %w", url, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("kmesh daemon returned %s: %s", resp.Status, body)
	}
	return string(body), nil
}

func main() {
	flag.Parse()

	server := mcp.NewServer(&mcp.Implementation{Name: "kmesh-mcp-prototype", Version: "v0.0.1"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_version",
		Description: "Get the Kmesh daemon version info from its /version endpoint",
	}, getVersion)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_daemon_health",
		Description: "Check whether the Kmesh daemon is ready, via /debug/ready",
	}, getDaemonHealth)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
