package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

type getDeviceArgs struct {
	ID int `json:"id" jsonschema:"The device ID to retrieve"`
}

type executeCommandArgs struct {
	ID     int            `json:"id"               jsonschema:"The device ID"`
	Action string         `json:"action"           jsonschema:"Action to execute (on, off, toggle, brightness, rgb, ct)"`
	Params map[string]any `json:"params,omitempty" jsonschema:"Optional action parameters"`
}

func RegisterTools(srv *sdkmcp.Server, client JunoClient) {
	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "get_devices",
		Description: "List all known devices",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, _ struct{}) (*sdkmcp.CallToolResult, any, error) {
		return handleGetDevices(ctx, client)
	})

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "get_device",
		Description: "Get a single device by its ID",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, args getDeviceArgs) (*sdkmcp.CallToolResult, any, error) {
		return handleGetDevice(ctx, client, args)
	})

	sdkmcp.AddTool(srv, &sdkmcp.Tool{
		Name:        "execute_command",
		Description: "Execute a command on a device",
	}, func(ctx context.Context, req *sdkmcp.CallToolRequest, args executeCommandArgs) (*sdkmcp.CallToolResult, any, error) {
		return handleExecuteCommand(ctx, client, args)
	})
}

func handleGetDevices(ctx context.Context, client JunoClient) (*sdkmcp.CallToolResult, any, error) {
	devices, err := client.GetDevices(ctx)
	if err != nil {
		return toolError(fmt.Sprintf("failed to get devices: %v", err)), nil, nil
	}
	data, err := json.Marshal(devices)
	if err != nil {
		return toolError("failed to marshal devices"), nil, nil
	}
	return textResult(string(data)), nil, nil
}

func handleGetDevice(ctx context.Context, client JunoClient, args getDeviceArgs) (*sdkmcp.CallToolResult, any, error) {
	device, err := client.GetDevice(ctx, args.ID)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return toolError(fmt.Sprintf("device %d not found", args.ID)), nil, nil
		}
		return toolError(fmt.Sprintf("failed to get device: %v", err)), nil, nil
	}
	data, err := json.Marshal(device)
	if err != nil {
		return toolError("failed to marshal device"), nil, nil
	}
	return textResult(string(data)), nil, nil
}

func handleExecuteCommand(ctx context.Context, client JunoClient, args executeCommandArgs) (*sdkmcp.CallToolResult, any, error) {
	if err := client.PerformAction(ctx, args.ID, args.Action, args.Params); err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			return toolError(fmt.Sprintf("device %d not found", args.ID)), nil, nil
		}
		return toolError(fmt.Sprintf("failed to execute command: %v", err)), nil, nil
	}
	return textResult("ok"), nil, nil
}

func textResult(text string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: text}},
	}
}

func toolError(msg string) *sdkmcp.CallToolResult {
	return &sdkmcp.CallToolResult{
		IsError: true,
		Content: []sdkmcp.Content{&sdkmcp.TextContent{Text: msg}},
	}
}
