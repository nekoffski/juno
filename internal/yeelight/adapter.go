package yeelight

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/device"
)

const (
	defaultSsdpAddr = "239.255.255.250:1982"
	ssdpSearchMsg   = "M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1982\r\nMAN: \"ssdp:discover\"\r\nST: wifi_bulb\r\n\r\n"
	defaultPort     = 55443
)

type Adapter struct {
	ssdpAddr    string
	lanAgentURL string
}

func NewAdapter(ssdpAddr, lanAgentURL string) *Adapter {
	return &Adapter{ssdpAddr: ssdpAddr, lanAgentURL: lanAgentURL}
}

type lanDiscoverRequest struct {
	Addr       string `json:"addr"`
	Message    string `json:"message"`
	TimeoutSec int    `json:"timeout_sec"`
}

type lanDiscoveryResult struct {
	IP          string `json:"ip"`
	RawResponse string `json:"raw_response"`
}

type lanDiscoverResponse struct {
	Devices []lanDiscoveryResult `json:"devices"`
}

func parseResponse(ip, rawResponse string) (device.DeviceAddr, bool) {
	if !strings.Contains(strings.ToLower(rawResponse), "yeelight") {
		return device.DeviceAddr{}, false
	}

	port := defaultPort
	for _, line := range strings.Split(rawResponse, "\r\n") {
		if strings.HasPrefix(strings.ToLower(line), "location:") {
			parts := strings.Split(strings.TrimSpace(line[9:]), ":")
			if len(parts) == 3 {
				if p, err := strconv.Atoi(parts[2]); err == nil {
					port = p
				}
			}
		}
	}

	return device.DeviceAddr{Ip: ip, Port: port}, true
}

func readResponses(conn net.PacketConn) ([]device.DeviceAddr, error) {
	var devices []device.DeviceAddr
	buf := make([]byte, 4096)

	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			break
		}

		udpAddr, ok := addr.(*net.UDPAddr)
		if !ok {
			continue
		}

		if d, ok := parseResponse(udpAddr.IP.String(), string(buf[:n])); ok {
			devices = append(devices, d)
		}
	}

	return devices, nil
}

func (a *Adapter) discoverViaLanAgent(ctx context.Context) ([]device.DeviceAddr, error) {
	reqBody := lanDiscoverRequest{
		Addr:       a.ssdpAddr,
		Message:    ssdpSearchMsg,
		TimeoutSec: 3,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.lanAgentURL+"/discover", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lan-agent request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lan-agent returned status %d", resp.StatusCode)
	}

	var lanResp lanDiscoverResponse
	if err := json.NewDecoder(resp.Body).Decode(&lanResp); err != nil {
		return nil, fmt.Errorf("failed to decode lan-agent response: %w", err)
	}

	var devices []device.DeviceAddr
	for _, r := range lanResp.Devices {
		if d, ok := parseResponse(r.IP, r.RawResponse); ok {
			devices = append(devices, d)
		}
	}
	return devices, nil
}

func (a *Adapter) Discover(ctx context.Context) ([]device.DeviceAddr, error) {
	if a.lanAgentURL != "" {
		return a.discoverViaLanAgent(ctx)
	}

	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to open UDP socket: %w", err)
	}
	defer conn.Close()

	dest, err := net.ResolveUDPAddr("udp4", a.ssdpAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SSDP address: %w", err)
	}

	if _, err = conn.WriteTo([]byte(ssdpSearchMsg), dest); err != nil {
		return nil, fmt.Errorf("failed to send discovery message: %w", err)
	}

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	return readResponses(conn)
}

func (a *Adapter) CreateDevice(ctx context.Context, id int, addr device.DeviceAddr, name string, publisher *bus.Publisher) (device.Device, error) {
	return createDevice(ctx, id, addr, name, publisher, a.lanAgentURL)
}

func (a *Adapter) Name() device.DeviceVendor {
	return device.DeviceVendorYeelight
}
