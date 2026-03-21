package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type EmptyInput struct{}

func registerQuotaTools(server *mcp.Server, client *VendelClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "check_quota",
		Description: "Check your current SMS quota, plan limits, and usage",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input EmptyInput) (*mcp.CallToolResult, any, error) {
		q, err := client.GetQuota(ctx)
		if err != nil {
			return errorResult("check quota", err), nil, nil
		}

		remaining := q.MaxSmsPerMonth - q.SmsSentThisMonth
		lines := []string{
			fmt.Sprintf("Plan: %s", q.Plan),
			fmt.Sprintf("SMS: %d/%d used (%d remaining)", q.SmsSentThisMonth, q.MaxSmsPerMonth, remaining),
			fmt.Sprintf("Devices: %d/%d", q.DevicesRegistered, q.MaxDevices),
			fmt.Sprintf("Scheduled SMS: %d/%d", q.ScheduledSmsActive, q.MaxScheduledSms),
			fmt.Sprintf("Integrations: %d/%d", q.IntegrationsCreated, q.MaxIntegrations),
			fmt.Sprintf("Quota resets: %s", q.ResetDate),
		}

		return textResult(strings.Join(lines, "\n")), nil, nil
	})

	// Resource: vendel://quota
	server.AddResource(&mcp.Resource{
		URI:         "vendel://quota",
		Name:        "quota",
		Title:       "SMS Quota",
		Description: "Current SMS quota and plan limits",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		q, err := client.GetQuota(ctx)
		if err != nil {
			return nil, fmt.Errorf("fetch quota: %w", err)
		}

		data, err := json.MarshalIndent(q, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("marshal quota: %w", err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      "vendel://quota",
				MIMEType: "application/json",
				Text:     string(data),
			}},
		}, nil
	})
}
