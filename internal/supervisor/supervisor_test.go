package supervisor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errInit = errors.New("init error")
var errRun = errors.New("run error")

type mockService struct {
	name    string
	initErr error
	runErr  error
	initFn  func(*MessageBus)
	runFn   func(context.Context)
}

func (m *mockService) Name() string { return m.name }

func (m *mockService) Init(mb *MessageBus) error {
	if m.initFn != nil {
		m.initFn(mb)
	}
	return m.initErr
}

func (m *mockService) Run(ctx context.Context) error {
	if m.runFn != nil {
		m.runFn(ctx)
	}
	return m.runErr
}

func newMock(name string) *mockService {
	return &mockService{name: name}
}

func TestNewSupervisor(t *testing.T) {
	a, b := newMock("a"), newMock("b")
	s := NewSupervisor(a, b)
	require.NotNil(t, s)
	assert.Len(t, s.services, 2)
	assert.NotNil(t, s.messageBus)
}

func TestInitServices_OK(t *testing.T) {
	initialized := make([]string, 0)
	mkSvc := func(name string) *mockService {
		return &mockService{name: name, initFn: func(*MessageBus) {
			initialized = append(initialized, name)
		}}
	}
	s := NewSupervisor(mkSvc("a"), mkSvc("b"), mkSvc("c"))
	require.NoError(t, s.initServices())
	assert.Equal(t, []string{"a", "b", "c"}, initialized)
}

func TestInitServices_FailsOnFirstError(t *testing.T) {
	second := false
	s := NewSupervisor(
		&mockService{name: "a", initErr: errInit},
		&mockService{name: "b", initFn: func(*MessageBus) { second = true }},
	)
	err := s.initServices()
	require.ErrorIs(t, err, errInit)
	assert.False(t, second, "second service should not be initialized after first fails")
}

func TestInitServices_MessageBusPassedToServices(t *testing.T) {
	var receivedBus *MessageBus
	s := NewSupervisor(&mockService{
		name: "a",
		initFn: func(mb *MessageBus) {
			receivedBus = mb
		},
	})
	require.NoError(t, s.initServices())
	assert.Same(t, s.messageBus, receivedBus)
}

func TestStartServices_OK(t *testing.T) {
	s := NewSupervisor(newMock("a"), newMock("b"))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.NoError(t, s.startServices(ctx))
}

func TestStartServices_SingleError(t *testing.T) {
	s := NewSupervisor(&mockService{name: "a", runErr: errRun})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := s.startServices(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, errRun)
}

func TestStartServices_MultipleErrors(t *testing.T) {
	errA := errors.New("error from a")
	errB := errors.New("error from b")
	s := NewSupervisor(
		&mockService{name: "a", runErr: errA},
		&mockService{name: "b", runErr: errB},
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := s.startServices(ctx)
	require.Error(t, err)
	assert.ErrorIs(t, err, errA)
	assert.ErrorIs(t, err, errB)
}

func TestStartServices_ContextCancellationPropagated(t *testing.T) {
	cancelled := make(chan struct{})
	s := NewSupervisor(&mockService{
		name: "a",
		runFn: func(ctx context.Context) {
			<-ctx.Done()
			close(cancelled)
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.startServices(ctx) }()

	cancel()

	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("service did not receive context cancellation in time")
	}

	require.NoError(t, <-done)
}

func TestStartServices_AllServicesRunConcurrently(t *testing.T) {
	ready := make(chan struct{}, 3)
	block := make(chan struct{})

	mkSvc := func(name string) *mockService {
		return &mockService{name: name, runFn: func(ctx context.Context) {
			ready <- struct{}{}
			select {
			case <-block:
			case <-ctx.Done():
			}
		}}
	}

	s := NewSupervisor(mkSvc("a"), mkSvc("b"), mkSvc("c"))
	ctx, cancel := context.WithCancel(context.Background())

	go s.startServices(ctx) //nolint:errcheck

	for i := 0; i < 3; i++ {
		select {
		case <-ready:
		case <-time.After(time.Second):
			cancel()
			t.Fatalf("service %d did not start in time", i)
		}
	}

	cancel()
}
