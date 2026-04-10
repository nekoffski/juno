package yeelight

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/nekoffski/juno/internal/device"
	"github.com/rs/zerolog/log"
)

type request struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type response struct {
	ID     int            `json:"id"`
	Result []any          `json:"result"`
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type pendingRequest struct {
	ch chan response
}

type notification struct {
	Params map[string]any
}

type notificationCallback = func(n notification)

type client struct {
	addr                    device.DeviceAddr
	conn                    net.Conn
	done                    chan struct{}
	mu                      sync.Mutex
	pending                 map[int]*pendingRequest
	nextID                  atomic.Int32
	notificationCallbackSet atomic.Bool
	onNotification          notificationCallback
}

func dialViaProxy(proxyURL string, deviceAddr string) (net.Conn, error) {
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid lan-agent URL: %w", err)
	}

	proxyHost := u.Host
	conn, err := net.Dial("tcp", proxyHost)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to lan-agent: %w", err)
	}

	connectReq := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Host: deviceAddr},
		Host:   deviceAddr,
		Header: make(http.Header),
	}
	if err := connectReq.Write(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to write CONNECT request: %w", err)
	}

	br := bufio.NewReaderSize(conn, 1)
	resp, err := http.ReadResponse(br, connectReq)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to read CONNECT response: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		conn.Close()
		return nil, fmt.Errorf("CONNECT to %s via lan-agent returned %d", deviceAddr, resp.StatusCode)
	}

	return conn, nil
}

func newClient(ctx context.Context, addr device.DeviceAddr, lanAgentURL string) (*client, error) {
	deviceAddr := net.JoinHostPort(addr.Ip, strconv.Itoa(addr.Port))

	var conn net.Conn
	var err error

	if lanAgentURL != "" {
		conn, err = dialViaProxy(lanAgentURL, deviceAddr)
	} else {
		conn, err = net.Dial("tcp", deviceAddr)
	}
	if err != nil {
		return nil, err
	}
	c := &client{
		addr:    addr,
		conn:    conn,
		done:    make(chan struct{}),
		pending: make(map[int]*pendingRequest),
	}
	go c.readLoop(ctx)
	return c, nil
}

func (c *client) setNotificationCallback(onNotification notificationCallback) {
	c.onNotification = onNotification
	c.notificationCallbackSet.Store(true)
}

func (c *client) readLoop(ctx context.Context) {
	defer func() {
		c.conn.Close()
		close(c.done)
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		c.conn.Close()
	}()

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		var msg response
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			log.Error().Err(err).Msg("could not unmarshal message")
			continue
		}
		log.Debug().RawJSON("message", scanner.Bytes()).Msg("received message")

		if msg.Method == "props" {
			if c.notificationCallbackSet.Load() && c.onNotification != nil {
				c.onNotification(notification{Params: msg.Params})
			}
			continue
		}

		c.mu.Lock()
		pr, ok := c.pending[msg.ID]
		if ok {
			delete(c.pending, msg.ID)
		}
		c.mu.Unlock()

		if !ok {
			log.Warn().Int("id", msg.ID).Msg("received response for unknown request ID")
			continue
		}

		pr.ch <- msg
	}
}

func waitForResponse(ctx context.Context, pr *pendingRequest) (response, error) {
	select {
	case r := <-pr.ch:
		return r, nil
	case <-ctx.Done():
		return response{}, ctx.Err()
	}
}

func (c *client) readProperties(ctx context.Context, props []string) (map[string]string, error) {
	params := make([]any, len(props))
	for i, p := range props {
		params[i] = p
	}
	resp, err := c.sendRequest(ctx, "get_prop", params)
	if err != nil {
		return nil, err
	}

	r, err := waitForResponse(ctx, resp)
	if err != nil {
		return nil, err
	}

	if r.Error != nil {
		return nil, fmt.Errorf("device error: %d - %s", r.Error.Code, r.Error.Message)
	}

	result := make(map[string]string)
	for i, prop := range props {
		if i < len(r.Result) {
			if val, ok := r.Result[i].(string); ok {
				result[prop] = val
			}
		}
	}
	return result, nil
}

func (c *client) sendRequest(ctx context.Context, method string, params []any) (*pendingRequest, error) {
	id := int(c.nextID.Add(1))
	req := request{
		ID:     id,
		Method: method,
		Params: params,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data = append(data, '\r', '\n')

	pr := &pendingRequest{ch: make(chan response, 1)}
	c.mu.Lock()
	c.pending[id] = pr
	c.mu.Unlock()

	select {
	case <-c.done:
		return nil, context.Canceled
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	log.Debug().Str("data", string(data)).Msg("sending message")
	if _, err := c.conn.Write(data); err != nil {
		return nil, err
	}

	return pr, nil
}

func (c *client) close() error {
	return c.conn.Close()
}
