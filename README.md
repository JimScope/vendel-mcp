# vendel-mcp

MCP server for [Vendel SMS Gateway](https://vendel.cc) — send and receive SMS from Claude.

Exposes Vendel's SMS gateway as MCP tools, allowing Claude Desktop, Claude Code, and any MCP client to send messages, check quotas, manage templates, and schedule SMS.

## Install

```bash
go install github.com/JimScope/vendel-mcp@latest
```

## Setup

### Claude Code

```bash
claude mcp add vendel -e VENDEL_URL=https://your-instance.com -e VENDEL_API_KEY=vk_your_key -- vendel-mcp
```

### Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "vendel": {
      "command": "vendel-mcp",
      "env": {
        "VENDEL_URL": "https://api.vendel.cc",
        "VENDEL_API_KEY": "vk_your_api_key"
      }
    }
  }
}
```

## Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `VENDEL_URL` | Yes | Your Vendel instance URL (e.g. `https://api.vendel.cc`) |
| `VENDEL_API_KEY` | Yes | Integration API key (starts with `vk_`) |

## Tools

| Tool | Description |
|------|-------------|
| `send_sms` | Send an SMS to one or more phone numbers |
| `list_messages` | List sent/received messages with filters |
| `get_message` | Get details of a specific message |
| `check_quota` | Check plan limits and current usage |
| `list_devices` | List registered gateway devices |
| `list_templates` | List available SMS templates |
| `send_template` | Send an SMS using a saved template |
| `schedule_sms` | Schedule an SMS for future delivery |
| `list_scheduled` | List scheduled SMS messages |

## Resources

| URI | Description |
|-----|-------------|
| `vendel://quota` | Current quota and plan limits as JSON |

## Development

```bash
go build -o vendel-mcp .
VENDEL_URL=http://localhost:8090 VENDEL_API_KEY=vk_test ./vendel-mcp
```

## License

MIT
