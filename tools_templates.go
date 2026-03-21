package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListTemplatesInput struct{}

type SendTemplateInput struct {
	TemplateID string   `json:"template_id" jsonschema:"The template record ID"`
	Recipients []string `json:"recipients" jsonschema:"Phone numbers in E.164 format (e.g. +1234567890)"`
	DeviceID   string   `json:"device_id,omitempty" jsonschema:"Specific device ID to send from (auto-selects if omitted)"`
}

func registerTemplateTools(server *mcp.Server, client *VendelClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_templates",
		Description: "List available SMS templates",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListTemplatesInput) (*mcp.CallToolResult, any, error) {
		result, err := listRecords[SmsTemplate](ctx, client, "sms_templates", &ListParams{
			Sort: "-created",
		})
		if err != nil {
			return errorResult("list templates", err), nil, nil
		}

		if len(result.Items) == 0 {
			return textResult("No templates found."), nil, nil
		}

		var lines []string
		lines = append(lines, fmt.Sprintf("Templates (%d):", result.TotalItems))
		for _, t := range result.Items {
			lines = append(lines, fmt.Sprintf("- %s [ID: %s]\n  Body: %s", t.Name, t.ID, t.Body))
		}

		return textResult(strings.Join(lines, "\n\n")), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "send_template",
		Description: "Send an SMS using an existing template",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SendTemplateInput) (*mcp.CallToolResult, any, error) {
		if !validateRecordID(input.TemplateID) {
			return errorResult("send template", fmt.Errorf("invalid template_id %q", input.TemplateID)), nil, nil
		}
		if len(input.Recipients) == 0 {
			return errorResult("send template", fmt.Errorf("recipients must not be empty")), nil, nil
		}

		template, err := getRecord[SmsTemplate](ctx, client, "sms_templates", input.TemplateID)
		if err != nil {
			return errorResult("fetch template", err), nil, nil
		}

		result, err := client.SendSms(ctx, &SendSmsRequest{
			Recipients: input.Recipients,
			Body:       template.Body,
			DeviceID:   input.DeviceID,
		})
		if err != nil {
			return errorResult("send template", err), nil, nil
		}

		lines := []string{
			fmt.Sprintf("Template \"%s\" sent", template.Name),
			fmt.Sprintf("  Recipients: %d", result.RecipientsCount),
			fmt.Sprintf("  Status: %s", result.Status),
			fmt.Sprintf("  Message IDs: %s", strings.Join(result.MessageIDs, ", ")),
		}
		if result.BatchID != "" {
			lines = append(lines, fmt.Sprintf("  Batch ID: %s", result.BatchID))
		}

		return textResult(strings.Join(lines, "\n")), nil, nil
	})
}
