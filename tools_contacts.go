package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ListContactsInput struct {
	Search  string `json:"search,omitempty" jsonschema:"Search contacts by name or phone number"`
	GroupID string `json:"group_id,omitempty" jsonschema:"Filter by contact group ID"`
	Page    int    `json:"page,omitempty" jsonschema:"Page number (default: 1)"`
	PerPage int    `json:"per_page,omitempty" jsonschema:"Items per page (default: 50)"`
}

type ListContactGroupsInput struct {
	Page    int `json:"page,omitempty" jsonschema:"Page number (default: 1)"`
	PerPage int `json:"per_page,omitempty" jsonschema:"Items per page (default: 50)"`
}

func registerContactTools(server *mcp.Server, client *VendelClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_contacts",
		Description: "List contacts in the user's address book with optional search and group filter",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListContactsInput) (*mcp.CallToolResult, any, error) {
		page := input.Page
		if page == 0 {
			page = 1
		}
		perPage := input.PerPage
		if perPage == 0 {
			perPage = 50
		}

		result, err := client.ListContacts(ctx, page, perPage, input.Search, input.GroupID)
		if err != nil {
			return errorResult("list contacts", err), nil, nil
		}

		if len(result.Items) == 0 {
			return textResult("No contacts found."), nil, nil
		}

		var lines []string
		lines = append(lines, fmt.Sprintf("Contacts (page %d/%d, %d total):", result.Page, result.TotalPages, result.TotalItems))
		for _, c := range result.Items {
			line := fmt.Sprintf("  [%s] %s — %s", c.ID, c.Name, c.PhoneNumber)
			if len(c.Groups) > 0 {
				line += fmt.Sprintf(" (groups: %s)", strings.Join(c.Groups, ", "))
			}
			if c.Notes != "" {
				line += fmt.Sprintf(" — %s", c.Notes)
			}
			lines = append(lines, line)
		}

		return textResult(strings.Join(lines, "\n")), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_contact_groups",
		Description: "List contact groups used to organize contacts",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ListContactGroupsInput) (*mcp.CallToolResult, any, error) {
		page := input.Page
		if page == 0 {
			page = 1
		}
		perPage := input.PerPage
		if perPage == 0 {
			perPage = 50
		}

		result, err := client.ListContactGroups(ctx, page, perPage)
		if err != nil {
			return errorResult("list contact groups", err), nil, nil
		}

		if len(result.Items) == 0 {
			return textResult("No contact groups found."), nil, nil
		}

		var lines []string
		lines = append(lines, fmt.Sprintf("Contact groups (page %d/%d, %d total):", result.Page, result.TotalPages, result.TotalItems))
		for _, g := range result.Items {
			lines = append(lines, fmt.Sprintf("  [%s] %s", g.ID, g.Name))
		}

		return textResult(strings.Join(lines, "\n")), nil, nil
	})
}
