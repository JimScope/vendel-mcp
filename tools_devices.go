package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListDevicesInput struct {
	DeviceType string `json:"device_type,omitempty" jsonschema:"Filter by device type (android or modem)"`
}

func registerDeviceTools(server *mcp.Server, client *VendelClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_devices",
		Description: "List registered SMS gateway devices and their status",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListDevicesInput) (*mcp.CallToolResult, any, error) {
		filter := ""
		if input.DeviceType != "" {
			if !validateEnum(input.DeviceType, []string{"android", "modem"}) {
				return errorResult("list devices", fmt.Errorf("invalid device_type %q: must be one of android, modem", input.DeviceType)), nil, nil
			}
			filter = fmt.Sprintf(`device_type="%s"`, input.DeviceType)
		}

		result, err := listRecords[SmsDevice](ctx, client, "sms_devices", &ListParams{
			Filter: filter,
			Sort:   "-created",
		})
		if err != nil {
			return errorResult("list devices", err), nil, nil
		}

		if len(result.Items) == 0 {
			return textResult("No devices registered."), nil, nil
		}

		var lines []string
		lines = append(lines, fmt.Sprintf("Devices (%d):", result.TotalItems))
		for _, d := range result.Items {
			lines = append(lines, fmt.Sprintf("- %s (%s) | %s | ID: %s", d.Name, d.DeviceType, d.PhoneNumber, d.ID))
		}

		return textResult(strings.Join(lines, "\n")), nil, nil
	})
}
