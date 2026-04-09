package device

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/core"
)

type DeviceService struct {
	sender     *bus.Sender
	publisher  *bus.Publisher
	adapters   map[DeviceVendor]VendorAdapter
	repo       DeviceRepository
	devices    map[DeviceAddr]Device
	devicesMtx sync.RWMutex
}

func NewDeviceService(repo DeviceRepository, adapters map[DeviceVendor]VendorAdapter) *DeviceService {
	return &DeviceService{
		repo:     repo,
		adapters: adapters,
		devices:  make(map[DeviceAddr]Device),
	}
}

func (s *DeviceService) Name() string {
	return "device-service"
}

func (s *DeviceService) onMessage(msg bus.Message) {
	switch req := msg.Payload.(type) {
	case core.HeartbeatRequest:
		log.Printf("Got heartbeat request")
		s.onHeartbeatRequest(&msg, req)

	case DiscoverDevicesRequest:
		log.Printf("Got discover devices request")
		s.onDiscoverDevicesRequest()

	case GetDevicesRequest:
		log.Printf("Got get devices request")
		s.onGetDevicesRequest(&msg)

	case GetDeviceByIdRequest:
		log.Printf("Got get device by id request for id %d", req.Id)
		s.onGetDeviceByIdRequest(&msg, req)

	case GetDevicePropertiesRequest:
		log.Printf("Got get device properties request for id %d and properties %v", req.Id, req.Properties)
		s.onGetDevicePropertiesRequest(&msg, req)

	case PerformDeviceActionRequest:
		log.Printf("Got perform device action request for id %d and action %s with params %v", req.Id, req.Action, req.Params)
		s.onPerformDeviceActionRequest(&msg, req)

	case DeleteDeviceRequest:
		log.Printf("Got delete device request for id %d", req.Id)
		s.onDeleteDeviceRequest(&msg, req)
	}
}

func (s *DeviceService) onDeleteDeviceRequest(msg *bus.Message, req DeleteDeviceRequest) {
	s.devicesMtx.Lock()
	var addrToDelete *DeviceAddr
	var devToClose Device
	for addr, dev := range s.devices {
		if dev != nil && dev.Model().Id == req.Id {
			addr := addr
			addrToDelete = &addr
			devToClose = dev
			break
		}
	}
	if addrToDelete == nil {
		s.devicesMtx.Unlock()
		msg.Reply(bus.Response{Err: core.ErrDeviceNotFound})
		return
	}
	delete(s.devices, *addrToDelete)
	s.devicesMtx.Unlock()

	if err := devToClose.Close(); err != nil {
		log.Printf("Error closing device %d: %v", req.Id, err)
	}
	if err := s.repo.DeleteDevice(context.Background(), req.Id); err != nil {
		log.Printf("Error deleting device %d from database: %v", req.Id, err)
	}
	msg.Reply(bus.Response{Payload: AckResponse{}})
}

func (s *DeviceService) onPerformDeviceActionRequest(msg *bus.Message, req PerformDeviceActionRequest) {
	dev := s.findDevice(req.Id)
	if dev == nil {
		msg.Reply(bus.Response{Err: core.ErrDeviceNotFound})
		return
	}

	if !dev.IsCapable(req.Action) {
		msg.Reply(bus.Response{Err: core.ErrDeviceNotCapable})
		return
	}

	params, err := parseActionParams(req.Action, req.Params)
	if err != nil {
		msg.Reply(bus.Response{Err: core.ErrInvalidArguments})
		return
	}

	err = dev.EnqueueAction(Action{Method: req.Action, Params: params})
	if err != nil {
		msg.Reply(bus.Response{Err: err})
		return
	}
	msg.Reply(bus.Response{Payload: AckResponse{}})
}

func (s *DeviceService) onGetDevicePropertiesRequest(msg *bus.Message, req GetDevicePropertiesRequest) {
	d := s.findDevice(req.Id)
	if d == nil {
		msg.Reply(bus.Response{Err: core.ErrDeviceNotFound})
		return
	}

	props := d.Model().Properties
	res := GetDevicePropertiesResponse{
		Properties: make(map[string]any),
	}

	for _, prop := range req.Properties {
		res.Properties[prop] = nil
		if val, exists := props[prop]; exists {
			res.Properties[prop] = val
		}
	}

	msg.Reply(bus.Response{Payload: res})
}

