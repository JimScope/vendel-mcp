package main

import (
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// textResult creates a successful CallToolResult with text content.
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: text},
		},
	}
}

// errorResult creates an error CallToolResult with a formatted message.
func errorResult(action string, err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Failed to %s: %s", action, err)},
		},
		IsError: true,
	}
}
