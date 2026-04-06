package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type GetMessageStatusInput struct {
	MessageID string `json:"message_id" jsonschema:"The message record ID"`
}

type GetBatchStatusInput struct {
	BatchID string `json:"batch_id" jsonschema:"The batch ID returned from send endpoints"`
}

func registerStatusTools(server *mcp.Server, client *VendelClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_message_status",
		Description: "Get the delivery status of a single SMS message",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetMessageStatusInput) (*mcp.CallToolResult, any, error) {
		if !validateRecordID(input.MessageID) {
			return errorResult("get message status", fmt.Errorf("invalid message_id %q", input.MessageID)), nil, nil
		}

		m, err := client.GetMessageStatus(ctx, input.MessageID)
		if err != nil {
			return errorResult("get message status", err), nil, nil
		}

		lines := []string{
			fmt.Sprintf("Message: %s", m.ID),
			fmt.Sprintf("  Status: %s", m.Status),
			fmt.Sprintf("  Recipient: %s", m.Recipient),
		}
		if m.BatchID != "" {
			lines = append(lines, fmt.Sprintf("  Batch ID: %s", m.BatchID))
		}
		if m.DeviceID != "" {
			lines = append(lines, fmt.Sprintf("  Device: %s", m.DeviceID))
		}
		if m.ErrorMessage != "" {
			lines = append(lines, fmt.Sprintf("  Error: %s", m.ErrorMessage))
		}
		lines = append(lines, fmt.Sprintf("  Created: %s", m.Created))
		lines = append(lines, fmt.Sprintf("  Updated: %s", m.Updated))

		return textResult(strings.Join(lines, "\n")), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_batch_status",
		Description: "Get the delivery status of all messages in a batch",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetBatchStatusInput) (*mcp.CallToolResult, any, error) {
		if input.BatchID == "" {
			return errorResult("get batch status", fmt.Errorf("batch_id is required")), nil, nil
		}

		b, err := client.GetBatchStatus(ctx, input.BatchID)
		if err != nil {
			return errorResult("get batch status", err), nil, nil
		}

		var lines []string
		lines = append(lines, fmt.Sprintf("Batch: %s (%d messages)", b.BatchID, b.Total))

		// Status summary
		var counts []string
		for status, count := range b.StatusCounts {
			counts = append(counts, fmt.Sprintf("%s: %d", status, count))
		}
		if len(counts) > 0 {
			lines = append(lines, fmt.Sprintf("  Summary: %s", strings.Join(counts, ", ")))
		}

		// Individual messages
		lines = append(lines, "  Messages:")
		for _, m := range b.Messages {
			line := fmt.Sprintf("    [%s] %s → %s", strings.ToUpper(m.Status), m.ID, m.Recipient)
			if m.ErrorMessage != "" {
				line += fmt.Sprintf(" (error: %s)", m.ErrorMessage)
			}
			lines = append(lines, line)
		}

		return textResult(strings.Join(lines, "\n")), nil, nil
	})
}
