package lan

import (
	"fmt"
	"net"
	"time"
)

type DiscoveryResult struct {
	IP          string `json:"ip"`
	RawResponse string `json:"raw_response"`
}

func discoverDevices(addr string, message string, timeoutSec int) ([]DiscoveryResult, error) {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to open UDP socket: %w", err)
	}
	defer conn.Close()

	dest, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target address: %w", err)
	}

	if _, err = conn.WriteTo([]byte(message), dest); err != nil {
		return nil, fmt.Errorf("failed to send discovery message: %w", err)
	}

	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	if err := conn.SetReadDeadline(deadline); err != nil {
		return nil, fmt.Errorf("failed to set read deadline: %w", err)
	}

	var results []DiscoveryResult
	buf := make([]byte, 4096)

	for {
		n, srcAddr, err := conn.ReadFrom(buf)
		if err != nil {
			// deadline exceeded or connection closed — stop reading
			break
		}

		udpAddr, ok := srcAddr.(*net.UDPAddr)
		if !ok {
			continue
		}

		results = append(results, DiscoveryResult{
			IP:          udpAddr.IP.String(),
			RawResponse: string(buf[:n]),
		})
	}

	return results, nil
}
