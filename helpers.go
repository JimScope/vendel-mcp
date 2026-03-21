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

// validateRecordID checks that a PocketBase record ID is safe for URL path interpolation.
func validateRecordID(id string) bool {
	if id == "" {
		return false
	}
	for _, c := range id {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// validateEnum checks if value is one of the allowed values.
func validateEnum(value string, allowed []string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	return false
}
