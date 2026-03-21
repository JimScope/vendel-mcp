package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	baseURL := os.Getenv("VENDEL_URL")
	apiKey := os.Getenv("VENDEL_API_KEY")

	if baseURL == "" || apiKey == "" {
		fmt.Fprintln(os.Stderr, "VENDEL_URL and VENDEL_API_KEY environment variables are required")
		os.Exit(1)
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

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
