package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockJunoClient is a test double for JunoClient.
type mockJunoClient struct {
	devices       []Device
	devicesByID   map[int]*Device
	actionErr     error
	getDevicesErr error
}

func (m *mockJunoClient) GetDevices(_ context.Context) ([]Device, error) {
	return m.devices, m.getDevicesErr
}

func (m *mockJunoClient) GetDevice(_ context.Context, id int) (*Device, error) {
	if m.devicesByID != nil {
		if d, ok := m.devicesByID[id]; ok {
			return d, nil
		}
	}
	return nil, ErrDeviceNotFound
}

func (m *mockJunoClient) PerformAction(_ context.Context, id int, _ string, _ map[string]any) error {
	if m.actionErr != nil {
		return m.actionErr
	}
	if m.devicesByID != nil {
		if _, ok := m.devicesByID[id]; !ok {
			return ErrDeviceNotFound
		}
	}
	return nil
}

// setupMCPSession creates an in-process MCP server+client session with the given mock client.
func setupMCPSession(t *testing.T, mock JunoClient) *sdkmcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	srv := sdkmcp.NewServer(&sdkmcp.Implementation{Name: "test-juno-mcp"}, nil)
	RegisterTools(srv, mock)

	t1, t2 := sdkmcp.NewInMemoryTransports()

	_, err := srv.Connect(ctx, t1, nil)
	require.NoError(t, err)

	c := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client"}, nil)
	cs, err := c.Connect(ctx, t2, nil)
	require.NoError(t, err)
	t.Cleanup(func() { cs.Close() })

	return cs
}

func TestRegisterTools_ListsThreeTools(t *testing.T) {
	cs := setupMCPSession(t, &mockJunoClient{})
	res, err := cs.ListTools(context.Background(), nil)
	require.NoError(t, err)
	names := make([]string, 0, len(res.Tools))
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	assert.ElementsMatch(t, []string{"get_devices", "get_device", "execute_command"}, names)
}

func TestHandlerGetDevices_ReturnsDeviceList(t *testing.T) {
	mock := &mockJunoClient{
		devices: []Device{
			{Id: 1, Name: "light-1", Capabilities: []string{"on", "off"}},
			{Id: 2, Name: "light-2"},
		},
	}
	cs := setupMCPSession(t, mock)
	res, err := cs.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: "get_devices"})
	require.NoError(t, err)
	require.False(t, res.IsError, "expected no tool error")
	text := res.Content[0].(*sdkmcp.TextContent).Text
	var devices []Device
	require.NoError(t, json.Unmarshal([]byte(text), &devices))
	assert.Len(t, devices, 2)
	assert.Equal(t, 1, devices[0].Id)
}

func TestHandlerGetDevices_EmptyList(t *testing.T) {
	cs := setupMCPSession(t, &mockJunoClient{devices: []Device{}})
	res, err := cs.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: "get_devices"})
	require.NoError(t, err)
	require.False(t, res.IsError)
	var devices []Device
	require.NoError(t, json.Unmarshal([]byte(res.Content[0].(*sdkmcp.TextContent).Text), &devices))
	assert.Empty(t, devices)
}

func TestHandlerGetDevices_ClientError(t *testing.T) {
	mock := &mockJunoClient{getDevicesErr: errors.New("connection refused")}
	cs := setupMCPSession(t, mock)
	res, err := cs.CallTool(context.Background(), &sdkmcp.CallToolParams{Name: "get_devices"})
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Contains(t, res.Content[0].(*sdkmcp.TextContent).Text, "failed to get devices")
}

func TestHandlerGetDevice_Found(t *testing.T) {
	mock := &mockJunoClient{
		devicesByID: map[int]*Device{
			7: {Id: 7, Name: "bulb", Capabilities: []string{"toggle"}},
		},
	}
	cs := setupMCPSession(t, mock)
	res, err := cs.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "get_device",
		Arguments: map[string]any{"id": 7},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)
	var device Device
	require.NoError(t, json.Unmarshal([]byte(res.Content[0].(*sdkmcp.TextContent).Text), &device))
	assert.Equal(t, 7, device.Id)
	assert.Equal(t, "bulb", device.Name)
}

func TestHandlerGetDevice_NotFound(t *testing.T) {
	cs := setupMCPSession(t, &mockJunoClient{devicesByID: map[int]*Device{}})
	res, err := cs.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "get_device",
		Arguments: map[string]any{"id": 999},
	})
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Contains(t, res.Content[0].(*sdkmcp.TextContent).Text, "not found")
}

func TestHandlerExecuteCommand_OK(t *testing.T) {
	mock := &mockJunoClient{
		devicesByID: map[int]*Device{
			3: {Id: 3, Name: "lamp"},
		},
	}
	cs := setupMCPSession(t, mock)
	res, err := cs.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "execute_command",
		Arguments: map[string]any{"id": 3, "action": "toggle"},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)
	assert.Equal(t, "ok", res.Content[0].(*sdkmcp.TextContent).Text)
}

func TestHandlerExecuteCommand_WithParams(t *testing.T) {
	mock := &mockJunoClient{
		devicesByID: map[int]*Device{
			3: {Id: 3, Name: "lamp"},
		},
	}
	cs := setupMCPSession(t, mock)
	res, err := cs.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name: "execute_command",
		Arguments: map[string]any{
			"id":     3,
			"action": "brightness",
			"params": map[string]any{"brightness": float64(75)},
		},
	})
	require.NoError(t, err)
	require.False(t, res.IsError)
	assert.Equal(t, "ok", res.Content[0].(*sdkmcp.TextContent).Text)
}

func TestHandlerExecuteCommand_DeviceNotFound(t *testing.T) {
	cs := setupMCPSession(t, &mockJunoClient{devicesByID: map[int]*Device{}})
	res, err := cs.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "execute_command",
		Arguments: map[string]any{"id": 404, "action": "toggle"},
	})
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Contains(t, res.Content[0].(*sdkmcp.TextContent).Text, "not found")
}

func TestHandlerExecuteCommand_ActionError(t *testing.T) {
	mock := &mockJunoClient{actionErr: errors.New("device busy")}
	cs := setupMCPSession(t, mock)
	res, err := cs.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "execute_command",
		Arguments: map[string]any{"id": 1, "action": "on"},
	})
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Contains(t, res.Content[0].(*sdkmcp.TextContent).Text, "failed to execute command")
}
