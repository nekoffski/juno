package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Device struct {
	Id           int                    `json:"id"`
	Name         string                 `json:"name"`
	Vendor       interface{}            `json:"vendor"`
	Status       interface{}            `json:"status"`
	Capabilities []string               `json:"capabilities"`
	Properties   map[string]interface{} `json:"properties"`
}

var ErrDeviceNotFound = errors.New("device not found")

type JunoClient interface {
	GetDevices(ctx context.Context) ([]Device, error)
	GetDevice(ctx context.Context, id int) (*Device, error)
	PerformAction(ctx context.Context, id int, action string, params map[string]any) error
}

type httpJunoClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string) JunoClient {
	return &httpJunoClient{baseURL: baseURL, client: &http.Client{}}
}

func (c *httpJunoClient) GetDevices(ctx context.Context) ([]Device, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/device", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var devices []Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, err
	}
	return devices, nil
}

func (c *httpJunoClient) GetDevice(ctx context.Context, id int) (*Device, error) {
	url := fmt.Sprintf("%s/device/id/%d", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrDeviceNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var device Device
	if err := json.NewDecoder(resp.Body).Decode(&device); err != nil {
		return nil, err
	}
	return &device, nil
}

type actionRequest struct {
	Params map[string]any `json:"params,omitempty"`
}

func (c *httpJunoClient) PerformAction(ctx context.Context, id int, action string, params map[string]any) error {
	body, err := json.Marshal(actionRequest{Params: params})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/device/id/%d/action/%s", c.baseURL, id, action)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrDeviceNotFound
	case http.StatusBadRequest:
		return fmt.Errorf("invalid action or parameters")
	default:
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}
