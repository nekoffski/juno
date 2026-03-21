package yeelight

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/nekoffski/juno/internal/device"
)

type request struct {
	ID     int    `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type response struct {
	ID     int               `json:"id"`
	Result []any             `json:"result"`
	Method string            `json:"method"`
	Params map[string]string `json:"params"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type pendingRequest struct {
	ch chan response
}

type notification struct {
	Params map[string]string
}

type notificationCallback = func(n notification)

type client struct {
	addr           device.DeviceAddr
	conn           net.Conn
	done           chan struct{}
	mu             sync.Mutex
	pending        map[int]*pendingRequest
	nextID         atomic.Int32
	onNotification notificationCallback
}

func newClient(ctx context.Context, addr device.DeviceAddr, onNotification notificationCallback) (*client, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(addr.Ip, strconv.Itoa(addr.Port)))
	if err != nil {
		return nil, err
	}
	c := &client{
		addr:           addr,
		conn:           conn,
		done:           make(chan struct{}),
		pending:        make(map[int]*pendingRequest),
		onNotification: onNotification,
	}
	go c.readLoop(ctx)
	return c, nil
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
			log.Printf("Could not unmarshal message: %v", err)
			continue
		}
		log.Printf("Received message: %s", scanner.Bytes())

		if msg.Method == "props" {
			if c.onNotification != nil {
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
			log.Printf("Received response for unknown request ID %d", msg.ID)
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

	log.Printf("Sending message: %s", data)
	if _, err := c.conn.Write(data); err != nil {
		return nil, err
	}

	return pr, nil
}

func (c *client) close() error {
	return c.conn.Close()
}
