package yeelight

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/nekoffski/juno/internal/device"
)

const (
	ssdpAddr      = "239.255.255.250:1982"
	ssdpSearchMsg = "M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1982\r\nMAN: \"ssdp:discover\"\r\nST: wifi_bulb\r\n\r\n"
	defaultPort   = 55443
)

type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func readResponses(conn net.PacketConn) ([]device.DeviceAddr, error) {
	var devices []device.DeviceAddr
	buf := make([]byte, 4096)

	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			break
		}

		response := string(buf[:n])
		if !strings.Contains(strings.ToLower(response), "yeelight") {
			continue
		}

		udpAddr, ok := addr.(*net.UDPAddr)
		if !ok {
			continue
		}

		port := defaultPort
		for _, line := range strings.Split(response, "\r\n") {
			if strings.HasPrefix(strings.ToLower(line), "location:") {
				parts := strings.Split(strings.TrimSpace(line[9:]), ":")
				if len(parts) == 3 {
					if p, err := strconv.Atoi(parts[2]); err == nil {
						port = p
					}
				}
			}
		}

		devices = append(devices, device.DeviceAddr{
			Ip:   udpAddr.IP.String(),
			Port: port,
		})
	}

	return devices, nil
}

func (a *Adapter) Discover(ctx context.Context) ([]device.DeviceAddr, error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to open UDP socket: %w", err)
	}
	defer conn.Close()

	dest, err := net.ResolveUDPAddr("udp4", ssdpAddr)
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

func (a *Adapter) CreateDevice(ctx context.Context, id int, addr device.DeviceAddr, name string) (device.Device, error) {
	return createDevice(ctx, id, addr, name)
}

func (a *Adapter) Name() device.DeviceVendor {
	return device.DeviceVendorYeelight
}
