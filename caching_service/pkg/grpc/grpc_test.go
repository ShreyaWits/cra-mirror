package grpc

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// --- Mock interfaces ---

type MockGRPCServer struct {
	mock.Mock
}

func (m *MockGRPCServer) Serve(lis net.Listener) error {
	args := m.Called(lis)
	return args.Error(0)
}
func (m *MockGRPCServer) GracefulStop() {
	m.Called()
}
func (m *MockGRPCServer) RegisterService(s *grpc.ServiceDesc, i interface{}) {
	m.Called()
}

type MockListener struct {
	mock.Mock
}

func (m *MockListener) Accept() (net.Conn, error) {
	args := m.Called()
	return args.Get(0).(net.Conn), args.Error(1)
}
func (m *MockListener) Close() error {
	args := m.Called()
	return args.Error(0)
}
func (m *MockListener) Addr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

// --- Tests ---

func TestNewGRPCServer_ConstructsInstance(t *testing.T) {
	server := new(MockGRPCServer)
	listener := new(MockListener)

	instance, err := NewGRPCServer(server, listener)
	assert.NoError(t, err)
	assert.NotNil(t, instance)
	assert.Equal(t, server, instance.server)
	assert.Equal(t, listener, instance.listener)
}

func TestGRPCServerInstance_Start_Success(t *testing.T) {
	server := new(MockGRPCServer)
	listener := new(MockListener)

	server.On("Serve", listener).Return(nil)
	listener.On("Addr").Return(&net.TCPAddr{Port: 12345})

	instance, _ := NewGRPCServer(server, listener)
	err := instance.Start()
	assert.NoError(t, err)
	server.AssertCalled(t, "Serve", listener)
	listener.AssertCalled(t, "Addr")
}

func TestGRPCServerInstance_Start_Error(t *testing.T) {
	server := new(MockGRPCServer)
	listener := new(MockListener)

	expectedErr := errors.New("serve failed")
	server.On("Serve", listener).Return(expectedErr)

	instance, _ := NewGRPCServer(server, listener)
	err := instance.Start()
	assert.Equal(t, expectedErr, err)
	server.AssertCalled(t, "Serve", listener)
}

func TestGRPCServerInstance_Stop(t *testing.T) {
	server := new(MockGRPCServer)
	listener := new(MockListener)

	server.On("GracefulStop").Return()

	instance, _ := NewGRPCServer(server, listener)
	instance.Stop()
	server.AssertCalled(t, "GracefulStop")
}

func TestGRPCServerInstance_GetServer(t *testing.T) {
	server := new(MockGRPCServer)
	listener := new(MockListener)

	instance, _ := NewGRPCServer(server, listener)
	assert.Equal(t, server, instance.GetServer())
}
