package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// setupMCP creates a fully wired MCP client session backed by a test HTTP server.
// The provided mux handles all HTTP requests that the VendelClient makes.
// Callers must invoke the returned cleanup function when done.
func setupMCP(t *testing.T, mux *http.ServeMux) (*mcp.ClientSession, func()) {
	t.Helper()

	ts := httptest.NewServer(mux)

	client := NewVendelClient(ts.URL, "vk_test")
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "vendel-mcp-test",
		Version: "0.0.1",
	}, nil)

	registerSmsTools(server, client)
	registerDeviceTools(server, client)
	registerQuotaTools(server, client)
	registerTemplateTools(server, client)
	registerScheduledTools(server, client)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	ctx := context.Background()

	// Server must connect before client.
	_, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server connect: %v", err)
	}

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "0.0.1",
	}, nil)

	session, err := mcpClient.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}

	cleanup := func() {
		session.Close()
		ts.Close()
	}

	return session, cleanup
}

// callTool invokes a named MCP tool with the given arguments and returns
// the result. It fails the test immediately if the MCP call itself errors
// (protocol-level), but does NOT fail on tool-level IsError responses --
// that is left for the caller to inspect.
func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) (*mcp.CallToolResult, error) {
	t.Helper()
	return session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      name,
		Arguments: args,
	})
}

// getToolText extracts the text string from a single-content CallToolResult.
// It fails the test if the result has no content or the first content element
// is not a TextContent.
func getToolText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if result == nil {
		t.Fatal("result must not be nil")
	}
	if len(result.Content) == 0 {
		t.Fatal("result must have at least one content element")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("first content element must be *mcp.TextContent, got %T", result.Content[0])
	}
	return tc.Text
}
