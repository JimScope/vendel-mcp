package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createServer builds the MCP server with all tools and resources registered.
func createServer(baseURL, apiKey string) (*mcp.Server, error) {
	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("VENDEL_URL and VENDEL_API_KEY environment variables are required")
	}

	client := NewVendelClient(baseURL, apiKey)
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "vendel-mcp",
		Version: "1.0.0",
	}, nil)

	registerSmsTools(server, client)
	registerDeviceTools(server, client)
	registerQuotaTools(server, client)
	registerTemplateTools(server, client)
	registerScheduledTools(server, client)
	registerContactTools(server, client)
	registerStatusTools(server, client)

	return server, nil
}

func main() {
	server, err := createServer(os.Getenv("VENDEL_URL"), os.Getenv("VENDEL_API_KEY"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