func (s *DeviceService) onHeartbeatRequest(msg *bus.Message, req core.HeartbeatRequest) {
	msg.Reply(bus.Response{Payload: core.HeartbeatResponse{Healthy: true, Magic: req.Magic}})
}

func (s *DeviceService) onDiscoverDevicesRequest() {
	go func() {
		if err := s.discover(context.Background()); err != nil {
			log.Printf("Failed to discover devices: %v", err)
		}
	}()
}

func (s *DeviceService) onGetDeviceByIdRequest(msg *bus.Message, req GetDeviceByIdRequest) {
	s.devicesMtx.RLock()
	defer s.devicesMtx.RUnlock()

	for _, dev := range s.devices {
		if dev == nil {
			continue
		}
		if dev.Model().Id == req.Id {
			model := dev.Model()
			msg.Reply(bus.Response{Payload: GetDeviceByIdResponse{Device: &model}})
			return
		}
	}
	msg.Reply(bus.Response{Err: core.ErrDeviceNotFound})
}

func (s *DeviceService) onGetDevicesRequest(msg *bus.Message) {
	s.devicesMtx.RLock()
	defer s.devicesMtx.RUnlock()

	res := GetDevicesResponse{
		Devices: make([]*DeviceModel, 0, len(s.devices)),
	}
	for _, dev := range s.devices {
		if dev == nil {
			continue
		}
		model := dev.Model()
		res.Devices = append(res.Devices, &model)
	}
	msg.Reply(bus.Response{Payload: res})
}

func (s *DeviceService) Init(ctx context.Context, mb *bus.MessageBus) error {
	s.sender = mb.NewSender()
	s.publisher = mb.NewPublisher()

	if err := bus.RegisterTopic(mb, "device.events"); err != nil {
		return err
	}

	if err := mb.RegisterReceiver(ctx, s.Name(), func(msg bus.Message) {
		s.onMessage(msg)
	}); err != nil {
		return err
	}

	go s.loadDevices(ctx)
	return nil
}

func (s *DeviceService) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (s *DeviceService) findDevice(id int) Device {
	s.devicesMtx.RLock()
	defer s.devicesMtx.RUnlock()

	for _, dev := range s.devices {
		if dev == nil {
			continue
		}
		if dev.Model().Id == id {
			return dev
		}
	}
	return nil
}

func (s *DeviceService) exists(addr DeviceAddr) bool {
	s.devicesMtx.RLock()
	defer s.devicesMtx.RUnlock()
	_, exists := s.devices[addr]
	return exists
}

func (s *DeviceService) loadDevices(ctx context.Context) {
	err := s.repo.FetchDevices(ctx, func(id int, addr DeviceAddr, vendor DeviceVendor, name string) {
		go func() {
			if s.exists(addr) {
				return
			}

			dev, err := s.adapters[vendor].CreateDevice(ctx, id, addr, name, s.publisher)
			if err != nil {
				log.Printf("Failed to create device with adapter: %v", err)
				return
			}

			s.devicesMtx.Lock()
			s.devices[addr] = dev
			s.devicesMtx.Unlock()
		}()
	})
	if err != nil {
		log.Printf("Failed to load devices from database: %v", err)
	}
}

func (s *DeviceService) addDevice(ctx context.Context, addr DeviceAddr, vendor DeviceVendor) {
	log.Printf("Adding device at %s:%d", addr.Ip, addr.Port)

	id, name, err := s.repo.InsertDevice(ctx, addr, vendor)
	if err != nil {
		log.Printf("Failed to insert device: %v", err)
		return
	}

	dev, err := s.adapters[vendor].CreateDevice(ctx, id, addr, name, s.publisher)
	if err != nil {
		log.Printf("Failed to create device with adapter: %v", err)
		return
	}

	s.devicesMtx.Lock()
	s.devices[addr] = dev
	s.devicesMtx.Unlock()
}

func (s *DeviceService) discover(ctx context.Context) error {
	for vendor, adapter := range s.adapters {
		log.Printf("Discovering devices for vendor %s", vendor)

		discoverCtx, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		devices, err := adapter.Discover(discoverCtx)
		if err != nil {
			log.Printf("Failed to discover devices for vendor %s: %v", vendor, err)
			continue
		}

		for _, device := range devices {
			if s.exists(device) {
				continue
			}
			s.devicesMtx.Lock()
			s.devices[device] = nil
			s.devicesMtx.Unlock()

			go func(device DeviceAddr, vendor DeviceVendor) {
				s.addDevice(ctx, device, vendor)
			}(device, vendor)
		}

	}

	return nil
}
