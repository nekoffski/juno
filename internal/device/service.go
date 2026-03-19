package device

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
)

type DeviceService struct {
	sender   *bus.Sender
	adapters map[DeviceVendor]VendorAdapter
	pool     *pgxpool.Pool
	devices  map[DeviceAddr]Device
}

func NewDeviceService(pool *pgxpool.Pool, adapters map[DeviceVendor]VendorAdapter) *DeviceService {
	return &DeviceService{
		pool:     pool,
		adapters: adapters,
		devices:  make(map[DeviceAddr]Device),
	}
}

func (s *DeviceService) Name() string {
	return "device-service"
}

// TODO: this needs context

func (s *DeviceService) onMessage(msg bus.Message) {
	switch req := msg.Payload.(type) {
	case core.HeartbeatRequest:
		log.Printf("Got heartbeat request")
		msg.Reply(bus.Response{Payload: core.HeartbeatResponse{Healthy: true, Magic: req.Magic}})

	case DiscoverDevicesRequest:
		log.Printf("Got discover devices request")
		if err := s.discover(context.Background()); err != nil {
			log.Printf("Failed to discover devices: %v", err)
		}
	}
}

func (s *DeviceService) Init(ctx context.Context, mb *bus.MessageBus) error {
	s.sender = mb.NewSender()
	mb.RegisterReceiver(ctx, s.Name(), func(msg bus.Message) {
		s.onMessage(msg)
	})

	s.loadDevices(ctx)
	return nil
}

func (s *DeviceService) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (s *DeviceService) loadDevices(ctx context.Context) error {
	return fetchDevices(ctx, s.pool, func(id int, addr DeviceAddr, vendor DeviceVendor, name string) {
		if _, exists := s.devices[addr]; exists {
			return
		}

		dev, err := s.adapters[vendor].CreateDevice(id, addr, name)
		if err != nil {
			log.Printf("Failed to create device with adapter: %v", err)
			return
		}

		s.devices[addr] = dev
	})
}

func (s *DeviceService) addDevice(ctx context.Context, addr DeviceAddr, vendor DeviceVendor) {
	log.Printf("Adding device at %s:%d", addr.Ip, addr.Port)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}
	defer tx.Rollback(ctx)

	id, name, err := insertDevice(ctx, tx, addr, vendor)
	if err != nil {
		log.Printf("Failed to insert device: %v", err)
		return
	}

	dev, err := s.adapters[vendor].CreateDevice(id, addr, name)
	if err != nil {
		log.Printf("Failed to create device with adapter: %v", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	s.devices[addr] = dev
}

func (s *DeviceService) discover(ctx context.Context) error {
	for vendor, adapter := range s.adapters {
		log.Printf("Discovering devices for vendor %s", vendor)

		ctx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		devices, err := adapter.Discover(ctx)
		if err != nil {
			log.Printf("Failed to discover devices for vendor %s: %v", vendor, err)
			continue
		}

		for _, device := range devices {
			if _, exists := s.devices[device]; exists {
				continue
			}
			s.addDevice(ctx, device, vendor)
		}

	}

	return nil
}
