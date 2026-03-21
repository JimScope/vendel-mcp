package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ScheduleSmsInput struct {
	Name           string   `json:"name" jsonschema:"Name/label for this scheduled SMS"`
	Recipients     []string `json:"recipients" jsonschema:"Phone numbers in E.164 format"`
	Body           string   `json:"body" jsonschema:"SMS message text (max 1600 characters)"`
	ScheduleType   string   `json:"schedule_type" jsonschema:"one_time for a single send or recurring for cron-based"`
	ScheduledAt    string   `json:"scheduled_at,omitempty" jsonschema:"When to send (ISO 8601 datetime). Required for one_time"`
	CronExpression string   `json:"cron_expression,omitempty" jsonschema:"Cron expression for recurring schedules (e.g. 0 9 * * MON)"`
	Timezone       string   `json:"timezone,omitempty" jsonschema:"IANA timezone (e.g. America/New_York). Defaults to UTC"`
	DeviceID       string   `json:"device_id,omitempty" jsonschema:"Specific device ID to send from"`
}

type ListScheduledInput struct {
	Status string `json:"status,omitempty" jsonschema:"Filter by schedule status (active paused completed)"`
}

func registerScheduledTools(server *mcp.Server, client *VendelClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "schedule_sms",
		Description: "Schedule an SMS for future delivery (one-time or recurring)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ScheduleSmsInput) (*mcp.CallToolResult, any, error) {
		if input.Name == "" {
			return errorResult("schedule SMS", fmt.Errorf("name must not be empty")), nil, nil
		}
		if len(input.Recipients) == 0 {
			return errorResult("schedule SMS", fmt.Errorf("recipients must not be empty")), nil, nil
		}
		if input.Body == "" {
			return errorResult("schedule SMS", fmt.Errorf("body must not be empty")), nil, nil
		}
		if !validateEnum(input.ScheduleType, []string{"one_time", "recurring"}) {
			return errorResult("schedule SMS", fmt.Errorf("invalid schedule_type %q: must be one of one_time, recurring", input.ScheduleType)), nil, nil
		}

		tz := input.Timezone
		if tz == "" {
			tz = "UTC"
		}

		data := map[string]any{
			"name":          input.Name,
			"recipients":    input.Recipients,
			"body":          input.Body,
			"schedule_type": input.ScheduleType,
			"timezone":      tz,
			"status":        "active",
		}
		if input.ScheduledAt != "" {
			data["scheduled_at"] = input.ScheduledAt
		}
		if input.CronExpression != "" {
			data["cron_expression"] = input.CronExpression
		}
		if input.DeviceID != "" {
			data["device_id"] = input.DeviceID
		}

		result, err := createRecord[ScheduledSms](ctx, client, "scheduled_sms", data)
		if err != nil {
			return errorResult("schedule SMS", err), nil, nil
		}

		lines := []string{
			"SMS scheduled",
			fmt.Sprintf("  ID: %s", result.ID),
			fmt.Sprintf("  Name: %s", result.Name),
			fmt.Sprintf("  Type: %s", result.ScheduleType),
			fmt.Sprintf("  Recipients: %d", len(input.Recipients)),
			fmt.Sprintf("  Timezone: %s", result.Timezone),
		}
		if result.ScheduledAt != "" {
			lines = append(lines, fmt.Sprintf("  Scheduled for: %s", result.ScheduledAt))
		}
		if result.CronExpression != "" {
			lines = append(lines, fmt.Sprintf("  Cron: %s", result.CronExpression))
		}
		if result.NextRunAt != "" {
			lines = append(lines, fmt.Sprintf("  Next run: %s", result.NextRunAt))
		}

		return textResult(strings.Join(lines, "\n")), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_scheduled",
		Description: "List scheduled SMS messages",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListScheduledInput) (*mcp.CallToolResult, any, error) {
		filter := ""
		if input.Status != "" {
			if !validateEnum(input.Status, []string{"active", "paused", "completed"}) {
				return errorResult("list scheduled SMS", fmt.Errorf("invalid status %q: must be one of active, paused, completed", input.Status)), nil, nil
			}
			filter = fmt.Sprintf(`status="%s"`, input.Status)
		}

		result, err := listRecords[ScheduledSms](ctx, client, "scheduled_sms", &ListParams{
			Filter: filter,
			Sort:   "-created",
		})
		if err != nil {
			return errorResult("list scheduled SMS", err), nil, nil
		}

		if len(result.Items) == 0 {
			return textResult("No scheduled SMS found."), nil, nil
		}

		var blocks []string
		for _, s := range result.Items {
			lines := []string{
				fmt.Sprintf("- %s [%s] | ID: %s", s.Name, strings.ToUpper(s.Status), s.ID),
				fmt.Sprintf("  Type: %s | Recipients: %d", s.ScheduleType, len(s.Recipients)),
			}
			if s.NextRunAt != "" {
				lines = append(lines, fmt.Sprintf("  Next run: %s", s.NextRunAt))
			}
			if s.LastRunAt != "" {
				lines = append(lines, fmt.Sprintf("  Last run: %s", s.LastRunAt))
			}
			if s.CronExpression != "" {
				lines = append(lines, fmt.Sprintf("  Cron: %s (%s)", s.CronExpression, s.Timezone))
			}
			blocks = append(blocks, strings.Join(lines, "\n"))
		}

		header := fmt.Sprintf("Scheduled SMS (%d):\n\n", result.TotalItems)
		return textResult(header + strings.Join(blocks, "\n\n")), nil, nil
	})
}
