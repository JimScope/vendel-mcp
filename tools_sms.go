package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SendSmsInput struct {
	Recipients []string `json:"recipients" jsonschema:"Phone numbers in E.164 format (e.g. +1234567890)"`
	Body       string   `json:"body" jsonschema:"SMS message text (max 1600 characters)"`
	DeviceID   string   `json:"device_id,omitempty" jsonschema:"Specific device ID to send from (auto-selects if omitted)"`
}

type ListMessagesInput struct {
	Type    string `json:"type,omitempty" jsonschema:"Filter by message type (outgoing or incoming)"`
	Status  string `json:"status,omitempty" jsonschema:"Filter by message status (pending assigned sending sent delivered failed received)"`
	Page    int    `json:"page,omitempty" jsonschema:"Page number (default: 1)"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Items per page (default: 20)"`
}

type GetMessageInput struct {
	MessageID string `json:"message_id" jsonschema:"The message record ID"`
}

func registerSmsTools(server *mcp.Server, client *VendelClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_sms",
		Description: "Send an SMS message to one or more phone numbers via Vendel gateway",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SendSmsInput) (*mcp.CallToolResult, any, error) {
		result, err := client.SendSms(&SendSmsRequest{
			Recipients: input.Recipients,
			Body:       input.Body,
			DeviceID:   input.DeviceID,
		})
		if err != nil {
			return errorResult("send SMS", err), nil, nil
		}

		lines := []string{
			"SMS queued successfully",
			fmt.Sprintf("  Recipients: %d", result.RecipientsCount),
			fmt.Sprintf("  Status: %s", result.Status),
			fmt.Sprintf("  Message IDs: %s", strings.Join(result.MessageIDs, ", ")),
		}
		if result.BatchID != "" {
			lines = append(lines, fmt.Sprintf("  Batch ID: %s", result.BatchID))
		}

		return textResult(strings.Join(lines, "\n")), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_messages",
		Description: "List SMS messages sent or received through Vendel",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListMessagesInput) (*mcp.CallToolResult, any, error) {
		var filters []string
		if input.Type != "" {
			filters = append(filters, fmt.Sprintf(`message_type="%s"`, input.Type))
		}
		if input.Status != "" {
			filters = append(filters, fmt.Sprintf(`status="%s"`, input.Status))
		}

		page := input.Page
		if page == 0 {
			page = 1
		}
		perPage := input.PerPage
		if perPage == 0 {
			perPage = 20
		}

		result, err := listRecords[SmsMessage](client, "sms_messages", &ListParams{
			Filter:  strings.Join(filters, " && "),
			Page:    page,
			PerPage: perPage,
			Sort:    "-created",
		})
		if err != nil {
			return errorResult("list messages", err), nil, nil
		}

		if len(result.Items) == 0 {
			return textResult("No messages found."), nil, nil
		}

		var lines []string
		lines = append(lines, fmt.Sprintf("Messages (page %d/%d, %d total):", result.Page, result.TotalPages, result.TotalItems))
		for _, m := range result.Items {
			direction := fmt.Sprintf("To: %s", m.To)
			if m.MessageType == "incoming" {
				direction = fmt.Sprintf("From: %s", m.FromNumber)
			}
			lines = append(lines, fmt.Sprintf("[%s] %s | %s | %s", strings.ToUpper(m.Status), direction, m.Body, m.Created))
		}

		return textResult(strings.Join(lines, "\n")), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_message",
		Description: "Get detailed information about a specific SMS message",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetMessageInput) (*mcp.CallToolResult, any, error) {
		m, err := getRecord[SmsMessage](client, "sms_messages", input.MessageID)
		if err != nil {
			return errorResult("get message", err), nil, nil
		}

		direction := fmt.Sprintf("To: %s", m.To)
		if m.MessageType == "incoming" {
			direction = fmt.Sprintf("From: %s", m.FromNumber)
		}

		lines := []string{
			fmt.Sprintf("Message: %s", m.ID),
			fmt.Sprintf("  Type: %s", m.MessageType),
			fmt.Sprintf("  Status: %s", m.Status),
			fmt.Sprintf("  %s", direction),
			fmt.Sprintf("  Body: %s", m.Body),
			fmt.Sprintf("  Created: %s", m.Created),
		}
		if m.SentAt != "" {
			lines = append(lines, fmt.Sprintf("  Sent at: %s", m.SentAt))
		}
		if m.DeliveredAt != "" {
			lines = append(lines, fmt.Sprintf("  Delivered at: %s", m.DeliveredAt))
		}
		if m.ErrorMessage != "" {
			lines = append(lines, fmt.Sprintf("  Error: %s", m.ErrorMessage))
		}
		if m.BatchID != "" {
			lines = append(lines, fmt.Sprintf("  Batch ID: %s", m.BatchID))
		}
		if m.Device != "" {
			lines = append(lines, fmt.Sprintf("  Device: %s", m.Device))
		}

		return textResult(strings.Join(lines, "\n")), nil, nil
	})
}
