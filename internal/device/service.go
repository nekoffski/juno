package device

import (
	"context"
	"log"

	"github.com/nekoffski/juno/internal/bus"
	"github.com/nekoffski/juno/internal/models"
)

type DeviceService struct {
	sender *bus.Sender
}

func NewDeviceService() *DeviceService {
	return &DeviceService{}
}

func (s *DeviceService) Name() string {
	return "device-service"
}

func (s *DeviceService) onMessage(msg bus.Message) {
	switch req := msg.Payload.(type) {
	case models.HeartbeatRequest:
		log.Printf("Got heartbeat request")
		msg.Reply(bus.Response{Payload: models.HeartbeatResponse{Healthy: true, Magic: req.Magic}})
	}
}

func (s *DeviceService) Init(ctx context.Context, mb *bus.MessageBus) error {
	s.sender = mb.NewSender()
	mb.RegisterReceiver(ctx, s.Name(), func(msg bus.Message) {
		s.onMessage(msg)
	})

	return nil
}

func (s *DeviceService) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
