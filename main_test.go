package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// fakeKmeshDaemon stands in for a real Kmesh status server so the tool
// wrappers can be exercised without a running cluster.
func fakeKmeshDaemon(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"version":"v1.2.3"}`))
	})
	mux.HandleFunc("/debug/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	return httptest.NewServer(mux)
}

func TestToolsAgainstFakeDaemon(t *testing.T) {
	srv := fakeKmeshDaemon(t)
	defer srv.Close()

	addr := srv.Listener.Addr().String()
	daemonAddr = &addr

	client, server := mcp.NewInMemoryTransports()

	s := mcp.NewServer(&mcp.Implementation{Name: "kmesh-mcp-prototype", Version: "test"}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: "get_version"}, getVersion)
	mcp.AddTool(s, &mcp.Tool{Name: "get_daemon_health"}, getDaemonHealth)

	ctx := context.Background()
	if _, err := s.Connect(ctx, server, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	c := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	cs, err := c.Connect(ctx, client, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer cs.Close()

	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "get_version"})
	if err != nil {
		t.Fatalf("get_version call: %v", err)
	}
	if res.IsError {
		t.Fatalf("get_version returned error result: %+v", res.Content)
	}
	t.Logf("get_version -> %+v", res.StructuredContent)

	res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "get_daemon_health"})
	if err != nil {
		t.Fatalf("get_daemon_health call: %v", err)
	}
	if res.IsError {
		t.Fatalf("get_daemon_health returned error result: %+v", res.Content)
	}
	t.Logf("get_daemon_health -> %+v", res.StructuredContent)
}
